package soh

type Font struct {
	Color string `json:"color"`
	Align string `json:"align"`
}

type Node struct {
	ID     int        `json:"id"`
	Label  string     `json:"label"`
	Image  string     `json:"image"`
	Fonts  Font       `json:"font"`
	Status string     `json:"status"`
	SOH    *HostState `json:"soh"`
}

type Edge struct {
	ID     int `json:"id"`
	Source int `json:"source"`
	Target int `json:"target"`
	Length int `json:"length"`
}

type Network struct {
	Started         bool        `json:"started"`
	Nodes           []Node      `json:"nodes"`
	Edges           []Edge      `json:"edges"`
	NetworkEvents   string      `json:"networkEvents"`
	RunningCount    int         `json:"running_count"`
	NotRunningCount int         `json:"notrunning_count"`
	NotDeployCount  int         `json:"notdeploy_count"`
	NotBootCount    int         `json:"notboot_count"`
	TotalCount      int         `json:"total_count"`
	Hosts           []string    `json:"hosts"`
	HostFlows       [][]float64 `json:"host_flows"`
}

type Reachability struct {
	Hostname  string `json:"hostname" mapstructure:"hostname" structs:"hostname"`
	Timestamp string `json:"timestamp" mapstructure:"timestamp" structs:"timestamp"`
	Error     string `json:"error" mapstructure:"error" structs:"error"`
}

type Process struct {
	Process   string `json:"process" mapstructure:"process" structs:"process"`
	Timestamp string `json:"timestamp" mapstructure:"timestamp" structs:"timestamp"`
	Error     string `json:"error" mapstructure:"error" structs:"error"`
}

type Listener struct {
	Listener  string `json:"listener" mapstructure:"listener" structs:"listener"`
	Timestamp string `json:"timestamp" mapstructure:"timestamp" structs:"timestamp"`
	Error     string `json:"error" mapstructure:"error" structs:"error"`
}

type HostState struct {
	Hostname     string         `json:"hostname" mapstructure:"hostname" structs:"hostname"`
	Reachability []Reachability `json:"reachability,omitempty" mapstructure:"reachability,omitempty" structs:"reachability,omitempty"`
	Processes    []Process      `json:"processes,omitempty" mapstructure:"processes,omitempty" structs:"processes,omitempty"`
	Listeners    []Listener     `json:"listeners,omitempty" mapstructure:"listeners,omitempty" structs:"listeners,omitempty"`
}
