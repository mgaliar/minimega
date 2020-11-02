package ifaces

type ScenarioSpec interface {
	Apps() []ScenarioApp
}

type ScenarioApp interface {
	Name() string
	Metadata() map[string]interface{}
	Hosts() []ScenarioAppHost
}

type ScenarioAppHost interface {
	Hostname() string
	Metadata() map[string]interface{}
}
