package soh

/*
	Font, Node, Edge, Network structs, are use in the State of Health Dashboard
	to display the state of each VM and connection in an experiment.
*/

type Font struct {
	Color string `json:"color"`
	Aling string `json:"aling"`
}

type Node struct {
	ID     int    `json:"id"`
	Label  string `json:"label"`
	Image  string `json:"image"`
	Shape  string `json:"shape"`
	Fonts  Font   `json:"font"`
	Status string `json:"status"`
}

type Edge struct {
	ID     int `json:"id"`
	From   int `json:"source"`
	To     int `json:"target"`
	Length int `json:"length"`
}

type Network struct {
	Nodes           []Node `json:"nodes"`
	Edges           []Edge `json:"edges"`
	NetworkEvents   string `json:"networkEvents"`
	RunningCount    int    `json:"running_count"`
	NotRunningCount int    `json:"notrunning_count"`
	NotDeployCount  int    `json:"notdeploy_count"`
	NotBootCount    int    `json:"notboot_count"`
	TotalCount      int    `json:"total_count"`
}
