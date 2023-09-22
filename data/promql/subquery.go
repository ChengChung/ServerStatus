package promql

import (
	"strings"
	"time"

	"github.com/VictoriaMetrics/metricsql"
	"github.com/chengchung/ServerStatus/datasource/metricsql/helper"
	"github.com/sirupsen/logrus"
)

var (
	query_tasks map[string]query_funct = map[string]query_funct{
		"property":     (*QueryClient).get_host_properties,
		"uptime":       (*QueryClient).get_host_uptime,
		"load":         (*QueryClient).get_host_load1,
		"cpu":          (*QueryClient).get_host_cpu_usage,
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
)

func (c *QueryClient) get_host_properties(rawctx interface{}) (interface{}, error) {
	ctx := rawctx.(*QueryProcedureContext)

	result := c.query_host_properties(ctx.up_hosts)

	return result, nil
}

func (c *QueryClient) get_host_uptime(rawctx interface{}) (interface{}, error) {
	ctx := rawctx.(*QueryProcedureContext)

	var opt helper.AlterOption
	opt.Filters = append(opt.Filters, c.current_prop.global_restrictions...)
	opt.Filters = append(opt.Filters, c.get_id_labels_filter(ctx.up_hosts))
	opt.AggrBy = []string{c.current_prop.id_label}

	expr, err := helper.AlterExpr(default_up_time_query_str, opt)
	if err != nil {
		return nil, err
	}

	uptimes, err := c.query_metrics_values(expr)

	return uptimes, err
}

func (c *QueryClient) get_host_load1(rawctx interface{}) (interface{}, error) {
	ctx := rawctx.(*QueryProcedureContext)

	var opt helper.AlterOption
	opt.Filters = append(opt.Filters, c.current_prop.global_restrictions...)
	opt.Filters = append(opt.Filters, c.get_id_labels_filter(ctx.up_hosts))

	expr, err := helper.AlterExpr(default_load1_source_metric, opt)
	if err != nil {
		return nil, err
	}

	load1, err := c.query_metrics_values(expr)

	return load1, err
}

func (c *QueryClient) get_host_cpu_usage(rawctx interface{}) (interface{}, error) {
	ctx := rawctx.(*QueryProcedureContext)

	var opt helper.AlterOption
	opt.Filters = append(opt.Filters, c.current_prop.global_restrictions...)
	opt.Filters = append(opt.Filters, c.get_id_labels_filter(ctx.up_hosts))
	opt.Filters = append(opt.Filters, default_cpu_usage_idle_filter)
	opt.AggrBy = []string{c.current_prop.id_label}

	expr, err := helper.AlterExpr(default_cpu_usage_query_str, opt)
	if err != nil {
		return nil, err
	}

	cpus, err := c.query_metrics_values(expr)

	return cpus, err
}

func (c *QueryClient) get_host_memory_total(rawctx interface{}) (interface{}, error) {
	ctx := rawctx.(*QueryProcedureContext)

	var opt helper.AlterOption
	opt.Filters = append(opt.Filters, c.current_prop.global_restrictions...)
	opt.Filters = append(opt.Filters, c.get_id_labels_filter(ctx.up_hosts))

	expr, err := helper.AlterExpr(default_memory_total_query_str, opt)
	if err != nil {
		return nil, err
	}

	memory_total, err := c.query_metrics_values(expr)

	return memory_total, err
}

func (c *QueryClient) get_host_memory_used(rawctx interface{}) (interface{}, error) {
	ctx := rawctx.(*QueryProcedureContext)

	var opt helper.AlterOption
	opt.Filters = append(opt.Filters, c.current_prop.global_restrictions...)
	opt.Filters = append(opt.Filters, c.get_id_labels_filter(ctx.up_hosts))

	expr, err := helper.AlterExpr(default_memory_used_query_str, opt)
	if err != nil {
		return nil, err
	}

	memory_used, err := c.query_metrics_values(expr)

	return memory_used, err
}

func (c *QueryClient) get_host_swap_total(rawctx interface{}) (interface{}, error) {
	ctx := rawctx.(*QueryProcedureContext)

	var opt helper.AlterOption
	opt.Filters = append(opt.Filters, c.current_prop.global_restrictions...)
	opt.Filters = append(opt.Filters, c.get_id_labels_filter(ctx.up_hosts))

	expr, err := helper.AlterExpr(default_swap_total_query_str, opt)
	if err != nil {
		return nil, err
	}

	swap_total, err := c.query_metrics_values(expr)

	return swap_total, err
}

func (c *QueryClient) get_host_swap_used(rawctx interface{}) (interface{}, error) {
	ctx := rawctx.(*QueryProcedureContext)

	var opt helper.AlterOption
	opt.Filters = append(opt.Filters, c.current_prop.global_restrictions...)
	opt.Filters = append(opt.Filters, c.get_id_labels_filter(ctx.up_hosts))

	expr, err := helper.AlterExpr(default_swap_used_query_str, opt)
	if err != nil {
		return nil, err
	}

	swap_used, err := c.query_metrics_values(expr)

	return swap_used, err
}

func (c *QueryClient) get_host_hdd_total(rawctx interface{}) (interface{}, error) {
	ctx := rawctx.(*QueryProcedureContext)

	var opt helper.AlterOption
	opt.Filters = append(opt.Filters, c.current_prop.global_restrictions...)
	opt.Filters = append(opt.Filters, c.get_id_labels_filter(ctx.up_hosts))
	opt.Filters = append(opt.Filters, default_hdd_filter)
	opt.AggrBy = []string{c.current_prop.id_label}

	expr, err := helper.AlterExpr(default_hdd_total_query_str, opt)
	if err != nil {
		return nil, err
	}

	hdd_total, err := c.query_metrics_values(expr)

	return hdd_total, err
}

func (c *QueryClient) get_host_hdd_used(rawctx interface{}) (interface{}, error) {
	ctx := rawctx.(*QueryProcedureContext)

	var opt helper.AlterOption
	opt.Filters = append(opt.Filters, c.current_prop.global_restrictions...)
	opt.Filters = append(opt.Filters, c.get_id_labels_filter(ctx.up_hosts))
	opt.Filters = append(opt.Filters, default_hdd_filter)
	opt.AggrBy = []string{c.current_prop.id_label}

	expr, err := helper.AlterExpr(default_hdd_used_query_str, opt)
	if err != nil {
		return nil, err
	}

	hdd_used, err := c.query_metrics_values(expr)

	return hdd_used, err
}

func (c *QueryClient) get_host_network_rate(rawctx interface{}, base_query string) (interface{}, error) {
	ctx := rawctx.(*QueryProcedureContext)

	queries := make([]string, 0)
	for _, host := range ctx.up_hosts {
		var opt helper.AlterOption
		opt.Filters = append(opt.Filters, c.current_prop.global_restrictions...)
		opt.Filters = append(opt.Filters, c.get_id_labels_filter([]string{host}))
		if filter, ok := ctx.network_matchers[host]; ok {
			opt.Filters = append(opt.Filters, filter...)
		} else {
			opt.Filters = append(opt.Filters, network_default_matchers...)
		}

		expr, err := helper.AlterExpr(base_query, opt)
		if err != nil {
			continue
		}

		queries = append(queries, expr)
	}

	query := strings.Join(queries, " or ")
	subexpr, err := metricsql.Parse(query)
	if err != nil {
		logrus.Errorf("invalid sub query %s", query)
		return nil, err
	}

	expr, err := helper.AlterExpr(sum_aggr_by_query_str, helper.AlterOption{
		AggrBy:   []string{c.current_prop.id_label},
		AggrArgs: []metricsql.Expr{subexpr},
	})

	network_rate, err := c.query_metrics_values(expr)

	return network_rate, err
}

func (c *QueryClient) get_host_network_rx_rate(rawctx interface{}) (interface{}, error) {
	return c.get_host_network_rate(rawctx, default_single_network_rx_rate_query_str)
}

func (c *QueryClient) get_host_network_tx_rate(rawctx interface{}) (interface{}, error) {
	return c.get_host_network_rate(rawctx, default_single_network_tx_rate_query_str)
}

func (c *QueryClient) get_host_network_total(rawctx interface{}, base_query string) (interface{}, error) {
	ctx := rawctx.(*QueryProcedureContext)

	now := time.Now()

	align_to := time.Duration(0)
	if ctx.network_metric_overwrites.Enable && len(ctx.network_metric_overwrites.RangeAlign) > 0 {
		duration, err := time.ParseDuration(ctx.network_metric_overwrites.RangeAlign)
		if err != nil {
			logrus.Errorf("invalid range align %s", ctx.network_metric_overwrites.RangeAlign)
		} else {
			align_to = duration
		}
	}

	queries := make([]string, 0)
	for _, host := range ctx.up_hosts {
		//	should not raise error
		raw_metric_expr, _ := metricsql.Parse(base_query)

		ref := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
		if date, ok := ctx.host_billing_settings[host]; ok {
			ref = time.Date(now.Year(), now.Month(), date.Day(), date.Hour(), date.Minute(), date.Second(), date.Nanosecond(), date.Location())
		}

		rollup_range := duration_to_str(time_since(get_last_date_of_time(ref), align_to))

		var opt helper.AlterOption
		opt.RollupOpt = &helper.RollupOption{Window: rollup_range}
		opt.RollupExpr = raw_metric_expr

		opt.Filters = append(opt.Filters, c.current_prop.global_restrictions...)
		opt.Filters = append(opt.Filters, c.get_id_labels_filter([]string{host}))
		if filter, ok := ctx.network_matchers[host]; ok {
			opt.Filters = append(opt.Filters, filter...)
		} else {
			opt.Filters = append(opt.Filters, network_default_matchers...)
		}
		inner_expr, err := helper.AlterExpr(increase_rollup_query_str, opt)
		if err != nil {
			continue
		}

		queries = append(queries, inner_expr)
	}

	query := strings.Join(queries, " or ")

	subexpr, err := metricsql.Parse(query)
	if err != nil {
		logrus.Errorf("invalid sub query %s", query)
		return nil, err
	}

	expr, err := helper.AlterExpr(sum_aggr_by_query_str, helper.AlterOption{
		AggrBy:   []string{c.current_prop.id_label},
		AggrArgs: []metricsql.Expr{subexpr},
	})

	return c.query_metrics_values(expr)
}

func (c *QueryClient) get_host_network_rx_total(rawctx interface{}) (interface{}, error) {
	ctx := rawctx.(*QueryProcedureContext)

	m := default_network_rx_source_metric
	if ctx.network_metric_overwrites.Enable && len(ctx.network_metric_overwrites.RxMetric) > 0 {
		m = ctx.network_metric_overwrites.RxMetric
	}
	return c.get_host_network_total(rawctx, m)
}

func (c *QueryClient) get_host_network_tx_total(rawctx interface{}) (interface{}, error) {
	ctx := rawctx.(*QueryProcedureContext)

	m := default_network_tx_source_metric
	if ctx.network_metric_overwrites.Enable && len(ctx.network_metric_overwrites.TxMetric) > 0 {
		m = ctx.network_metric_overwrites.TxMetric
	}
	return c.get_host_network_total(rawctx, m)
}
