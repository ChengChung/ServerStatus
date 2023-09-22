package promql

import (
	"fmt"
	"time"

	"github.com/VictoriaMetrics/metricsql"
	"github.com/chengchung/ServerStatus/common/utils"
)

func get_down_hosts(last []string, cur_up []string) []string {
	down_hosts := make([]string, 0)

	cur_up_map := make(map[string]interface{})
	for _, up := range cur_up {
		cur_up_map[up] = struct{}{}
	}

	for _, host := range last {
		if _, ok := cur_up_map[host]; !ok {
			down_hosts = append(down_hosts, host)
		}
	}

	return down_hosts
}

func (c *QueryClient) get_id_labels_filter(ids []string) metricsql.LabelFilter {
	return metricsql.LabelFilter{
		Label:    c.current_prop.id_label,
		Value:    utils.RegQuoteOr(ids),
		IsRegexp: true,
	}
}

func get_last_date_of_time(t time.Time) time.Time {
	now := time.Now()
	t_this_month := time.Date(now.Year(), now.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), now.Location())
	if now.After(t_this_month) {
		return t_this_month
	} else {
		return t_this_month.AddDate(0, -1, 0)
	}
}

func time_since(prev time.Time, align time.Duration) time.Duration {
	now := time.Now()
	diff := now.Sub(prev)
	if align >= time.Minute {
		//	to keep the window cover the whole time range
		diff = diff/align*align + align - time.Minute
	}

	return diff
}

func duration_to_str(d time.Duration) string {
	return fmt.Sprintf("%ds", int64(d.Seconds()))
}
