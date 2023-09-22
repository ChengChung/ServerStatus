package promql

import (
	"fmt"
	"time"

	"github.com/VictoriaMetrics/metricsql"
	"github.com/chengchung/ServerStatus/common/concurrency"
	"github.com/chengchung/ServerStatus/common/utils"
	"github.com/chengchung/ServerStatus/datasource/metricsql/helper"
	"github.com/chengchung/ServerStatus/proto"
	"github.com/sirupsen/logrus"
)

var query_procedures []pipeline_funct = []pipeline_funct{
	(*QueryClient).refresh_latest_info,
	(*QueryClient).update_hosts,
	(*QueryClient).prepare_host_ctx,
	(*QueryClient).update_hosts_metrics,
	(*QueryClient).write_info,
}

func (c *QueryClient) refresh_latest_info(rawctx interface{}) {
	ctx := rawctx.(*QueryProcedureContext)
	c.lock.Lock()
	defer c.lock.Unlock()

	if c.candidate_prop != nil {
		c.current_prop = c.candidate_prop
		c.candidate_prop = nil
	}
	last_hosts := make([]string, 0)
	for k := range c.report {
		last_hosts = append(last_hosts, k)
	}
	ctx.last_hosts = last_hosts

	logrus.Info("finish refresh latest client info")
}

func (c *QueryClient) update_hosts(rawctx interface{}) {
	ctx := rawctx.(*QueryProcedureContext)

	var up_hosts []string
	var down_hosts []string
	var err error

	cur_env := c.current_prop

	if cur_env.enable_dynamic_import {
		//	动态模式，所有hosts由查询得到，同时对于下线节点在报告中予以保留
		up_hosts, err = c.get_potential_hosts(helper.AlterOption{
			Filters: cur_env.global_restrictions,
		})
		down_hosts = get_down_hosts(ctx.last_hosts, up_hosts)
	} else {
		//	静态模式，所有hosts由配置给出
		var opt helper.AlterOption
		opt.Filters = append(opt.Filters, cur_env.global_restrictions...)
		opt.Filters = append(opt.Filters, c.get_id_labels_filter(cur_env.known_hosts))
		up_hosts, err = c.get_potential_hosts(opt)
		down_hosts = get_down_hosts(cur_env.known_hosts, up_hosts)
	}
	if err != nil {
		logrus.Error(err)
	} else {
		logrus.Infof("get up hosts %v", up_hosts)
	}
	if len(down_hosts) != 0 {
		logrus.Infof("get down hosts %v", down_hosts)
	}

	ctx.up_hosts = up_hosts
	ctx.down_hosts = down_hosts

	logrus.Info("finish fetching latest up hosts")
}

func (c *QueryClient) prepare_host_ctx(rawctx interface{}) {
	ctx := rawctx.(*QueryProcedureContext)

	network_matchers := make(map[string][]metricsql.LabelFilter)
	host_billing_settings := make(map[string]time.Time)
	for host, value := range c.current_prop.overwrites {
		if len(value.NetworkDevices) == 0 {
			continue
		}
		network_matchers[host] = []metricsql.LabelFilter{{
			Label:    default_network_device_label,
			Value:    utils.RegQuoteOr(value.NetworkDevices),
			IsRegexp: true,
		}}

		if len(value.BillingDate) > 0 {
			t, err := time.Parse(time.RFC3339, value.BillingDate)
			if err != nil {
				logrus.Errorf("invalid billing date %s for host %s", value.BillingDate, host)
			} else {
				host_billing_settings[host] = t
			}
		}
	}

	ctx.network_metric_overwrites = c.current_prop.network_overwrites
	ctx.host_billing_settings = host_billing_settings

	ctx.network_matchers = network_matchers
}

func (c *QueryClient) update_hosts_metrics(rawctx interface{}) {
	ctx := rawctx.(*QueryProcedureContext)

	if len(ctx.up_hosts) == 0 {
		return
	}

	task_group := concurrency.NewBatchTask(len(query_tasks))
	for k, f := range query_tasks {
		task := task_group.DispatchTask()
		go c.run_task(k, task, f, rawctx)
	}

	results := task_group.WaitForFinish()

	for _, host := range ctx.up_hosts {
		ctx.results[host] = make(map[string]interface{})
	}

	for _, rawres := range results {
		result := rawres.(concurrency.TaskResult)
		propertyType := result.ID
		rawvalue := result.Result
		if rawvalue == nil {
			logrus.Errorf("get nil result for property %s", propertyType)
			continue
		}
		if propertyType == "property" {
			m := rawvalue.(map[string]proto.ServerProperty)
			for host, value := range m {
				ctx.results[host][propertyType] = value
			}
		} else {
			m := rawvalue.(map[string]float64)
			for host, value := range m {
				ctx.results[host][propertyType] = value
			}
		}
	}
}

func (c *QueryClient) run_task(taskname string, task *concurrency.Task, funct query_funct, rawctx interface{}) {
	result, err := funct(c, rawctx)
	if err != nil {
		logrus.Error(err)
		task.AnswerWithID(taskname, nil)
	} else {
		task.AnswerWithID(taskname, result)
	}

	logrus.Infof("finish executing task %s", taskname)
}

func (c *QueryClient) write_info(rawctx interface{}) {
	c.lock.Lock()
	defer c.lock.Unlock()

	ctx := rawctx.(*QueryProcedureContext)
	c.last_updated = time.Now().Unix()

	//	静态配置条件下，节点条目以配置内容为准，因此考虑到节点删除的情况，需要先进行清空
	if !c.current_prop.enable_dynamic_import {
		c.report = make(map[string]proto.ServerStatus)
	}

	for _, host := range ctx.down_hosts {
		hostname := host
		if len(c.current_prop.overwrites[host].DisplayName) > 0 {
			hostname = c.current_prop.overwrites[host].DisplayName
		}
		c.report[host] = proto.ServerStatus{
			Name:    hostname,
			Host:    host,
			Online4: false,
			Online6: false,
		}
	}

	for host, m := range ctx.results {
		hostname := host
		if len(c.current_prop.overwrites[host].DisplayName) > 0 {
			hostname = c.current_prop.overwrites[host].DisplayName
		}
		status := proto.ServerStatus{
			Name:    hostname,
			Host:    host,
			Online4: true,
			Online6: false,
		}
		for k, v := range m {
			switch v := v.(type) {
			case proto.ServerProperty:
				status.ServerProperty = v
			case float64:
				switch k {
				case "uptime":
					secs := int64(v)
					days := int64(0)
					if secs > int64(ONE_DAY_SECONDS) {
						days = secs / int64(ONE_DAY_SECONDS)
						secs = secs % int64(ONE_DAY_SECONDS)
					}
					days_str := ""
					if days > 0 {
						days_str = fmt.Sprintf("%dd", days)
					}
					duration := time.Duration(secs * int64(time.Second))
					status.Uptime = days_str + duration.String()
				case "load":
					status.SetTypeValue(k, v)
				default:
					status.SetTypeValue(k, int64(v))
				}
			default:
				logrus.Errorf("invalid result value for host %s key %s value type %T value %s", host, k, v, v)
			}
		}
		c.report[host] = status
	}
}
