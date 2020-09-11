# Welcome

This is the documentation for minimega's phenix orchestration tool.

## Getting Started with phenix

### Building

Simply run `make bin/phenix`.

### Using

The following output results from `bin/phenix --help`:

```
minimega phenix

Global Options:
  -help
        show this help message
  -level value
        set log level: [debug, info, warn, error, fatal] [PHENIX_LEVEL] (default error)
  -logfile string
        specify file to log to [PHENIX_LOGFILE]
  -minimega-base string
        base path for minimega [PHENIX_MINIMEGA_BASE] (default "/tmp/minimega")
  -store string
        path to Bolt store file [PHENIX_STORE] (default "phenix.bdb")
  -v    log on stderr [PHENIX_V] (default true)
  -verbose
        log on stderr [PHENIX_VERBOSE] (default true)

Subcommands:
  list [all,topology,scenario,experiment] - get a list of configs
  get <kind/name>                         - get an existing config
  create <path/to/config>                 - create a new config
  edit <kind/name>                        - edit an existing config
  delete <kind/name>                      - delete a config
  experiment <start,stop> <name>          - start an existing experiment
  docs <port>                             - start documentation server on port (default 8000)
```

Right now, you can create configs and **start** an experiment (which will
simply print out the minimega script).

As an example:

```
$> bin/phenix create data/topology.yml data/scenario.yml data/experiment.yml
Topology/foo-bar-topo config created
Scenario/foo-bar-scenario config created
experiment app sink not found
host app protonuke not found
host app wireguard not found
Experiment/foobar config created
```
... or ...
```
$> bin/phenix list

+------------+----------------------+------------------+---------------------------+
|    KIND    |       VERSION        |       NAME       |          CREATED          |
+------------+----------------------+------------------+---------------------------+
| Topology   | phenix.sandia.gov/v1 | foo-bar-topo     | 2020-04-17T12:13:48-06:00 |
| Scenario   | phenix.sandia.gov/v1 | foo-bar-scenario | 2020-04-17T12:13:48-06:00 |
| Experiment | phenix.sandia.gov/v1 | foobar           | 2020-04-17T12:13:48-06:00 |
+------------+----------------------+------------------+---------------------------+
```
... or ...
```
$> bin/phenix get scenario/foo-bar-scenario
apiVersion: phenix.sandia.gov/v1
kind: Scenario
metadata:
    name: foo-bar-scenario
    created: "2020-04-17T12:13:48-06:00"
    updated: "2020-04-17T12:13:48-06:00"
    annotations:
        topology: foo-bar-topo
spec:
    apps:
        experiment:
          - metadata: {}
            name: sink
        host:
          - hosts:
              - hostname: turbine-01
                metadata:
                    args: -logfile /var/log/protonuke.log -level debug -http -https
                        -smtp -ssh 192.168.100.100
            name: protonuke
          - hosts:
              - hostname: turbine-01
                metadata:
                    infrastructure:
                        address: 10.255.255.1/24
                        listen_port: 51820
                        private_key: GLlxWJom8cQViGHojqOUShWIZG7IsSX8
                    peers:
                        allowed_ips: 10.255.255.10/32
                        public_key: +joyya2F9g72qbKBtPDn00mIevG1j1OqeN76ylFLsiE=
            name: wireguard
```
... or ...
```
$> bin/phenix experiment start foobar
namespace foobar
ns queueing true

disk snapshot bennu.qc2 0b02f5d75d22_foobar_turbine-01_snapshot 
clear vm config
vm config vcpus 1
vm config cpu Broadwell
vm config memory 512
vm config snapshot true
vm config disk 0b02f5d75d22_foobar_turbine-01_snapshot
vm config qemu-append -vga qxl
vm config net ot MGMT
vm launch kvm turbine-01

disk snapshot bennu.qc2 0b02f5d75d22_foobar_turbine-02_snapshot 
clear vm config
vm config vcpus 1
vm config cpu Broadwell
vm config memory 512
vm config snapshot true
vm config disk 0b02f5d75d22_foobar_turbine-02_snapshot
vm config qemu-append -vga qxl
vm config net ot MGMT
vm launch kvm turbine-02

$> bin/phenix experiment stop foobar
```

You can also edit configs in place via something like:

```
$> bin/phenix edit experiment/foobar
```

## Advanced Usage

!!! todo
    `-level` discussion on the different level of logs
    
    `-store` discussion on Bolt store file
    
    Further discussion on the [configuration](http://localhost:8000/configuration/) files