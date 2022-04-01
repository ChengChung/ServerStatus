package datasource

import (
	"fmt"

	"github.com/chengchung/ServerStatus/common/configuration"
	"github.com/chengchung/ServerStatus/datasource/prometheus_ds"
	"github.com/chengchung/ServerStatus/proto"
)

type t_init_funct func(interface{}) error

var init_functs map[string]t_init_funct

func init() {
	init_functs = make(map[string]t_init_funct)
	init_functs["prometheus"] = prometheus_ds.InitAPIClients
}

func InitDS(dss []interface{}) error {
	for _, rawcfg := range dss {
		var cfg proto.BasicDataSourceConf
		if err := configuration.ParseConf(rawcfg, &cfg); err == nil {
			if f, ok := init_functs[cfg.Type]; ok {
				if err := f(rawcfg); err != nil {
					return err
				}
			} else {
				return fmt.Errorf("invalid datasource type %s", cfg.Type)
			}
		} else {
			return err
		}
	}

	return nil
}

func GetPrometheusV1APIClient(id string) *prometheus_ds.PrometheusV1APIClient {
	return prometheus_ds.GetAPIInstance(id)
}
