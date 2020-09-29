# Configuration Files

This is documentation for configuration files used for topology, scenario, and
experiment.

## Experiment

An experiment is a simplified configuration that includes two configurations:
`topology` and `scenario` (the latter is optional). The following is an example
of an experiment configuration; following it are the relevant topology and
scenario configuration descriptions and examples.

```
apiVersion: phenix.sandia.gov/v1
kind: Experiment
metadata:
  name: foobar
  annotations:
    topology: foobar
    scenario: foobar
```

1. Topology -- included in this configuration are the system descriptions and
configurations for each VM or container, as well as any networking settings
required to connect all of the systems together. This configuration becomes
the basis for most of the minimega commands later created in the relevant 
minimega startup script.

The following is an example of a topology configuration. This topology is made
up of three VMs and a single router. See the [Topology](#topo) section below for
further descriptions on the key values. 

```
apiVersion: phenix.sandia.gov/v1
kind: Topology
metadata:
  name: foobar
spec:
  nodes:
  - type: VirtualMachine
    general:
      hostname: host-00
      snapshot: true
    hardware:
      os_type: linux
      drives:
      - image: miniccc.qc2
    injections:
      - src: foo/bar/sucka.fish
        dst: /data/sucka.fish
      - src: /foo/bar/sucka/fish.sh
        dst: /data/fish.sh
    network:
      interfaces:
      - name: IF0
        vlan: corp
        address: 192.168.10.1
        mask: 24
        gateway: 192.168.10.254
        proto: static
        type: ethernet
      - name: IF1
        vlan: MGMT
        address: 172.16.10.1
        mask: 16
        proto: static
        type: ethernet
  - type: VirtualMachine
    general:
      hostname: host-01
      snapshot: true
    hardware:
      os_type: linux
      drives:
      - image: miniccc.qc2
    network:
      interfaces:
      - name: IF0
        vlan: corp
        address: 192.168.10.2
        mask: 24
        gateway: 192.168.10.254
        proto: static
        type: ethernet
      - name: IF1
        vlan: MGMT
        address: 172.16.10.2
        mask: 16
        proto: static
        type: ethernet
      - name: S0
        vlan: foobar
        address: 10.0.0.1
        mask: 24
        proto: static
        type: serial
        udp_port: 8989
        baud_rate: 9600
        device: /dev/ttyS0
  - type: VirtualMachine
    general:
      hostname: AD1
      snapshot: true
    hardware:
      os_type: windows
      drives:
      - image: win-svr-2k8.qc2
    network:
      interfaces:
      - name: IF0
        vlan: corp
        address: 192.168.10.250
        mask: 24
        gateway: 192.168.10.254
        proto: static
        type: ethernet
      - name: IF1
        vlan: MGMT
        address: 172.16.10.3
        mask: 16
        proto: static
        type: ethernet
  - type: Router
    labels:
      ntp-server: "true"
    general:
      hostname: router-00
      snapshot: true
    hardware:
      os_type: linux
      drives:
      - image: vyatta.qc2
    network:
      interfaces:
      - name: IF0
        vlan: corp
        address: 192.168.10.254
        mask: 24
        proto: static
        type: ethernet
        ruleset_in: test
      - name: IF1
        vlan: MGMT
        address: 172.16.10.254
        mask: 16
        proto: static
        type: ethernet
      rulesets:
      - name: test
        default: drop
        rules:
        - id: 10
          action: accept
          protocol: all
          source:
            address: 1.1.1.1
            port: 53
```

2. Scenario -- included in this configuration would be any phenix user apps
assigned to portions of the experiment or specific hosts.

The following is an example of a scenario configuration. Included in this 
scenario is an experiment-wide user app of `test-user-app` and then two host
specific apps: `protonuke` and `wireguard`. For any host with the tags of either
`protonuke` or `wireguard` (or both), these apps and the relevant configuration
will be set in the minimega startup script. See the [Scenario](#scen) section 
below for further descriptions on the key values.

```
apiVersion: phenix.sandia.gov/v1
kind: Scenario
metadata:
  name: foobar
  annotations:
    topology: foobar
spec:
  apps:
    experiment:
    - name: test-user-app
      # The map associated w/ the app name here would contain any
      # configuration details needed to configure the app.
      metadata: {}
    host:
    - name: protonuke
      hosts:
      - hostname: host-00 # hostname of topology node to apply it to
        metadata:
          # protonuke app metadata for this topology node
          args: -logfile /var/log/protonuke.log -level debug -http -https -smtp -ssh 192.168.100.100
    - name: wireguard
      hosts:
      - hostname: host-00 # hostname of topology node to apply it to
        metadata:
          # wireguard app metadata for this topology node
          infrastructure:
            private_key: GLlxWJom8cQViGHojqOUShWIZG7IsSX8
            address: 10.255.255.1/24
            listen_port: 51820
          peers:
            public_key: +joyya2F9g72qbKBtPDn00mIevG1j1OqeN76ylFLsiE=
            allowed_ips: 10.255.255.10/32
```

<a name="topo"></a>
## Topology

!!! todo
    Add some content on the topology configuration. Should we link to the JSON
    schema used for validation?

### Default Settings

If left unmodified the following are the default settings for each node:

- Memory will be set to 512MB.
- Snapshot will be set to TODO--default value
- No network settings will be included
- TODO--Additional default settings

### Required Values

- Each topology must have a unique name, these should be lowercase. Names cannot
include spaces.
- Each node in the topology must have a type of: TODO--list types (Virtual 
Machine, Router, Container).
- Each node must also have a unique name.
- Each node will need to have an OS type of: linux or windows.
- Each node will need to have a filesystem assigned. This can be either: TODO--
list options.
- TODO--Additional required values

### Optional Values

Optional values for a node in the topology configuration can include:

- Static network configurations.
- Specific memory values (e.g., 1-8GB).
- Specific VCPUs values (e.g., 1-4).
- Additional disk storage.
- File injections.
- Tags for host-based user apps.
- Routing ruleset(s).
- TODO--Additional optional values

<a name="scen"></a>
## Scenario

Scenario configurations must have (1) a unique name and (2) an associated 
topology name. Possible apps can be experiment wide or host specific apps with
included metadata unique to each app. A single scenario configuration can 
both experiment wide and host specific apps of any number. The metadata will be
specific to each app.