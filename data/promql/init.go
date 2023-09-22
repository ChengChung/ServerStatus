package promql

import (
	"github.com/VictoriaMetrics/metricsql"
	"github.com/chengchung/ServerStatus/proto"
	"github.com/sirupsen/logrus"
)

var op_mapping map[string]metricsql.LabelFilter = map[string]metricsql.LabelFilter{
	"=":  {IsNegative: false, IsRegexp: false},
	"!=": {IsNegative: true, IsRegexp: false},
	"=~": {IsNegative: false, IsRegexp: true},
	"!~": {IsNegative: true, IsRegexp: true},
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

	global_restrictions := make([]metricsql.LabelFilter, 0)
	for _, matcher := range cfg.GlobalMetricMatchers {
		var filter metricsql.LabelFilter
		if op, ok := op_mapping[matcher.Op]; !ok {
			logrus.Errorf("invalid operator %s", matcher.Op)
			continue
		} else {
			filter = op
			filter.Label = matcher.Label
			filter.Value = matcher.Value
			global_restrictions = append(global_restrictions, filter)
		}
	}

	id_label := cfg.IdentityLabelName
	if len(id_label) == 0 {
		id_label = default_id_label
	}

	return &Property{
		id_label:              id_label,
		known_hosts:           known_hosts,
		enable_dynamic_import: enable_dynamic_import,
		global_restrictions:   global_restrictions,
		overwrites:            hosts_overwrites,
		network_overwrites:    cfg.NetworkOverwrites,
	}
}
