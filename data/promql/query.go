package promql

import (
	"fmt"

	"github.com/chengchung/ServerStatus/datasource/metricsql/helper"
	"github.com/chengchung/ServerStatus/proto"
	"github.com/prometheus/common/model"
	"github.com/sirupsen/logrus"
)

func (c *QueryClient) get_potential_hosts(opt helper.AlterOption) ([]string, error) {
	expr, err := helper.AlterExpr(default_static_label_source_metric, opt)
	if err != nil {
		logrus.Error(err.Error())
		return nil, nil
	}

	if result, err := c.query_metrics_values(expr); err != nil {
		return nil, err
	} else {
		up_hosts := make([]string, 0)
		for k := range result {
			up_hosts = append(up_hosts, k)
		}
		return up_hosts, nil
	}
}

func (c *QueryClient) query_metrics_values(expr string) (map[string]float64, error) {
	res := make(map[string]float64)

	logrus.Debugf("execute query %s", expr)

	model_value, warnings, err := c.client.Query(expr)
	if err != nil {
		logrus.Errorf("execute query %s error %s", expr, err)
		return nil, err
	}
	if len(warnings) > 0 {
		logrus.Warn(warnings)
	}

	values, ok := model_value.(model.Vector)
	if !ok {
		logrus.Errorf("query %s return invalid type %T", expr, model_value)
		return nil, fmt.Errorf("invalid returned type %T", model_value)
	}

	id_label := c.current_prop.id_label
	for _, value := range values {
		if id, ok := value.Metric[model.LabelName(id_label)]; !ok {
			logrus.Warnf("query %s not found label %s", expr, id_label)
		} else {
			res[string(id)] = float64(value.Value)
		}
	}

	return res, nil
}

func (c *QueryClient) query_metric_labels(expr string) (map[string]map[string]string, error) {
	logrus.Debugf("execute query %s", expr)

	res := make(map[string]map[string]string)

	model_value, warnings, err := c.client.Query(expr)
	if err != nil {
		logrus.Errorf("execute query %s error %s", expr, err)
		return nil, err
	}
	if len(warnings) > 0 {
		logrus.Warn(warnings)
	}

	values, ok := model_value.(model.Vector)
	if !ok {
		logrus.Errorf("query %s return invalid type %T", expr, model_value)
		return nil, fmt.Errorf("invalid returned type %T", model_value)
	}

	id_label := c.current_prop.id_label
	for _, value := range values {
		var host_id string
		if id, ok := value.Metric[model.LabelName(id_label)]; !ok {
			logrus.Warnf("query %s not found label %s", expr, id_label)
			continue
		} else {
			host_id = string(id)
			res[host_id] = make(map[string]string)
		}
		for k, v := range value.Metric {
			if k == model.LabelName(id_label) {
				continue
			}
			res[host_id][string(k)] = string(v)
		}
	}

	return res, nil
}

func (c *QueryClient) query_host_properties(ids []string) map[string]proto.ServerProperty {
	var opt helper.AlterOption
	opt.Filters = append(opt.Filters, c.current_prop.global_restrictions...)
	opt.Filters = append(opt.Filters, c.get_id_labels_filter(ids))
	expr, err := helper.AlterExpr(default_static_label_source_metric, opt)
	if err != nil {
		logrus.Error(err.Error())
		return nil
	}
	properties, err := c.query_metric_labels(expr)
	if err != nil {
		logrus.Error(err.Error())
		return nil
	}

	res := make(map[string]proto.ServerProperty)

	for host, property := range properties {
		var p proto.ServerProperty
		for k, v := range property {
			switch proto.ServerPropertyFields(k) {
			case proto.Type:
				p.Type = v
			case proto.Region:
				p.Region = v
			case proto.Location:
				p.Location = v
			}
		}
		res[host] = p
	}

	return res
}
