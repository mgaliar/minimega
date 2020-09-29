package soh

import (
	"fmt"
	"regexp"
	"strings"

	"phenix/api/experiment"
	"phenix/api/vm"
)

var vlanAliasRegex = regexp.MustCompile(`(.*) \(\d*\)`)

func Get(exp string) (*Network, error) {
	// Create an empty network
	network := new(Network)

	// Create structure to format nodes' font
	font := Font{"whitesmoke", "Center"}

	// fetch all the VMs in the experiment
	vms, err := vm.List(exp)
	if err != nil {
		return nil, fmt.Errorf("getting experiment %s VMs: %w", exp, err)
	}

	if !experiment.Running(exp) {
		return network, nil
	}

	// Internally use to track connections , VM's state, and whether or not the
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

		/*
			Depending on the VM state we set the icon.
			node struct: {ID, Name, imagePath(placeholder), IconType, fontFormat, InternalVMtype}
		*/
		if vmState == "running" {
			node := Node{vm.ID, vm.Name, "running", "image", font, "running"}
			network.Nodes = append(network.Nodes, node)
		} else if vmState == "notrunning" {
			node := Node{vm.ID, vm.Name, "notrunning", "image", font, "notrunning"}
			network.Nodes = append(network.Nodes, node)
		} else if vmState == "notboot" {
			node := Node{vm.ID, vm.Name, "notboot", "image", font, "notboot"}
			network.Nodes = append(network.Nodes, node)
		} else if vmState == "notdeploy" {
			node := Node{vm.ID, vm.Name, "notdeploy", "image", font, "notdeploy"}
			network.Nodes = append(network.Nodes, node)
		}

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
				//Create structure to format nodes' font
				font := Font{"whitesmoke", "center"}
				node := Node{ifaceCount, vmIface, "interface", "image", font, "interface"}
				network.Nodes = append(network.Nodes, node)
				ifaceCount++
			}

			// If already exists get interface's id and connect the node
			id, _ := interfaces[vmIface]

			// create and edge for the node and interface
			edge := Edge{edgeCount, vm.ID, id, 150}
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
