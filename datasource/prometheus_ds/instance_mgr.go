package prometheus_ds

import (
	"sync"

	"github.com/chengchung/ServerStatus/common/configuration"
	"github.com/chengchung/ServerStatus/proto"
	"github.com/sirupsen/logrus"
)

var (
	clients map[string]*PrometheusV1APIClient
	mu      *sync.RWMutex
)

func init() {
	clients = make(map[string]*PrometheusV1APIClient)
	mu = &sync.RWMutex{}
}

func InitAPIClients(rawcfg interface{}) error {
	var cfg proto.DataSourceConf
	if err := configuration.ParseConf(rawcfg, &cfg); err != nil {
		return err
	}

	return RegisterAPIInstance(cfg.Name, cfg.Url)
}

func GetAPIInstance(id string) *PrometheusV1APIClient {
	mu.RLock()
	defer mu.RUnlock()

	return clients[id]
}

func RegisterAPIInstance(id string, url string) error {
	mu.Lock()
	defer mu.Unlock()

	if v1api, err := NewPrometheusV1APIClient(url); err != nil {
		return err
	} else {
		clients[id] = v1api
		logrus.Infof("register prometheus api client %s: %s", id, url)
	}

	return nil
}
