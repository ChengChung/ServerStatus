package proto

type ServerProperty struct {
	Type     string `json:"type"`
	Location string `json:"location"`
	Region   string `json:"region"`
}

type ServerStatus struct {
	Name         string  `json:"name"`
	Host         string  `json:"host"`
	Online4      bool    `json:"online4"`
	Online6      bool    `json:"online6"`
	Uptime       string  `json:"uptime"`
	Load         float64 `json:"load"`
	Network_rx   int64   `json:"network_rx"`
	Network_tx   int64   `json:"network_tx"`
	Network_in   int64   `json:"network_in"`
	Network_out  int64   `json:"network_out"`
	CPU          int64   `json:"cpu"`
	Memory_total int64   `json:"memory_total"`
	Memory_used  int64   `json:"memory_used"`
	Swap_total   int64   `json:"swap_total"`
	Swap_used    int64   `json:"swap_used"`
	Hdd_total    int64   `json:"hdd_total"`
	Hdd_used     int64   `json:"hdd_used"`
	Custom       string  `json:"custom"`

	ServerProperty
}

type ServerStatusList struct {
	Servers    []ServerStatus `json:"servers"`
	UpdateTime int64          `json:"updated"`
}
