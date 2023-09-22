package promql

import (
	"github.com/VictoriaMetrics/metricsql"
)

var (
	default_static_label_source_metric = "up"
	default_id_label                   = "hostname"
)

var default_up_time_query_str = `sum(time() - node_boot_time_seconds) by (ID_LABEL)`

var default_load1_source_metric = "node_load1"

var (
	default_cpu_usage_query_str   = `100 - (avg(irate(node_cpu_seconds_total[1m])) by (ID_LABEL) * 100)`
	default_cpu_usage_idle_filter = metricsql.LabelFilter{Label: "mode", Value: "idle"}
)

var (
	default_memory_total_query_str = `node_memory_MemTotal_bytes / 1024`
	default_memory_used_query_str  = `(node_memory_MemTotal_bytes - node_memory_MemAvailable_bytes) / 1024`
)

var (
	default_swap_total_query_str = `node_memory_SwapTotal_bytes / 1024`
	default_swap_used_query_str  = `(node_memory_SwapTotal_bytes - node_memory_SwapFree_bytes) / 1024`
)

var (
	default_hdd_total_query_str = `sum(node_filesystem_size_bytes / 1024 / 1024) by (ID_LABEL)`
	default_hdd_used_query_str  = `sum((node_filesystem_size_bytes - node_filesystem_free_bytes) / 1024 / 1024) by (ID_LABEL)`
	default_hdd_filter          = metricsql.LabelFilter{Label: "fstype", Value: "ext4|xfs|ubifs", IsRegexp: true}
)

var (
	default_network_rx_source_metric = `node_network_receive_bytes_total`
	default_network_tx_source_metric = `node_network_transmit_bytes_total`

	default_single_network_rx_rate_query_str       = `irate(node_network_receive_bytes_total[2m])`
	default_single_network_tx_rate_query_str       = `irate(node_network_transmit_bytes_total[2m])`
	sum_aggr_by_query_str                          = `sum(a) by (ID_LABEL)`
	increase_rollup_query_str                      = `increase(a[0m])`
	custom_increase_rollup_total_network_query_str = `sum_over_time(a[0m])`
)

var default_network_device_label = `device`

var ONE_DAY_SECONDS = 24 * 60 * 60

var network_default_matchers []metricsql.LabelFilter = []metricsql.LabelFilter{
	{
		Label:      default_network_device_label,
		Value:      `tap.*|veth.*|br.*|docker.*|virbr*|lo*`,
		IsNegative: true,
		IsRegexp:   true,
	},
}
