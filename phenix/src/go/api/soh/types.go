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
	Started         bool   `json:"started"`
	Nodes           []Node `json:"nodes"`
	Edges           []Edge `json:"edges"`
	NetworkEvents   string `json:"networkEvents"`
	RunningCount    int    `json:"running_count"`
	NotRunningCount int    `json:"notrunning_count"`
	NotDeployCount  int    `json:"notdeploy_count"`
	NotBootCount    int    `json:"notboot_count"`
	TotalCount      int    `json:"total_count"`
}

type Reachability struct {
	Hostname  string `mapstructure:"hostname" structs:"hostname"`
	Timestamp string `mapstructure:"timestamp" structs:"timestamp"`
	Error     string `mapstructure:"error" structs:"error"`
}

type Process struct {
	Process   string `mapstructure:"process" structs:"process"`
	Timestamp string `mapstructure:"timestamp" structs:"timestamp"`
	Error     string `mapstructure:"error" structs:"error"`
}

type Listener struct {
	Listener  string `mapstructure:"listener" structs:"listener"`
	Timestamp string `mapstructure:"timestamp" structs:"timestamp"`
	Error     string `mapstructure:"error" structs:"error"`
}

type HostState struct {
	Hostname     string         `mapstructure:"hostname" structs:"hostname"`
	Reachability []Reachability `mapstructure:"reachability,omitempty" structs:"reachability,omitempty"`
	Processes    []Process      `mapstructure:"processes,omitempty" structs:"processes,omitempty"`
	Listeners    []Listener     `mapstructure:"listener,omitempty" structs:"listener,omitempty"`
}
