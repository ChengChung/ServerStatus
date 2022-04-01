package prometheus_data

import (
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/chengchung/ServerStatus/common/concurrency"
	"github.com/chengchung/ServerStatus/common/configuration"
	"github.com/chengchung/ServerStatus/common/utils"
	"github.com/chengchung/ServerStatus/datasource"
	"github.com/chengchung/ServerStatus/datasource/prometheus_ds"
	"github.com/chengchung/ServerStatus/proto"
	"github.com/sirupsen/logrus"
)

var default_up_time_query_str = `sum(time() - %s) by (##ID_LABEL##)`
var default_up_time_source_metric = `node_boot_time_seconds`

var default_load1_source_metric = "node_load1"

var default_cpu_cnt_query_str = `sum(count(%s) by (cpu, ##ID_LABEL##)) by (##ID_LABEL##)`
var default_cpu_cnt_source_metric = `node_cpu_seconds_total`
var default_cpu_cnt_matcher utils.PrometheusMatcher = utils.NewPrometheusMatcher("mode", "=", "system")

var default_memory_total_source_metric = `node_memory_MemTotal_bytes`
var default_memory_avail_source_metric = `node_memory_MemAvailable_bytes`
var default_memory_total_query_str = `%s / 1024`
var default_memory_used_query_str = `(%s - %s) / 1024`

var defautl_swap_total_source_metric = `node_memory_SwapTotal_bytes`
var default_swap_free_source_metric = `node_memory_SwapFree_bytes`
var default_swap_total_query_str = `%s / 1024`
var default_swap_used_query_str = `(%s - %s) / 1024`

var default_hdd_total_source_metric = `node_filesystem_size_bytes`
var default_hdd_free_source_metric = `node_filesystem_free_bytes`
var default_hdd_total_query_str = `%s / 1024 / 1024`
var default_hdd_used_query_str = `(%s - %s) / 1024 / 1024`
var default_hdd_matcher utils.PrometheusMatcher = utils.NewPrometheusMatcher("fstype", "=~", "ext4|xfs|ubifs")

var default_network_rx_source_metric = `node_network_receive_bytes_total`
var default_network_tx_source_metric = `node_network_transmit_bytes_total`

var default_single_network_rate_query_str = `irate(%s[2m])`
var default_singe_network_sum_query_str = `increase(%s[##TIME_RANGE##])`
var default_network_rate_query_str = `sum(%s) by (##ID_LABEL##)`
var default_network_sum_query_str = `sum(%s) by (##ID_LABEL##)`

var default_network_device_label = `device`

var ONE_DAY_SECONDS = 24 * 60 * 60

type NodesImportMode string

const (
	NODES_IMPORT_AUTO   NodesImportMode = "AUTO"   //	节点将从查询中得到，动态增加
	NODES_IMPORT_STATIC NodesImportMode = "STATIC" //	节点将从文件中读取
)

var (
	query_procedures []pipeline_funct = []pipeline_funct{
		(*QueryClient).refresh_latest_info,
		(*QueryClient).update_hosts,
		(*QueryClient).prepare_host_ctx,
		(*QueryClient).update_hosts_metrics,
		(*QueryClient).write_info,
	}
	query_tasks map[string]query_funct = map[string]query_funct{
		"property":     (*QueryClient).get_host_properties,
		"uptime":       (*QueryClient).get_host_uptime,
		"load":         (*QueryClient).get_host_load1,
		"cpu":          (*QueryClient).get_host_cpu_cnt,
		"memory_total": (*QueryClient).get_host_memory_total,
		"memory_used":  (*QueryClient).get_host_memory_used,
		"swap_total":   (*QueryClient).get_host_swap_total,
		"swap_used":    (*QueryClient).get_host_swap_used,
		"hdd_total":    (*QueryClient).get_host_hdd_total,
		"hdd_used":     (*QueryClient).get_host_hdd_used,
		"network_in":   (*QueryClient).get_host_network_rx_total,
		"network_out":  (*QueryClient).get_host_network_tx_total,
		"network_rx":   (*QueryClient).get_host_network_rx_rate,
		"network_tx":   (*QueryClient).get_host_network_tx_rate,
	}
	network_default_matchers []utils.PrometheusMatcher = []utils.PrometheusMatcher{
		utils.NewPrometheusMatcher(default_network_device_label, "!~", `tap.*|veth.*|br.*|docker.*|virbr*|lo*`),
	}
)

type Property struct {
	known_hosts           []string
	enable_dynamic_import bool

	global_restrictions []utils.PrometheusMatcher
	overwrites          map[string]proto.NodesOverwrites

	id_label string
}

type QueryClient struct {
	client *prometheus_ds.PrometheusV1APIClient

	current_prop   *Property
	candidate_prop *Property

	report       map[string]proto.ServerStatus
	last_updated int64

	lock *sync.Mutex
}

type pipeline_funct func(*QueryClient, interface{})
type query_funct func(*QueryClient, interface{}) (interface{}, error)

type QueryProcedureContext struct {
	last_hosts []string
	up_hosts   []string
	down_hosts []string

	network_matchers map[string][]utils.PrometheusMatcher

	results map[string]map[string]interface{} // host: {property: value}
}

func NewQueryProcedureContext() *QueryProcedureContext {
	return &QueryProcedureContext{
		results: make(map[string]map[string]interface{}),
	}
}

func load_config(cfg proto.NodesConf) *Property {
	hosts_overwrites := make(map[string]proto.NodesOverwrites)
	known_hosts := make([]string, 0)

	for _, node := range cfg.Nodes {
		hosts_overwrites[node.HostName] = node.Overwrites
		known_hosts = append(known_hosts, node.HostName)
	}

	enable_dynamic_import := false
	if cfg.Mode == string(NODES_IMPORT_AUTO) {
		enable_dynamic_import = true
	}

	global_restrictions := make([]utils.PrometheusMatcher, 0)
	for _, matcher := range cfg.GlobalMetricMatchers {
		global_restrictions = append(global_restrictions, utils.NewPrometheusMatcher(matcher.Label, matcher.Op, matcher.Value))
	}

	return &Property{
		id_label:              cfg.IdentityLabelName,
		known_hosts:           known_hosts,
		enable_dynamic_import: enable_dynamic_import,
		global_restrictions:   global_restrictions,
		overwrites:            hosts_overwrites,
	}
}

func NewAPIClient(rawcfg interface{}) *QueryClient {
	var cfg proto.NodesConf
	if err := configuration.ParseConf(rawcfg, &cfg); err != nil {
		logrus.Error(err)
		return nil
	}

	client := datasource.GetPrometheusV1APIClient(cfg.DefaultDataSource)
	if client == nil {
		logrus.Errorf("invalid datasource id %s", cfg.DefaultDataSource)
		return nil
	}

	return &QueryClient{
		lock:         &sync.Mutex{},
		client:       client,
		current_prop: load_config(cfg),
		report:       make(map[string]proto.ServerStatus),
	}
}

func (c *QueryClient) Refresh() {
	ctx := NewQueryProcedureContext()
	for _, f := range query_procedures {
		f(c, ctx)
	}
}

func (c *QueryClient) ResetConf(rawcfg interface{}) {
	//	the change of default_data_source will not be applied
	var cfg proto.NodesConf
	if err := configuration.ParseConf(rawcfg, &cfg); err != nil {
		logrus.Error(err)
		return
	}
	c.lock.Lock()
	defer c.lock.Unlock()
	c.candidate_prop = load_config(cfg)
}

func (c *QueryClient) GetReport() proto.ServerStatusList {
	c.lock.Lock()
	defer c.lock.Unlock()

	servers := make([]proto.ServerStatus, 0)
	known_servers := make(map[string]interface{})
	unknown_servers := make([]string, 0)
	//	先按配置中罗列的host节点顺序对报告输出结果进行排序
	for _, host := range c.current_prop.known_hosts {
		if server, ok := c.report[host]; ok {
			servers = append(servers, server)
		}
		known_servers[host] = struct{}{}
	}

	//	未在配置中罗列的节点，按字典序进行排序
	for k := range c.report {
		if _, ok := known_servers[k]; !ok {
			unknown_servers = append(unknown_servers, k)
		}
	}
	sort.Strings(unknown_servers)
	for _, host := range unknown_servers {
		servers = append(servers, c.report[host])
	}

	list := proto.ServerStatusList{
		UpdateTime: c.last_updated,
		Servers:    servers,
	}

	return list
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
		up_hosts, err = c.client.GetPotentialIdentifications(prometheus_ds.LabelValueQuery{
			IDLabel:      cur_env.id_label,
			Restrictions: cur_env.global_restrictions,
		})
		down_hosts = get_down_hosts(ctx.last_hosts, up_hosts)
	} else {
		//	静态模式，所有hosts由配置给出
		restrictions := make([]utils.PrometheusMatcher, 0)
		restrictions = append(restrictions, cur_env.global_restrictions...)
		restrictions = append(restrictions, utils.NewRegQuoteMetaMatchers(cur_env.id_label, cur_env.known_hosts))
		up_hosts, err = c.client.GetPotentialIdentifications(prometheus_ds.LabelValueQuery{
			IDLabel:      cur_env.id_label,
			Restrictions: restrictions,
		})
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

	network_matchers := make(map[string][]utils.PrometheusMatcher)
	for host, value := range c.current_prop.overwrites {
		if len(value.NetworkDevices) == 0 {
			continue
		}
		network_matchers[host] = []utils.PrometheusMatcher{utils.NewRegQuoteMetaMatchers(default_network_device_label, value.NetworkDevices)}
	}

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

//	以下为具体指标的查询语句
func (c *QueryClient) get_host_properties(rawctx interface{}) (interface{}, error) {
	ctx := rawctx.(*QueryProcedureContext)

	query_s := prometheus_ds.LabelValueQuery{
		IDLabel:      c.current_prop.id_label,
		IDs:          ctx.up_hosts,
		Labels:       proto.DefaultPropertyLabelMapping,
		Restrictions: c.current_prop.global_restrictions,
	}
	result := c.client.GetHostProperties(query_s)
	return result, nil
}

func (c *QueryClient) get_host_uptime(rawctx interface{}) (interface{}, error) {
	ctx := rawctx.(*QueryProcedureContext)

	uptimes, err := c.client.GetMetricValues(prometheus_ds.MetricValueQuery{
		IDLabel:      c.current_prop.id_label,
		IDs:          ctx.up_hosts,
		Query:        default_up_time_query_str,
		QueryMetrics: []string{default_up_time_source_metric},
		Restrictions: c.current_prop.global_restrictions,
	})

	return uptimes, err
}

func (c *QueryClient) get_host_load1(rawctx interface{}) (interface{}, error) {
	ctx := rawctx.(*QueryProcedureContext)

	load1, err := c.client.GetMetricValues(prometheus_ds.MetricValueQuery{
		IDLabel:      c.current_prop.id_label,
		IDs:          ctx.up_hosts,
		QueryMetrics: []string{default_load1_source_metric},
		Restrictions: c.current_prop.global_restrictions,
	})

	return load1, err
}

func (c *QueryClient) get_host_cpu_cnt(rawctx interface{}) (interface{}, error) {
	ctx := rawctx.(*QueryProcedureContext)

	cpu_restrictions := make([]utils.PrometheusMatcher, 0)
	cpu_restrictions = append(cpu_restrictions, c.current_prop.global_restrictions...)
	cpu_restrictions = append(cpu_restrictions, default_cpu_cnt_matcher)
	cpus, err := c.client.GetMetricValues(prometheus_ds.MetricValueQuery{
		IDLabel:      c.current_prop.id_label,
		IDs:          ctx.up_hosts,
		Query:        default_cpu_cnt_query_str,
		QueryMetrics: []string{default_cpu_cnt_source_metric},
		Restrictions: cpu_restrictions,
	})

	return cpus, err
}

func (c *QueryClient) get_host_memory_total(rawctx interface{}) (interface{}, error) {
	ctx := rawctx.(*QueryProcedureContext)

	memory_total, err := c.client.GetMetricValues(prometheus_ds.MetricValueQuery{
		IDLabel:      c.current_prop.id_label,
		IDs:          ctx.up_hosts,
		Query:        default_memory_total_query_str,
		QueryMetrics: []string{default_memory_total_source_metric},
		Restrictions: c.current_prop.global_restrictions,
	})

	return memory_total, err
}

func (c *QueryClient) get_host_memory_used(rawctx interface{}) (interface{}, error) {
	ctx := rawctx.(*QueryProcedureContext)

	memroy_used, err := c.client.GetMetricValues(prometheus_ds.MetricValueQuery{
		IDLabel:      c.current_prop.id_label,
		IDs:          ctx.up_hosts,
		Query:        default_memory_used_query_str,
		QueryMetrics: []string{default_memory_total_source_metric, default_memory_avail_source_metric},
		Restrictions: c.current_prop.global_restrictions,
	})

	return memroy_used, err
}

func (c *QueryClient) get_host_swap_total(rawctx interface{}) (interface{}, error) {
	ctx := rawctx.(*QueryProcedureContext)

	swap_total, err := c.client.GetMetricValues(prometheus_ds.MetricValueQuery{
		IDLabel:      c.current_prop.id_label,
		IDs:          ctx.up_hosts,
		Query:        default_swap_total_query_str,
		QueryMetrics: []string{defautl_swap_total_source_metric},
		Restrictions: c.current_prop.global_restrictions,
	})

	return swap_total, err
}

func (c *QueryClient) get_host_swap_used(rawctx interface{}) (interface{}, error) {
	ctx := rawctx.(*QueryProcedureContext)

	swap_used, err := c.client.GetMetricValues(prometheus_ds.MetricValueQuery{
		IDLabel:      c.current_prop.id_label,
		IDs:          ctx.up_hosts,
		Query:        default_swap_used_query_str,
		QueryMetrics: []string{defautl_swap_total_source_metric, default_swap_free_source_metric},
		Restrictions: c.current_prop.global_restrictions,
	})

	return swap_used, err
}

func (c *QueryClient) get_host_hdd_total(rawctx interface{}) (interface{}, error) {
	ctx := rawctx.(*QueryProcedureContext)

	hdd_restrictions := make([]utils.PrometheusMatcher, 0)
	hdd_restrictions = append(hdd_restrictions, c.current_prop.global_restrictions...)
	hdd_restrictions = append(hdd_restrictions, default_hdd_matcher)
	hdd_total, err := c.client.GetMetricValues(prometheus_ds.MetricValueQuery{
		IDLabel:      c.current_prop.id_label,
		IDs:          ctx.up_hosts,
		Query:        default_hdd_total_query_str,
		QueryMetrics: []string{default_hdd_total_source_metric},
		Restrictions: hdd_restrictions,
	})

	return hdd_total, err
}

func (c *QueryClient) get_host_hdd_used(rawctx interface{}) (interface{}, error) {
	ctx := rawctx.(*QueryProcedureContext)

	hdd_restrictions := make([]utils.PrometheusMatcher, 0)
	hdd_restrictions = append(hdd_restrictions, c.current_prop.global_restrictions...)
	hdd_restrictions = append(hdd_restrictions, default_hdd_matcher)
	hdd_used, err := c.client.GetMetricValues(prometheus_ds.MetricValueQuery{
		IDLabel:      c.current_prop.id_label,
		IDs:          ctx.up_hosts,
		Query:        default_hdd_used_query_str,
		QueryMetrics: []string{default_hdd_total_source_metric, default_hdd_free_source_metric},
		Restrictions: hdd_restrictions,
	})

	return hdd_used, err
}

func (c *QueryClient) get_host_network_rx_total(rawctx interface{}) (interface{}, error) {
	ctx := rawctx.(*QueryProcedureContext)

	network_rx_total_multi := prometheus_ds.MultiMetricValueQueryWithTimeRange{
		IDLabel:             c.current_prop.id_label,
		IDs:                 ctx.up_hosts,
		Query:               default_singe_network_sum_query_str,
		QueryMetrics:        []string{default_network_rx_source_metric},
		TimeRangeType:       proto.THIS_MONTH_SO_FAR,
		GlobalRestrictions:  c.current_prop.global_restrictions,
		DefaultRestrictions: network_default_matchers,
		CustomResctrictions: ctx.network_matchers,
	}
	network_rx_total, err := c.client.GetMetricValues(prometheus_ds.MetricValueQuery{
		Query:                  default_network_sum_query_str,
		QueryMetrics:           []string{network_rx_total_multi.String()},
		DisableDefaultMatchers: true,
	})

	return network_rx_total, err
}

func (c *QueryClient) get_host_network_tx_total(rawctx interface{}) (interface{}, error) {
	ctx := rawctx.(*QueryProcedureContext)

	network_tx_total_multi := prometheus_ds.MultiMetricValueQueryWithTimeRange{
		IDLabel:             c.current_prop.id_label,
		IDs:                 ctx.up_hosts,
		Query:               default_singe_network_sum_query_str,
		QueryMetrics:        []string{default_network_tx_source_metric},
		TimeRangeType:       proto.THIS_MONTH_SO_FAR,
		GlobalRestrictions:  c.current_prop.global_restrictions,
		DefaultRestrictions: network_default_matchers,
		CustomResctrictions: ctx.network_matchers,
	}
	network_tx_total, err := c.client.GetMetricValues(prometheus_ds.MetricValueQuery{
		Query:                  default_network_sum_query_str,
		QueryMetrics:           []string{network_tx_total_multi.String()},
		DisableDefaultMatchers: true,
	})

	return network_tx_total, err
}

func (c *QueryClient) get_host_network_rx_rate(rawctx interface{}) (interface{}, error) {
	ctx := rawctx.(*QueryProcedureContext)

	network_rx_rate_multi := prometheus_ds.MultiMetricValueQueryWithTimeRange{
		IDLabel:             c.current_prop.id_label,
		IDs:                 ctx.up_hosts,
		Query:               default_single_network_rate_query_str,
		QueryMetrics:        []string{default_network_rx_source_metric},
		GlobalRestrictions:  c.current_prop.global_restrictions,
		DefaultRestrictions: network_default_matchers,
		CustomResctrictions: ctx.network_matchers,
	}
	network_rx_rate, err := c.client.GetMetricValues(prometheus_ds.MetricValueQuery{
		Query:                  default_network_rate_query_str,
		QueryMetrics:           []string{network_rx_rate_multi.String()},
		DisableDefaultMatchers: true,
	})

	return network_rx_rate, err
}

func (c *QueryClient) get_host_network_tx_rate(rawctx interface{}) (interface{}, error) {
	ctx := rawctx.(*QueryProcedureContext)

	network_tx_rate_multi := prometheus_ds.MultiMetricValueQueryWithTimeRange{
		IDLabel:             c.current_prop.id_label,
		IDs:                 ctx.up_hosts,
		Query:               default_single_network_rate_query_str,
		QueryMetrics:        []string{default_network_tx_source_metric},
		GlobalRestrictions:  c.current_prop.global_restrictions,
		DefaultRestrictions: network_default_matchers,
		CustomResctrictions: ctx.network_matchers,
	}
	network_tx_rate, err := c.client.GetMetricValues(prometheus_ds.MetricValueQuery{
		Query:                  default_network_rate_query_str,
		QueryMetrics:           []string{network_tx_rate_multi.String()},
		DisableDefaultMatchers: true,
	})

	return network_tx_rate, err
}
