package prometheus_ds

import (
	"fmt"

	"github.com/chengchung/ServerStatus/common/concurrency"
	"github.com/chengchung/ServerStatus/proto"
	"github.com/prometheus/common/model"
	"github.com/sirupsen/logrus"
)

func (c *PrometheusV1APIClient) GetPotentialIdentifications(para LabelValueQuery) ([]string, error) {
	para.AutoComplete()

	// query := MetricValueQuery{
	// 	QueryMetrics:           []string{default_static_label_source_metric},
	// 	Restrictions:           para.Restrictions,
	// 	DisableDefaultMatchers: true,
	// }
	// logrus.Info(query.String())
	// all_hosts, err := c.GetStaticLabelValues(query.String(), para.IDLabel)
	// if err != nil {
	// 	return nil, err
	// }

	query := MetricValueQuery{
		Query:                  "%s==1",
		QueryMetrics:           []string{default_static_label_source_metric},
		Restrictions:           para.Restrictions,
		DisableDefaultMatchers: true,
	}

	if result, err := c.GetMetricValues(query); err != nil {
		return nil, err
	} else {
		up_hosts := make([]string, 0)
		for k := range result {
			up_hosts = append(up_hosts, k)
		}
		return up_hosts, nil
	}
}

func (c *PrometheusV1APIClient) GetHostProperties(para LabelValueQuery) map[string]proto.ServerProperty {
	para.AutoComplete()

	batch := concurrency.NewBatchTask(len(para.IDs))
	for _, id := range para.IDs {
		task := batch.DispatchTask()

		go func(id string) {
			var property proto.ServerProperty
			for key, qry_label := range para.Labels {
				qr := MetricValueQuery{
					IDs:          []string{id},
					IDLabel:      para.IDLabel,
					QueryMetrics: []string{default_static_label_source_metric},
					Restrictions: para.Restrictions,
				}
				qr.AutoComplete()

				label_values, err := c.GetStaticLabelValues(qr.String(), qry_label)
				if err != nil {
					logrus.Error(err)
					continue
				}
				if len(label_values) == 0 {
					logrus.Errorf("fail to get label value %s for id %s", qry_label, id)
					continue
				}

				//	only use the first one even there is more than one value
				label_value := label_values[0]

				switch key {
				case proto.Type:
					property.Type = label_value
				case proto.Region:
					property.Region = label_value
				case proto.Location:
					property.Location = label_value
				default:
					logrus.Errorf("unknown property type %s", key)
				}
			}

			task.AnswerWithID(id, property)
		}(id)
	}

	task_results := batch.WaitForFinish()
	res := make(map[string]proto.ServerProperty)
	for _, raw := range task_results {
		result := raw.(concurrency.TaskResult)
		res[result.ID] = result.Result.(proto.ServerProperty)
	}

	return res
}

func (c *PrometheusV1APIClient) GetMetricValues(query MetricValueQuery) (map[string]float64, error) {
	query.AutoComplete()

	res := make(map[string]float64)

	query_str := query.String()
	model_value, warnings, err := c.Query(query_str)
	if err != nil {
		logrus.Errorf("execute query %s error %s", query_str, err)
		return nil, err
	}
	if len(warnings) > 0 {
		logrus.Warn(warnings)
	}

	values, ok := model_value.(model.Vector)
	if !ok {
		logrus.Errorf("query %s return invalid type %T", query_str, model_value)
		return nil, fmt.Errorf("invalid returned type %T", model_value)
	}

	for _, value := range values {
		if id, ok := value.Metric[model.LabelName(query.IDLabel)]; !ok {
			logrus.Warnf("query %s not found label %s", query_str, query.IDLabel)
		} else {
			res[string(id)] = float64(value.Value)
		}
	}

	return res, nil
}
