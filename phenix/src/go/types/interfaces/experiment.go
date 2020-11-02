package ifaces

import "context"

type VLANSpec interface {
	Init() error

	Aliases() map[string]int
	Min() int
	Max() int

	SetAliases(map[string]int)
	SetMin(int)
	SetMax(int)
}

type ExperimentSpec interface {
	Init() error

	ExperimentName() string
	BaseDir() string
	Topology() TopologySpec
	Scenario() ScenarioSpec
	VLANs() VLANSpec
	Schedules() map[string]string
	RunLocal() bool

	SetVLANAlias(string, int, bool) error
	SetVLANRange(int, int, bool) error
	SetSchedule(map[string]string)

	VerifyScenario(context.Context) error
	ScheduleNode(string, string) error
}

type ExperimentStatus interface {
	Init() error

	StartTime() string
	AppStatus() map[string]interface{}
	VLANs() map[string]int
	Schedules() map[string]string

	SetStartTime(string)
	SetAppStatus(string, interface{})
	SetVLANs(map[string]int)
	SetSchedule(map[string]string)

	ResetAppStatus()
}
