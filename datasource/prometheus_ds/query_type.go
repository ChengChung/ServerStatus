package prometheus_ds

import (
	"fmt"
	"strings"
	"time"

	"github.com/chengchung/ServerStatus/common/utils"
	"github.com/chengchung/ServerStatus/proto"
)

var id_label_placeholder = `##ID_LABEL##`
var time_range_placeholder = `##TIME_RANGE##`

var default_static_label_source_metric = "up"
var default_id_label = "hostname"

type LabelValueQuery struct {
	IDs          []string
	IDLabel      string
	QueryMetric  string
	Labels       map[proto.ServerPropertyFields]string
	Restrictions []utils.PrometheusMatcher
}

func (q *LabelValueQuery) AutoComplete() {
	if len(q.IDLabel) == 0 {
		q.IDLabel = default_id_label
	}
	if len(q.QueryMetric) == 0 {
		q.QueryMetric = default_static_label_source_metric
	}
}

type MetricValueQuery struct {
	IDs                    []string
	IDLabel                string
	Query                  string
	QueryMetrics           []string
	Restrictions           []utils.PrometheusMatcher
	DisableDefaultMatchers bool
}

func (q *MetricValueQuery) AutoComplete() {
	if len(q.IDLabel) == 0 {
		q.IDLabel = default_id_label
	}
	if len(q.Query) == 0 {
		q.Query = "%s"
	}

	q.Query = strings.ReplaceAll(q.Query, id_label_placeholder, q.IDLabel)
}

func (q *MetricValueQuery) String() string {
	metric_with_matchers := make([]interface{}, len(q.QueryMetrics))

	matchers := make([]utils.PrometheusMatcher, 0)
	matchers = append(matchers, q.Restrictions...)
	if !q.DisableDefaultMatchers {
		id_matchers := utils.NewRegQuoteMetaMatchers(q.IDLabel, q.IDs)
		matchers = append(matchers, id_matchers)
	}

	for idx, metric := range q.QueryMetrics {
		query := utils.QueryString{
			MetricName: metric,
			Matchers:   matchers,
		}
		metric_with_matchers[idx] = query.String()
	}

	return fmt.Sprintf(q.Query, metric_with_matchers...)
}

type MultiMetricValueQueryWithTimeRange struct {
	IDs                 []string
	IDLabel             string
	Query               string
	QueryMetrics        []string
	TimeRangeType       proto.TimeRangeType
	GlobalRestrictions  []utils.PrometheusMatcher
	DefaultRestrictions []utils.PrometheusMatcher
	CustomResctrictions map[string][]utils.PrometheusMatcher
}

func (q *MultiMetricValueQueryWithTimeRange) AutoComplete() {
	if len(q.IDLabel) == 0 {
		q.IDLabel = default_id_label
	}
	if len(q.Query) == 0 {
		q.Query = "%s"
	}

	q.Query = strings.ReplaceAll(q.Query, id_label_placeholder, q.IDLabel)
}

func (q *MultiMetricValueQueryWithTimeRange) String() string {
	q.AutoComplete()

	var multi_res string
	for idx, id := range q.IDs {
		matchers := make([]utils.PrometheusMatcher, 0)
		matchers = append(matchers, q.GlobalRestrictions...)
		if custom_matchers, ok := q.CustomResctrictions[id]; ok {
			matchers = append(matchers, custom_matchers...)
		} else {
			matchers = append(matchers, q.DefaultRestrictions...)
		}

		single := MetricValueQuery{
			IDs:          []string{id},
			IDLabel:      q.IDLabel,
			Query:        q.Query,
			QueryMetrics: q.QueryMetrics,
			Restrictions: matchers,
		}
		multi_res += single.String()

		if idx < len(q.IDs)-1 {
			multi_res += " or "
		}
	}

	if q.TimeRangeType == proto.THIS_MONTH_SO_FAR {
		t := time.Now()
		t_prev := time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location())
		diff := t.Sub(t_prev)
		tr := fmt.Sprintf("%ds", int64(diff.Seconds()))
		multi_res = strings.ReplaceAll(multi_res, time_range_placeholder, tr)
	}

	return multi_res
}
