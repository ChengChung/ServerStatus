package promql

import (
	"sort"
	"sync"

	"github.com/chengchung/ServerStatus/common/configuration"
	"github.com/chengchung/ServerStatus/datasource"
	"github.com/chengchung/ServerStatus/proto"
	"github.com/sirupsen/logrus"
)

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

func (c *QueryClient) Refresh() {
	ctx := NewQueryProcedureContext()
	for _, f := range query_procedures {
		f(c, ctx)
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
