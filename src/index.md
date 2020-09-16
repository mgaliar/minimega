# Welcome

This is the documentation for minimega's phenix orchestration tool.

## Getting Started with phenix

### Building

Simply run `make bin/phenix`.
<br>

### Using

The following output results from `bin/phenix help`:

```
A cli application for phēnix

Usage:
  phenix [flags]
  phenix [command]

Available Commands:
  config      Configuration file management
  experiment  Experiment management
  help        Help about any command
  image       Virtual disk image management
  ui          Run the phenix UI
  util        Utility commands
  version     print version information
  vlan        Used to manage VLANs
  vm          Virtual machine management

Flags:
      --base-dir.minimega string   base minimega directory (default "/tmp/minimega")
      --base-dir.phenix string     base phenix directory (default "/phenix")
  -h, --help                       help for phenix
      --hostname-suffixes string   hostname suffixes to strip
      --log.error-file string      log fatal errors to file (default "/var/log/phenix/error.log")
      --log.error-stderr           log fatal errors to STDERR
      --store.endpoint string      endpoint for storage service (default "bolt:///etc/phenix/store.bdb")

Use "phenix [command] --help" for more information about a command.
```
<br>

Further documentation on the above can be found at:

* [config](configuration.md)
* [experiment](experiment.md)
* [image](image.md)
* [vm](vms.md)

!!! todo
    Do we need additional documentation for: ui, util, vlan?

### Store

The phenix tool uses a data store as the storage service for all of data needed throughout the various capabilities. Some important considerations are worth understanding prior to working with phenix.

1. If you are running as a standard user, the store is created in your home directory by default.
2. If you are running as a root user, the default location will be `/etc/phenix/store.bdb`.
3. It is possible to configure the store endpoint either by including the location as a flag with each command using `--store.endpoint <string>`.
4. Finally, there are global values that can be set in a YAML file; see the [config documentation](configuration.md) for more information.

## Advanced Usage

!!! todo
    `-level` discussion on the different level of logs
    
    `-store` discussion on Bolt store file
    
    Further discussion on the phenix app configuration file
