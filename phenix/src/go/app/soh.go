package app

import (
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"time"

	"phenix/internal/mm"
	"phenix/types"

	"github.com/activeshadow/structs"
	"github.com/fatih/color"
	"github.com/mitchellh/mapstructure"
)

type GroupError struct {
	Error error
	Args  []interface{}
}

type ErrGroup struct {
	sync.WaitGroup

	Errors []GroupError
}

func (this *ErrGroup) AddError(err error, args ...interface{}) {
	this.Errors = append(this.Errors, GroupError{Error: err, Args: args})
}

type SOH struct{}

func (SOH) Init(...Option) error {
	return nil
}

func (SOH) Name() string {
	return "soh"
}

func (SOH) Configure(exp *types.Experiment) error {
	return nil
}

func (SOH) PreStart(exp *types.Experiment) error {
	// TODO: inject ICMP allow into any rulesets in topology routers

	return nil
}

type hostProcesses struct {
	Hostname  string   `mapstructure:"hostname"`
	Processes []string `mapstructure:"processes"`
}

type hostListeners struct {
	Hostname  string   `mapstructure:"hostname"`
	Listeners []string `mapstructure:"listeners"`
}

type sohMetadata struct {
	Reachability  string          `mapstructure:"testReachability"`
	SkipHosts     []string        `mapstructure:"skipHosts"`
	HostProcesses []hostProcesses `mapstructure:"hostProcesses"`
	HostListeners []hostListeners `mapstructure:"hostListeners"`
}

type reachability struct {
	Hostname  string `structs:"hostname"`
	Timestamp string `structs:"timestamp"`
	Error     string `structs:"error"`
}

type process struct {
	Process   string `structs:"process"`
	Timestamp string `structs:"timestamp"`
	Error     string `structs:"error"`
}

type listener struct {
	Listener  string `structs:"listener"`
	Timestamp string `structs:"timestamp"`
	Error     string `structs:"error"`
}

type hostState struct {
	Hostname     string         `structs:"hostname"`
	Reachability []reachability `structs:"reachability,omitempty"`
	Processes    []process      `structs:"processes,omitempty"`
	Listeners    []listener     `structs:"listener,omitempty"`
}

func (SOH) PostStart(exp *types.Experiment) error {
	var (
		ms map[string]interface{}
		md sohMetadata
	)

	for _, app := range exp.Spec.Scenario.Apps.Experiment {
		if app.Name == "soh" {
			ms = app.Metadata
		}
	}

	if ms == nil {
		return fmt.Errorf("soh app must have metadata defined")
	}

	if err := mapstructure.Decode(ms, &md); err != nil {
		return fmt.Errorf("decoding metadata: %w", err)
	}

	if md.Reachability == "" {
		md.Reachability = "off"
	}

	/*
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		mm.StartC2Processor(ctx)
	*/

	printer := color.New(color.FgBlue)

	printer.Println("  Starting SoH checks...")

	var (
		ns = exp.Spec.ExperimentName
		// VMs to execute commands on (IP -> hostname mapping)
		hosts = make(map[string]string)
		// VLAN to IPs mapping
		vlans = make(map[string][]string)
	)

	wg := new(mm.ErrGroup)

	for _, node := range exp.Spec.Topology.Nodes {
		if !strings.EqualFold(node.Type, "VirtualMachine") {
			continue
		}

		if *node.General.DoNotBoot {
			continue
		}

		host := node.General.Hostname
		var skip bool

		for _, skipHost := range md.SkipHosts {
			if host == skipHost {
				skip = true
				break
			}
		}

		if skip {
			printer.Printf("  Skipping host %s per config\n", host)
			continue
		}

		for _, iface := range node.Network.Interfaces {
			if strings.EqualFold(iface.VLAN, "MGMT") {
				continue
			}

			if iface.Type == "serial" {
				continue
			}

			cidr := fmt.Sprintf("%s/%d", iface.Address, iface.Mask)

			printer.Printf("  Waiting for IP %s on host %s to be set...\n", cidr, host)

			isNetworkingConfigured(wg, ns, host, cidr, iface.Gateway)

			hosts[iface.Address] = node.General.Hostname

			vlans[iface.VLAN] = append(vlans[iface.VLAN], iface.Address)
		}
	}

	// Wait for IP address / gateway configuration to be set for each VM, as well
	// as wait for each gateway to be reachable.
	wg.Wait()

	printer = color.New(color.FgRed)

	for _, err := range wg.Errors {
		host := err.Args[0].(string)
		printer.Printf("  [✗] failed to confirm networking on %s: %v\n", host, err.Error)
	}

	rand.Seed(time.Now().Unix())
	status := make(map[string]hostState)

	if strings.EqualFold(md.Reachability, "off") {
		printer.Println("  Reachability test is disabled")
	} else {
		printer.Printf("  Reachability test set to %s mode\n", md.Reachability)

		wg := new(mm.ErrGroup)
		printer := color.New(color.FgBlue)

		for _, host := range hosts {
			for _, ips := range vlans {
				// Each host should try to ping a single random host in each VLAN.
				if strings.EqualFold(md.Reachability, "sample") {
					idx := rand.Intn(len(ips))
					target := ips[idx]

					printer.Printf("  Pinging %s from host %s\n", target, host)

					pingTest(wg, ns, host, target)
				}

				// Each host should try to ping every host in each VLAN.
				if strings.EqualFold(md.Reachability, "full") {
					for _, ip := range ips {
						printer.Printf("  Pinging %s from host %s\n", ip, host)

						pingTest(wg, ns, host, ip)
					}
				}
			}
		}

		// Wait for hosts to test reachability to other hosts.
		wg.Wait()

		printer = color.New(color.FgRed)

		for _, err := range wg.Errors {
			var (
				host   = err.Args[0].(string)
				target = err.Args[1].(string)
			)

			// Convert target IP to hostname.
			target = hosts[target]

			r := reachability{
				Hostname:  target,
				Timestamp: time.Now().Format(time.RFC3339),
				Error:     "ping failed",
			}

			state, ok := status[host]
			if !ok {
				state = hostState{Hostname: host}
			}

			state.Reachability = append(state.Reachability, r)
			status[host] = state

			printer.Printf("  [✗] failed to ping %s from %s\n", target, host)
		}
	}

	wg = new(mm.ErrGroup)
	printer = color.New(color.FgBlue)

	for _, p := range md.HostProcesses {
		var skip bool

		for _, skipHost := range md.SkipHosts {
			if p.Hostname == skipHost {
				skip = true
				break
			}
		}

		if skip {
			printer.Printf("  Skipping host %s per config\n", p.Hostname)
			continue
		}

		for _, proc := range p.Processes {
			printer.Printf("  Checking for process %s on host %s\n", proc, p.Hostname)

			procTest(wg, ns, p.Hostname, proc)
		}
	}

	wg.Wait()

	printer = color.New(color.FgRed)

	for _, err := range wg.Errors {
		var (
			host = err.Args[0].(string)
			proc = err.Args[1].(string)
		)

		p := process{
			Process:   proc,
			Timestamp: time.Now().Format(time.RFC3339),
			Error:     "process not found",
		}

		state, ok := status[host]
		if !ok {
			state = hostState{Hostname: host}
		}

		state.Processes = append(state.Processes, p)
		status[host] = state

		printer.Printf("  [✗] process %s not running on host %s\n", proc, host)
	}

	wg = new(mm.ErrGroup)
	printer = color.New(color.FgBlue)

	for _, p := range md.HostListeners {
		var skip bool

		for _, skipHost := range md.SkipHosts {
			if p.Hostname == skipHost {
				skip = true
				break
			}
		}

		if skip {
			printer.Printf("  Skipping host %s per config\n", p.Hostname)
			continue
		}

		for _, port := range p.Listeners {
			printer.Printf("  Checking for listener %s on host %s\n", port, p.Hostname)

			portTest(wg, ns, p.Hostname, port)
		}
	}

	wg.Wait()

	printer = color.New(color.FgRed)

	for _, err := range wg.Errors {
		var (
			host = err.Args[0].(string)
			port = err.Args[1].(string)
		)

		l := listener{
			Listener:  port,
			Timestamp: time.Now().Format(time.RFC3339),
			Error:     "listener not found",
		}

		state, ok := status[host]
		if !ok {
			state = hostState{Hostname: host}
		}

		state.Listeners = append(state.Listeners, l)
		status[host] = state

		printer.Printf("  [✗] not listening on port %s on host %s\n", port, host)
	}

	if len(status) > 0 {
		var states []map[string]interface{}

		for _, state := range status {
			states = append(states, structs.Map(state))
		}

		exp.Status.Apps["soh"] = states
	}

	return nil
}

func (SOH) Cleanup(exp *types.Experiment) error {
	if err := mm.ClearC2Responses(mm.C2NS(exp.Spec.ExperimentName)); err != nil {
		return fmt.Errorf("deleting minimega C2 responses: %w", err)
	}

	return nil
}

func portTest(wg *mm.ErrGroup, ns, host, port string) {
	exec := fmt.Sprintf("ss -lntu state all 'sport = %s'", port)

	cmd := &mm.C2ParallelCommand{
		Wait:    wg,
		Options: []mm.C2Option{mm.C2NS(ns), mm.C2VM(host), mm.C2Command(exec)},
		Expected: func(resp string) error {
			lines := trim(resp)

			if len(lines) <= 1 {
				return mm.NewGroupError(fmt.Errorf("not listening on port"), host, port)
			}

			return nil
		},
	}

	mm.ScheduleC2ParallelCommand(cmd)
}

func procTest(wg *mm.ErrGroup, ns, host, proc string) {
	exec := fmt.Sprintf("pgrep %s", proc)

	cmd := &mm.C2ParallelCommand{
		Wait:    wg,
		Options: []mm.C2Option{mm.C2NS(ns), mm.C2VM(host), mm.C2Command(exec)},
		Expected: func(resp string) error {
			if resp == "" {
				return mm.NewGroupError(fmt.Errorf("process not running"), host, proc)
			}

			return nil
		},
	}

	mm.ScheduleC2ParallelCommand(cmd)
}

func pingTest(wg *mm.ErrGroup, ns, host, target string) {
	exec := fmt.Sprintf("ping -c 1 %s", target)

	cmd := &mm.C2ParallelCommand{
		Wait:    wg,
		Options: []mm.C2Option{mm.C2NS(ns), mm.C2VM(host), mm.C2Command(exec)},
		Expected: func(resp string) error {
			if strings.Contains(resp, "0 received") {
				return mm.NewGroupError(fmt.Errorf("no successful pings"), host, target)
			}

			return nil
		},
	}

	mm.ScheduleC2ParallelCommand(cmd)
}

func isNetworkingConfigured(wg *mm.ErrGroup, ns, host, addr, gateway string) {
	// First, we wait for the IP address to be set on the interface. Then, we wait
	// for the default gateway to be set. Last, we wait for the default gateway to
	// be up (pingable). This is all done via nested commands streamed to the C2
	// processor within `expected` functions.
	ipExpected := func(resp string) error {
		// If `resp` doesn't contain the IP address, then the IP address isn't
		// configured yet, so keep retrying the C2 command.
		if !strings.Contains(resp, addr) {
			return mm.C2RetryError{Delay: 5 * time.Second}
		}

		if gateway != "" {
			// The IP address is now set, so schedule a C2 command for determining if
			// the default gateway is set.
			gwExpected := func(resp string) error {
				expected := fmt.Sprintf("default via %s", gateway)

				// If `resp` doesn't contain the default gateway, then the default gateway
				// isn't configured yet, so keep retrying the C2 command.
				if !strings.Contains(resp, expected) {
					return mm.C2RetryError{Delay: 5 * time.Second}
				}

				// The default gateway is now set, so schedule a C2 command for
				// determining if the default gateway is up (pingable).
				gwPingExpected := func(resp string) error {
					// If `resp` contains `0 received`, the default gateway isn't up
					// (pingable) yet, so keep retrying the C2 command.
					if strings.Contains(resp, "0 received") {
						return mm.C2RetryError{Delay: 5 * time.Second}
					}

					return nil
				}

				exec := fmt.Sprintf("ping -c 1 %s", gateway)

				cmd := &mm.C2ParallelCommand{
					Wait:     wg,
					Options:  []mm.C2Option{mm.C2NS(ns), mm.C2VM(host), mm.C2Command(exec)},
					Expected: gwPingExpected,
				}

				mm.ScheduleC2ParallelCommand(cmd)

				return nil
			}

			cmd := &mm.C2ParallelCommand{
				Wait:     wg,
				Options:  []mm.C2Option{mm.C2NS(ns), mm.C2VM(host), mm.C2Command("ip route")},
				Expected: gwExpected,
			}

			mm.ScheduleC2ParallelCommand(cmd)
		}

		return nil
	}

	cmd := &mm.C2ParallelCommand{
		Wait:     wg,
		Options:  []mm.C2Option{mm.C2NS(ns), mm.C2VM(host), mm.C2Command("ip addr")},
		Expected: ipExpected,
	}

	mm.ScheduleC2ParallelCommand(cmd)
}

func trim(str string) []string {
	var trimmed []string

	for _, l := range strings.Split(str, "\n") {
		if l == "" {
			continue
		}

		trimmed = append(trimmed, strings.TrimSpace(l))
	}

	return trimmed
}
