package app

import (
	"errors"
	"fmt"
	"math/rand"
	"path/filepath"
	"strings"
	"time"

	"phenix/internal/mm"
	"phenix/types"
	v1 "phenix/types/version/v1"

	"github.com/activeshadow/structs"
	"github.com/fatih/color"
	"github.com/mitchellh/mapstructure"
)

type hostProcesses struct {
	Hostname  string   `mapstructure:"hostname"`
	Processes []string `mapstructure:"processes"`
}

type hostListeners struct {
	Hostname  string   `mapstructure:"hostname"`
	Listeners []string `mapstructure:"listeners"`
}

type sohMetadata struct {
	C2Timeout     string          `mapstructure:"c2Timeout"`
	Reachability  string          `mapstructure:"testReachability"`
	SkipHosts     []string        `mapstructure:"skipHosts"`
	HostProcesses []hostProcesses `mapstructure:"hostProcesses"`
	HostListeners []hostListeners `mapstructure:"hostListeners"`

	// set after parsing
	c2Timeout time.Duration
}

func (this *sohMetadata) init() error {
	if this.Reachability == "" {
		// Default to reachability test being disabled if not specified in the
		// scenario app config.
		this.Reachability = "off"
	}

	if this.C2Timeout == "" {
		// Default C2 timeout to 5m if not specified in the scenario app config.
		this.c2Timeout = 5 * time.Minute
	} else {
		var err error

		if this.c2Timeout, err = time.ParseDuration(this.C2Timeout); err != nil {
			return fmt.Errorf("parsing C2 timeout setting '%s': %w", this.C2Timeout, err)
		}
	}

	return nil
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

type SOH struct {
	// App configuration metadata (from scenario config)
	md sohMetadata

	// Track hosts with active C2
	c2Hosts map[string]struct{}
	// Track hosts that should be tested for reachability
	// (ie. hosts that have at least one interface in an experiment VLAN)
	reachabilityHosts map[string]struct{}
	// Track IP -> Hostname mapping
	addrHosts map[string]string
	// Track VLAN -> IPs mapping
	vlans map[string][]string
	// Track hosts that failed network config test
	failedNetwork map[string]struct{}

	// Track app status for Experiment Config status
	status map[string]hostState
}

func newSOH() *SOH {
	return &SOH{
		c2Hosts:           make(map[string]struct{}),
		reachabilityHosts: make(map[string]struct{}),
		addrHosts:         make(map[string]string),
		vlans:             make(map[string][]string),
		failedNetwork:     make(map[string]struct{}),
		status:            make(map[string]hostState),
	}
}

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

func (this *SOH) PostStart(exp *types.Experiment) error {
	// *** LOAD APP CONFIGURATION METADATA *** //

	var ms map[string]interface{}

	for _, app := range exp.Spec.Scenario.Apps.Experiment {
		if app.Name == "soh" {
			ms = app.Metadata
		}
	}

	if ms == nil {
		return fmt.Errorf("soh app must have metadata defined")
	}

	if err := mapstructure.Decode(ms, &this.md); err != nil {
		return fmt.Errorf("decoding app metadata: %w", err)
	}

	if err := this.md.init(); err != nil {
		return fmt.Errorf("initializing app metadata: %w", err)
	}

	printer := color.New(color.FgBlue)

	printer.Println("  Starting SoH checks...")

	// *** WAIT FOR NODES TO HAVE NETWORKING CONFIGURED *** //

	ns := exp.Spec.ExperimentName
	wg := new(mm.ErrGroup)

	for _, node := range exp.Spec.Topology.Nodes {
		if !strings.EqualFold(node.Type, "VirtualMachine") {
			continue
		}

		if *node.General.DoNotBoot {
			continue
		}

		host := node.General.Hostname

		if skip(node, this.md.SkipHosts) {
			printer.Printf("  Skipping host %s per config\n", host)
			continue
		}

		// Assume C2 is working in this host. The host will get removed from this
		// mapping the first time C2 is proven to not be working.
		this.c2Hosts[host] = struct{}{}

		for _, iface := range node.Network.Interfaces {
			if strings.EqualFold(iface.VLAN, "MGMT") {
				continue
			}

			if iface.Type == "serial" {
				continue
			}

			this.reachabilityHosts[host] = struct{}{}
			this.addrHosts[iface.Address] = host
			this.vlans[iface.VLAN] = append(this.vlans[iface.VLAN], iface.Address)

			cidr := fmt.Sprintf("%s/%d", iface.Address, iface.Mask)

			printer.Printf("  Waiting for IP %s on host %s to be set...\n", cidr, host)

			isNetworkingConfigured(wg, ns, host, cidr, iface.Gateway)
		}
	}

	// Wait for IP address / gateway configuration to be set for each VM, as well
	// as wait for each gateway to be reachable.
	wg.Wait()

	printer = color.New(color.FgRed)

	for _, err := range wg.Errors {
		host := err.Meta["host"].(string)

		printer.Printf("  [✗] failed to confirm networking on %s: %v\n", host, err)

		if errors.Is(err, mm.ErrC2ClientNotActive) {
			delete(this.c2Hosts, host)
		} else {
			this.failedNetwork[host] = struct{}{}
		}
	}

	rand.Seed(time.Now().Unix())

	// *** RUN ACTUAL STATE OF HEALTH CHECKS *** //

	this.waitForReachabilityTest(ns)
	this.waitForProcTest(ns)
	this.waitForPortTest(ns)

	// *** WRITE RESULTS TO EXPERIMENT STATUS *** //

	if len(this.status) > 0 {
		var states []map[string]interface{}

		for _, state := range this.status {
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

func (this *SOH) waitForReachabilityTest(ns string) {
	if strings.EqualFold(this.md.Reachability, "off") {
		printer := color.New(color.FgYellow)
		printer.Println("  Reachability test is disabled")
		return
	}

	printer := color.New(color.FgBlue)

	printer.Printf("  Reachability test set to %s mode\n", this.md.Reachability)

	wg := new(mm.ErrGroup)

	for host := range this.reachabilityHosts {
		// Assume we're not skipping this host by default.
		var skipHost error

		if _, ok := this.c2Hosts[host]; !ok {
			// This host is known to not have C2 active, so don't test from it.
			skipHost = fmt.Errorf("C2 not active on host")
		}

		if _, ok := this.failedNetwork[host]; ok {
			// This host failed the network config test, so don't test from it.
			skipHost = fmt.Errorf("networking not configured on host")
		}

		for _, ips := range this.vlans {
			// Each host should try to ping a single random host in each VLAN.
			if strings.EqualFold(this.md.Reachability, "sample") {
				var targeted bool

				// Range over IPs to prevent this for-loop from going on forever if
				// all IPs in VLAN failed network connectivity test.
				for range ips {
					idx := rand.Intn(len(ips))
					targetIP := ips[idx]

					targetHost := this.addrHosts[targetIP]

					if _, ok := this.failedNetwork[targetHost]; ok {
						continue
					}

					targeted = true

					if skipHost != nil {
						wg.AddError(skipHost, map[string]interface{}{"host": host, "target": targetIP})
					} else {
						printer.Printf("  Pinging %s (%s) from host %s\n", targetHost, targetIP, host)
						pingTest(wg, ns, host, targetIP)
					}

					break
				}

				if !targeted {
					// Choose random host in VLAN to create error for.
					idx := rand.Intn(len(ips))
					targetIP := ips[idx]

					// This target host failed the network config test, so don't try
					// to do any reachability to it.
					var (
						err  = fmt.Errorf("networking not configured on target")
						meta = map[string]interface{}{"host": host, "target": targetIP}
					)

					wg.AddError(err, meta)
				}
			}

			// Each host should try to ping every host in each VLAN.
			if strings.EqualFold(this.md.Reachability, "full") {
				for _, targetIP := range ips {
					targetHost := this.addrHosts[targetIP]

					if _, ok := this.failedNetwork[targetHost]; ok {
						// This target host failed the network config test, so don't try
						// to do any reachability to it.
						var (
							err  = fmt.Errorf("networking not configured on target")
							meta = map[string]interface{}{"host": host, "target": targetIP}
						)

						wg.AddError(err, meta)
						continue
					}

					if skipHost != nil {
						wg.AddError(skipHost, map[string]interface{}{"host": host, "target": targetIP})
					} else {
						printer.Printf("  Pinging %s from host %s\n", targetIP, host)
						pingTest(wg, ns, host, targetIP)
					}
				}
			}
		}
	}

	// Wait for hosts to test reachability to other hosts.
	wg.Wait()

	printer = color.New(color.FgRed)

	for _, err := range wg.Errors {
		var (
			host   = err.Meta["host"].(string)
			target = err.Meta["target"].(string)
		)

		if errors.Is(err, mm.ErrC2ClientNotActive) {
			delete(this.c2Hosts, host)
		}

		// Convert target IP to hostname.
		hostname := this.addrHosts[target]

		r := reachability{
			Hostname:  fmt.Sprintf("%s (%s)", hostname, target),
			Timestamp: time.Now().Format(time.RFC3339),
			Error:     err.Error(),
		}

		state, ok := this.status[host]
		if !ok {
			state = hostState{Hostname: host}
		}

		state.Reachability = append(state.Reachability, r)
		this.status[host] = state

		printer.Printf("  [✗] failed to ping %s (%s) from %s\n", hostname, target, host)
	}
}

func (this *SOH) waitForProcTest(ns string) {
	wg := new(mm.ErrGroup)
	printer := color.New(color.FgBlue)

	for _, p := range this.md.HostProcesses {
		// If the host isn't in the C2 hosts map, then don't operate on it since it
		// was likely skipped for a reason.
		if _, ok := this.c2Hosts[p.Hostname]; !ok {
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
			host = err.Meta["host"].(string)
			proc = err.Meta["proc"].(string)
		)

		if errors.Is(err, mm.ErrC2ClientNotActive) {
			delete(this.c2Hosts, host)
		}

		p := process{
			Process:   proc,
			Timestamp: time.Now().Format(time.RFC3339),
			Error:     err.Error(),
		}

		state, ok := this.status[host]
		if !ok {
			state = hostState{Hostname: host}
		}

		state.Processes = append(state.Processes, p)
		this.status[host] = state

		printer.Printf("  [✗] process %s not running on host %s\n", proc, host)
	}
}

func (this *SOH) waitForPortTest(ns string) {
	wg := new(mm.ErrGroup)
	printer := color.New(color.FgBlue)

	for _, p := range this.md.HostListeners {
		// If the host isn't in the C2 hosts map, then don't operate on it since it
		// was likely skipped for a reason.
		if _, ok := this.c2Hosts[p.Hostname]; !ok {
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
			host = err.Meta["host"].(string)
			port = err.Meta["port"].(string)
		)

		if errors.Is(err, mm.ErrC2ClientNotActive) {
			delete(this.c2Hosts, host)
		}

		l := listener{
			Listener:  port,
			Timestamp: time.Now().Format(time.RFC3339),
			Error:     err.Error(),
		}

		state, ok := this.status[host]
		if !ok {
			state = hostState{Hostname: host}
		}

		state.Listeners = append(state.Listeners, l)
		this.status[host] = state

		printer.Printf("  [✗] not listening on port %s on host %s\n", port, host)
	}
}

func isNetworkingConfigured(wg *mm.ErrGroup, ns, host, addr, gateway string) {
	retryUntil := time.Now().Add(5 * time.Minute)

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
						if time.Now().After(retryUntil) {
							return fmt.Errorf("retry time expired waiting for gateway to be up")
						}

						return mm.C2RetryError{Delay: 5 * time.Second}
					}

					return nil
				}

				exec := fmt.Sprintf("ping -c 1 %s", gateway)

				cmd := &mm.C2ParallelCommand{
					Wait:     wg,
					Options:  []mm.C2Option{mm.C2NS(ns), mm.C2VM(host), mm.C2Command(exec)},
					Meta:     map[string]interface{}{"host": host},
					Expected: gwPingExpected,
				}

				mm.ScheduleC2ParallelCommand(cmd)

				return nil
			}

			cmd := &mm.C2ParallelCommand{
				Wait:     wg,
				Options:  []mm.C2Option{mm.C2NS(ns), mm.C2VM(host), mm.C2Command("ip route")},
				Meta:     map[string]interface{}{"host": host},
				Expected: gwExpected,
			}

			mm.ScheduleC2ParallelCommand(cmd)
		}

		return nil
	}

	cmd := &mm.C2ParallelCommand{
		Wait:     wg,
		Options:  []mm.C2Option{mm.C2NS(ns), mm.C2VM(host), mm.C2Command("ip addr")},
		Meta:     map[string]interface{}{"host": host},
		Expected: ipExpected,
	}

	mm.ScheduleC2ParallelCommand(cmd)
}

func pingTest(wg *mm.ErrGroup, ns, host, target string) {
	exec := fmt.Sprintf("ping -c 1 %s", target)

	cmd := &mm.C2ParallelCommand{
		Wait:    wg,
		Options: []mm.C2Option{mm.C2NS(ns), mm.C2VM(host), mm.C2Command(exec)},
		Meta:    map[string]interface{}{"host": host, "target": target},
		Expected: func(resp string) error {
			if strings.Contains(resp, "0 received") {
				return fmt.Errorf("no successful pings")
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
		Meta:    map[string]interface{}{"host": host, "proc": proc},
		Expected: func(resp string) error {
			if resp == "" {
				return fmt.Errorf("process not running")
			}

			return nil
		},
	}

	mm.ScheduleC2ParallelCommand(cmd)
}

func portTest(wg *mm.ErrGroup, ns, host, port string) {
	exec := fmt.Sprintf("ss -lntu state all 'sport = %s'", port)

	cmd := &mm.C2ParallelCommand{
		Wait:    wg,
		Options: []mm.C2Option{mm.C2NS(ns), mm.C2VM(host), mm.C2Command(exec)},
		Meta:    map[string]interface{}{"host": host, "port": port},
		Expected: func(resp string) error {
			lines := trim(resp)

			if len(lines) <= 1 {
				return fmt.Errorf("not listening on port")
			}

			return nil
		},
	}

	mm.ScheduleC2ParallelCommand(cmd)
}

func skip(node *v1.Node, toSkip []string) bool {
	for _, skipHost := range toSkip {
		// Check to see if this is a reference to an image. If so, skip this host if
		// it's using the referenced image.
		if ext := filepath.Ext(skipHost); ext == ".qc2" || ext == ".qcow2" {
			if filepath.Base(node.Hardware.Drives[0].Image) == skipHost {
				return true
			}
		}

		// Check to see if this node's hostname matches one to be skipped.
		if node.General.Hostname == skipHost {
			return true
		}
	}

	return false
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
