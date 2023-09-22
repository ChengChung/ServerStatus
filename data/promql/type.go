package promql

import (
	"sync"

	"github.com/VictoriaMetrics/metricsql"
	"github.com/chengchung/ServerStatus/datasource/prometheus/client"
	"github.com/chengchung/ServerStatus/proto"
)

type NodesImportMode string

const (
	NODES_IMPORT_AUTO   NodesImportMode = "AUTO"   //	节点将从查询中得到，动态增加
	NODES_IMPORT_STATIC NodesImportMode = "STATIC" //	节点将从文件中读取
)

type QueryClient struct {
	client *client.PrometheusV1APIClient

	current_prop *Property

	lock           *sync.Mutex
	candidate_prop *Property

	report       map[string]proto.ServerStatus
	last_updated int64
}

type Property struct {
	known_hosts           []string
	enable_dynamic_import bool

	global_restrictions []metricsql.LabelFilter
	overwrites          map[string]proto.NodesOverwrites

	id_label string
}

type pipeline_funct func(*QueryClient, interface{})
type query_funct func(*QueryClient, interface{}) (interface{}, error)

type QueryProcedureContext struct {
	last_hosts []string
	up_hosts   []string
	down_hosts []string

	network_matchers map[string][]metricsql.LabelFilter

	results map[string]map[string]interface{} // host: {property: value}
}

func NewQueryProcedureContext() *QueryProcedureContext {
	return &QueryProcedureContext{
		results: make(map[string]map[string]interface{}),
	}
}

type ServerPropertyFields string
