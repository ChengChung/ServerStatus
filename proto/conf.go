package proto

import "encoding/json"

type MetricMatcher struct {
	Label string `json:"label"`
	Op    string `json:"op"`
	Value string `json:"value"`
}

type NodesOverwrites struct {
	DisplayName    string   `json:"hostname"`
	NetworkDevices []string `json:"net_devices"`
}

type NodeConf struct {
	HostName   string          `json:"hostname"`
	Overwrites NodesOverwrites `json:"overwrites"`
}

type NodesConf struct {
	DefaultDataSource string     `json:"default_data_source"`
	IdentityLabelName string     `json:"id_label"`
	Mode              string     `json:"mode"`
	Nodes             []NodeConf `json:"list"`

	GlobalMetricMatchers []MetricMatcher `json:"global_matcher"`
}

type BasicDataSourceConf struct {
	Type string `json:"type"`
}

type DataSourceConf struct {
	Type string `json:"type"`
	Name string `json:"name"`
	Url  string `json:"url"`
}

type MainConf struct {
	Version         uint32            `json:"version"`
	Listen          string            `json:"listen"`
	RefreshInterval uint64            `json:"refresh_interval"`
	ScrapeInterval  uint64            `json:"scrape_interval"`
	LogPath         string            `json:"log_path"`
	LogLevel        string            `json:"log_level"`
	Nodes           json.RawMessage   `json:"nodes"`
	DataSources     []json.RawMessage `json:"data_sources"`
}
