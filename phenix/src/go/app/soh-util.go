package app

import (
	"fmt"

	"phenix/tmpl"
	"phenix/types"
	ifaces "phenix/types/interfaces"
	"phenix/types/version"

	"github.com/mitchellh/mapstructure"
)

func (this *SOH) buildElasticServerNode(exp *types.Experiment, ip string, cidr int) (ifaces.NodeSpec, error) {
	var (
		name = this.md.PacketCapture.ElasticServer.Hostname
		tz   = "Etc/UTC"

		startupDir   = exp.Spec.BaseDir() + "/startup"
		hostnameFile = startupDir + "/" + name + "-hostname.sh"
		timezoneFile = startupDir + "/" + name + "-timezone.sh"
		ifaceFile    = startupDir + "/" + name + "-interfaces"

		elasticConfigFile = fmt.Sprintf("%s/%s-elasticsearch.yml", startupDir, name)
		kibanaConfigFile  = fmt.Sprintf("%s/%s-kibana.yml", startupDir, name)
	)

	spec := map[string]interface{}{
		"labels": map[string]string{"soh-elastic-server": "true"},
		"type":   "VirtualMachine",
		"general": map[string]interface{}{
			"hostname": name,
			"vm_type":  "kvm",
		},
		"hardware": map[string]interface{}{
			"vcpus":  4,
			"memory": 4096,
			"drives": []map[string]interface{}{
				{
					"image": this.md.PacketCapture.ElasticImage,
				},
			},
			"os_type": "linux",
		},
		"injections": []map[string]interface{}{
			{
				"src": hostnameFile,
				"dst": "/etc/phenix/startup/1_hostname-start.sh",
			},
			{
				"src": timezoneFile,
				"dst": "/etc/phenix/startup/2_timezone-start.sh",
			},
			{
				"src": ifaceFile,
				"dst": "/etc/network/interfaces",
			},
			{
				"src": elasticConfigFile,
				"dst": "/etc/elasticsearch/elasticsearch.yml",
			},
			{
				"src": kibanaConfigFile,
				"dst": "/etc/kibana/kibana.yml",
			},
		},
		"network": map[string]interface{}{
			"interfaces": []map[string]interface{}{
				{
					"name":    "IF0",
					"type":    "ethernet",
					"vlan":    "MGMT",
					"address": ip,
					"mask":    cidr,
					"proto":   "static",
					"bridge":  "phenix",
				},
			},
		},
	}

	node, _ := version.GetStoredSpecForKind("Node")

	if err := mapstructure.Decode(spec, &node); err != nil {
		return nil, fmt.Errorf("decoding node spec for Elastic server: %w", err)
	}

	if err := tmpl.CreateFileFromTemplate("linux_hostname.tmpl", name, hostnameFile); err != nil {
		return nil, fmt.Errorf("generating linux hostname config: %w", err)
	}

	if err := tmpl.CreateFileFromTemplate("linux_timezone.tmpl", tz, timezoneFile); err != nil {
		return nil, fmt.Errorf("generating linux timezone config: %w", err)
	}

	if err := tmpl.CreateFileFromTemplate("linux_interfaces.tmpl", node, ifaceFile); err != nil {
		return nil, fmt.Errorf("generating linux interfaces config: %w", err)
	}

	data := struct {
		Hostname       string
		ExperimentName string
	}{
		Hostname:       name,
		ExperimentName: exp.Spec.ExperimentName(),
	}

	if err := tmpl.CreateFileFromTemplate("elasticsearch.yml.tmpl", data, elasticConfigFile); err != nil {
		return nil, fmt.Errorf("generating elasticsearch config: %w", err)
	}

	if err := tmpl.CreateFileFromTemplate("kibana.yml.tmpl", name, kibanaConfigFile); err != nil {
		return nil, fmt.Errorf("generating kibana config: %w", err)
	}

	return node.(ifaces.NodeSpec), nil
}

func (this *SOH) buildPacketBeatNode(exp *types.Experiment, target ifaces.NodeSpec, es, ip string, cidr int) (ifaces.NodeSpec, []string, error) {
	var (
		monitored = target.General().Hostname()
		name      = monitored + "-monitor"
		tz        = "Etc/UTC"

		startupDir   = exp.Spec.BaseDir() + "/startup"
		hostnameFile = startupDir + "/" + name + "-hostname.sh"
		timezoneFile = startupDir + "/" + name + "-timezone.sh"
		ifaceFile    = startupDir + "/" + name + "-interfaces"

		packetBeatConfigFile = fmt.Sprintf("%s/%s-packetbeat.yml", startupDir, name)

		monitors []string
	)

	nets := []map[string]interface{}{
		{
			"name":    "IF0",
			"type":    "ethernet",
			"vlan":    "MGMT",
			"address": ip,
			"mask":    cidr,
			"proto":   "static",
			"bridge":  "phenix",
		},
	}

	for i, ifaceToMonitor := range this.md.PacketCapture.CaptureHosts[monitored] {
		for j, iface := range target.Network().Interfaces() {
			if iface.Name() == ifaceToMonitor {
				monitorIface := map[string]interface{}{
					"name":   fmt.Sprintf("MONITOR%d", i),
					"type":   "ethernet",
					"vlan":   iface.VLAN(),
					"proto":  "static",
					"bridge": "phenix",
				}

				nets = append(nets, monitorIface)

				monitors = append(monitors, fmt.Sprintf("%s %d", monitored, j))

				break
			}
		}
	}

	spec := map[string]interface{}{
		"labels": map[string]string{"soh-monitor-node": "true"},
		"type":   "VirtualMachine",
		"general": map[string]interface{}{
			"hostname": name,
			"vm_type":  "kvm",
		},
		"hardware": map[string]interface{}{
			"vcpus":  1,
			"memory": 512,
			"drives": []map[string]interface{}{
				{
					"image": this.md.PacketCapture.PacketBeatImage,
				},
			},
			"os_type": "linux",
		},
		"injections": []map[string]interface{}{
			{
				"src": hostnameFile,
				"dst": "/etc/phenix/startup/1_hostname-start.sh",
			},
			{
				"src": timezoneFile,
				"dst": "/etc/phenix/startup/2_timezone-start.sh",
			},
			{
				"src": ifaceFile,
				"dst": "/etc/network/interfaces",
			},
			{
				"src": packetBeatConfigFile,
				"dst": "/etc/packetbeat/packetbeat.yml",
			},
		},
		"network": map[string]interface{}{
			"interfaces": nets,
		},
	}

	node, _ := version.GetStoredSpecForKind("Node")

	if err := mapstructure.Decode(spec, &node); err != nil {
		return nil, nil, fmt.Errorf("decoding node spec for Elastic server: %w", err)
	}

	if err := tmpl.CreateFileFromTemplate("linux_hostname.tmpl", name, hostnameFile); err != nil {
		return nil, nil, fmt.Errorf("generating linux hostname config: %w", err)
	}

	if err := tmpl.CreateFileFromTemplate("linux_timezone.tmpl", tz, timezoneFile); err != nil {
		return nil, nil, fmt.Errorf("generating linux timezone config: %w", err)
	}

	if err := tmpl.CreateFileFromTemplate("linux_interfaces.tmpl", node, ifaceFile); err != nil {
		return nil, nil, fmt.Errorf("generating linux interfaces config: %w", err)
	}

	data := struct {
		ElasticServer string
		Hostname      string
	}{
		ElasticServer: es,
		Hostname:      name,
	}

	if err := tmpl.CreateFileFromTemplate("packetbeat.yml.tmpl", data, packetBeatConfigFile); err != nil {
		return nil, nil, fmt.Errorf("generating packetbeat config: %w", err)
	}

	return node.(ifaces.NodeSpec), monitors, nil
}
