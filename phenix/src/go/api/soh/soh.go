package soh

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"phenix/api/experiment"
	"phenix/api/vm"

	"github.com/mitchellh/mapstructure"
	"github.com/olivere/elastic/v7"
)

var vlanAliasRegex = regexp.MustCompile(`(.*) \(\d*\)`)

func Get(expName, statusFilter string) (*Network, error) {
	// Create an empty network
	network := new(Network)

	// Create structure to format nodes' font
	font := Font{
		Color: "whitesmoke",
		Align: "center",
	}

	exp, err := experiment.Get(expName)
	if err != nil {
		return nil, fmt.Errorf("unable to get experiment %s: %w", expName, err)
	}

	// fetch all the VMs in the experiment
	vms, err := vm.List(expName)
	if err != nil {
		return nil, fmt.Errorf("getting experiment %s VMs: %w", expName, err)
	}

	status := make(map[string]*HostState)

	if exp.Running() {
		network.Started = true

		for app, data := range exp.Status.AppStatus() {
			if app == "soh" {
				var statuses []*HostState

				if err := mapstructure.Decode(data, &statuses); err != nil {
					return nil, fmt.Errorf("unable to decode state of health details: %w", err)
				}

				for _, s := range statuses {
					status[s.Hostname] = s
				}
			}
		}
	}

	// Internally use to track connections, VM's state, and whether or not the
	// VM is in minimega
	var (
		interfaces      = make(map[string]int)
		ifaceCount      = len(vms) + 1
		edgeCount       int
		runningCount    int
		notRunningCount int
		notDeployCount  int
		notBootCount    int
	)

	// Traverse the experiment VMs and create topology
	for _, vm := range vms {
		var vmState string

		if vm.Running {
			vmState = "running"
			runningCount++
		} else {
			vmState = "notrunning"
			notRunningCount++
		}

		/*
			An empty `vm.State` means the VM was not found in minimega. If the VM
			was supposed to boot (ie. DNB is false) and it's not in minimega then
			it's likely that someone has flushed it since deployment.
		*/
		if vm.State == "" {
			if vm.DoNotBoot == true {
				vmState = "notboot"
				notBootCount++
			} else {
				vmState = "notdeploy"
				notDeployCount++
			}
		}

		if statusFilter != "" && vmState != statusFilter {
			continue
		}

		node := Node{
			ID:     vm.ID,
			Label:  vm.Name,
			Image:  vm.OSType,
			Fonts:  font,
			Status: vmState,
		}

		if soh, ok := status[vm.Name]; ok {
			node.SOH = soh
		}

		network.Nodes = append(network.Nodes, node)

		// Look at the VM's interface and create an interface node, ignoring MGMT
		// VLAN
		for _, vmIface := range vm.Networks {
			if match := vlanAliasRegex.FindStringSubmatch(vmIface); match != nil {
				vmIface = match[1]
			}

			if strings.ToUpper(vmIface) == "MGMT" {
				continue
			}

			// If we got a new interface create the node
			if _, ok := interfaces[vmIface]; !ok {
				interfaces[vmIface] = ifaceCount

				node := Node{
					ID:     ifaceCount,
					Label:  vmIface,
					Image:  "Switch",
					Fonts:  font,
					Status: "ignore",
				}

				network.Nodes = append(network.Nodes, node)
				ifaceCount++
			}

			// If already exists get interface's id and connect the node
			id, _ := interfaces[vmIface]

			// create and edge for the node and interface
			edge := Edge{
				ID:     edgeCount,
				Source: vm.ID,
				Target: id,
				Length: 150,
			}

			network.Edges = append(network.Edges, edge)
			edgeCount++
		}
	}

	network.RunningCount = runningCount
	network.NotRunningCount = notRunningCount
	network.NotBootCount = notBootCount
	network.NotDeployCount = notDeployCount
	network.TotalCount = runningCount + notRunningCount + notBootCount + notDeployCount

	return network, err
}

func GetFlows(ctx context.Context, expName string) ([]string, [][]float64, error) {
	exp, err := experiment.Get(expName)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to get experiment %s: %w", expName, err)
	}

	svr := exp.Spec.Topology().FindNodesWithLabels("soh-elastic-server")

	if len(svr) == 0 {
		return nil, nil, fmt.Errorf("no SoH Elastic server found in experiment")
	}

	c := new(http.Client)

	client, err := elastic.NewClient(
		elastic.SetHttpClient(c),
		elastic.SetScheme("http"),
		elastic.SetURL(svr[0].Network().Interfaces()[0].Address()),
		elastic.SetSniff(false),
	)

	if err != nil {
		return nil, nil, fmt.Errorf("connecting to Elastic server: %w", err)
	}

	q := elastic.NewBoolQuery().Must(elastic.NewTermQuery("type", "flow"))

	result, err := client.Search().
		Index("packetbeat-*").
		Query(q).
		Size(10000).
		Do(ctx)

	if err != nil {
		return nil, nil, fmt.Errorf("querying Elastic for PacketBeat data: %w", err)
	}

	if result.Hits == nil {
		return nil, nil, fmt.Errorf("no flow data found")
	}

	if len(result.Hits.Hits) == 0 {
		return nil, nil, fmt.Errorf("no flow data found")
	}

	raw := make(map[string]map[string]float64)

	for _, hit := range result.Hits.Hits {
		var (
			src   = hit.Fields["source.ip"].(string)
			dst   = hit.Fields["destination.ip"].(string)
			bytes = hit.Fields["network.bytes"].(float64)
		)

		v, ok := raw[src]
		if !ok {
			v = make(map[string]float64)
		}

		v[dst] += bytes
		raw[src] = v
	}

	var hosts []string

	for k := range raw {
		hosts = append(hosts, k)
	}

	flows := make([][]float64, len(hosts))

	for i, s := range hosts {
		flows[i] = make([]float64, len(hosts))

		for j, d := range hosts {
			flows[i][j] = raw[s][d]
		}
	}

	return hosts, flows, nil
}
