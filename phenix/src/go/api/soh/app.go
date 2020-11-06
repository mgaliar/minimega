package soh

import (
	"errors"
	"fmt"
	"math/rand"
	"net"
	"strings"
	"time"

	"phenix/app"
	"phenix/internal/mm"
	"phenix/tmpl"
	"phenix/types"
	ifaces "phenix/types/interfaces"
	"phenix/types/version"

	"github.com/activeshadow/structs"
	"github.com/fatih/color"
	"github.com/mitchellh/mapstructure"
)

func init() {
	app.RegisterUserApp(newSOH())
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
	status map[string]HostState

	// Experiment apps to examine hosts for SoH profile data
	apps []ifaces.ScenarioApp
}

func newSOH() *SOH {
	return &SOH{
		c2Hosts:           make(map[string]struct{}),
		reachabilityHosts: make(map[string]struct{}),
		addrHosts:         make(map[string]string),
		vlans:             make(map[string][]string),
		failedNetwork:     make(map[string]struct{}),
		status:            make(map[string]HostState),
	}
}

func (SOH) Init(...app.Option) error {
	return nil
}

func (SOH) Name() string {
	return "soh"
}

func (this *SOH) Configure(exp *types.Experiment) error {
	if err := this.decodeMetadata(exp); err != nil {
		return err
	}

	if len(this.md.PacketCapture.CaptureHosts) == 0 {
		return nil
	}

	ip, mask, _ := net.ParseCIDR(this.md.PacketCapture.ElasticServer.IPAddress)
	cidr, _ := mask.Mask.Size()

	if _, err := this.buildElasticServerNode(exp, ip.String(), cidr); err != nil {
		return fmt.Errorf("building Elastic server node: %w", err)
	}

	exp.Spec.Topology().Init()

	return nil
}

func (this *SOH) PreStart(exp *types.Experiment) error {
	// TODO: inject ICMP allow into any rulesets in topology routers???

	return nil
}

func (this *SOH) deployCapture(exp *types.Experiment) error {
	if err := this.decodeMetadata(exp); err != nil {
		return err
	}

	if len(this.md.PacketCapture.CaptureHosts) == 0 {
		return nil
	}

	currentIP, mask, _ := net.ParseCIDR(this.md.PacketCapture.ElasticServer.IPAddress)
	cidr, _ := mask.Mask.Size()
	svrAddr := currentIP.String()

	var (
		caps     []ifaces.NodeSpec
		sched    = make(map[string]string)
		monitors = make(map[string][]string)
	)

	for nodeToMonitor := range this.md.PacketCapture.CaptureHosts {
		node := exp.Spec.Topology().FindNodeByName(nodeToMonitor)

		if node == nil {
			// TODO: yell loudly
			continue
		}

		ip := nextIP(currentIP)

		cap, mon, err := this.buildPacketBeatNode(exp, node, svrAddr, ip.String(), cidr)
		if err != nil {
			return fmt.Errorf("building PacketBeat node: %w", err)
		}

		caps = append(caps, cap)

		sched[cap.General().Hostname()] = exp.Status.Schedules()[nodeToMonitor]
		monitors[cap.General().Hostname()] = mon
	}

	spec := map[string]interface{}{
		"experimentName": exp.Spec.ExperimentName(),
		"topology": map[string]interface{}{
			"nodes": caps,
		},
		"schedules": sched,
	}

	expMonitor, _ := version.GetStoredSpecForKind("Experiment")

	if err := mapstructure.Decode(spec, &expMonitor); err != nil {
		return fmt.Errorf("decoding experiment spec for monitor nodes: %w", err)
	}

	data := struct {
		Exp ifaces.ExperimentSpec
		Mon map[string][]string
	}{
		Exp: expMonitor.(ifaces.ExperimentSpec),
		Mon: monitors,
	}

	filename := fmt.Sprintf("%s/mm_files/%s-monitor.mm", exp.Spec.BaseDir(), exp.Spec.ExperimentName())

	if err := tmpl.CreateFileFromTemplate("packet_capture_script.tmpl", data, filename); err != nil {
		return fmt.Errorf("generating packet capture script: %w", err)
	}

	if err := mm.ReadScriptFromFile(filename); err != nil {
		return fmt.Errorf("reading packet capture script: %w", err)
	}

	return nil
}

func (this *SOH) PostStart(exp *types.Experiment) error {
	if err := this.decodeMetadata(exp); err != nil {
		return err
	}

	this.apps = exp.Spec.Scenario().Apps()

	if err := this.deployCapture(exp); err != nil {
		return err
	}

	printer := color.New(color.FgBlue)

	printer.Println("  Starting SoH checks...")

	// *** WAIT FOR NODES TO HAVE NETWORKING CONFIGURED *** //

	ns := exp.Spec.ExperimentName()
	wg := new(mm.ErrGroup)

	for _, node := range exp.Spec.Topology().Nodes() {
		if !strings.EqualFold(node.Type(), "VirtualMachine") {
			continue
		}

		if *node.General().DoNotBoot() {
			continue
		}

		host := node.General().Hostname()

		if skip(node, this.md.SkipHosts) {
			printer.Printf("  Skipping host %s per config\n", host)
			continue
		}

		// Assume C2 is working in this host. The host will get removed from this
		// mapping the first time C2 is proven to not be working.
		this.c2Hosts[host] = struct{}{}

		for _, iface := range node.Network().Interfaces() {
			if strings.EqualFold(iface.VLAN(), "MGMT") {
				continue
			}

			if iface.Type() == "serial" {
				continue
			}

			this.reachabilityHosts[host] = struct{}{}
			this.addrHosts[iface.Address()] = host
			this.vlans[iface.VLAN()] = append(this.vlans[iface.VLAN()], iface.Address())

			if !this.md.SkipNetworkConfig {
				cidr := fmt.Sprintf("%s/%d", iface.Address(), iface.Mask())

				printer.Printf("  Waiting for IP %s on host %s to be set...\n", cidr, host)

				isNetworkingConfigured(wg, ns, host, cidr, iface.Gateway())
			}
		}
	}

	if this.md.SkipNetworkConfig {
		printer = color.New(color.FgYellow)
		printer.Println("  Skipping initial network configuration tests per config")
	}

	notifier := periodicallyNotify("waiting for initial network configurations to be validated...", 5*time.Second)

	// Wait for IP address / gateway configuration to be set for each VM, as well
	// as wait for each gateway to be reachable.
	wg.Wait()
	close(notifier)

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

		exp.Status.SetAppStatus("soh", states)
	}

	return nil
}

func (SOH) Cleanup(exp *types.Experiment) error {
	if err := mm.ClearC2Responses(mm.C2NS(exp.Spec.ExperimentName())); err != nil {
		return fmt.Errorf("deleting minimega C2 responses: %w", err)
	}

	return nil
}
