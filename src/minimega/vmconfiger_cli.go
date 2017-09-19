// Code generated by "vmconfiger -type BaseConfig,KVMConfig,ContainerConfig"; DO NOT EDIT

package main

import (
	"bytes"
	"fmt"
	"minicli"
	log "minilog"
	"os"
	"path/filepath"
	"strconv"
)

func checkPath(v string) string {
	// Ensure that relative paths are always relative to /files/
	if !filepath.IsAbs(v) {
		v = filepath.Join(*f_iomBase, v)
	}

	if _, err := os.Stat(v); os.IsNotExist(err) {
		log.Warn("file does not exist: %v", v)
	}

	return v
}

var vmconfigerCLIHandlers = []minicli.Handler{
	{
		HelpShort: "configures filesystem",
		HelpLong: `Configure the filesystem to use for launching a container. This should
be a root filesystem for a linux distribution (containing /dev, /proc,
/sys, etc.)

Note: this configuration only applies to containers and must be specified.
`,
		Patterns: []string{
			"vm config filesystem [value]",
		},
		Call: wrapSimpleCLI(func(ns *Namespace, c *minicli.Command, r *minicli.Response) error {
			if len(c.StringArgs) == 0 {
				r.Response = ns.vmConfig.FilesystemPath
				return nil
			}

			v := checkPath(c.StringArgs["value"])

			ns.vmConfig.FilesystemPath = v

			return nil
		}),
	},
	{
		HelpShort: "configures hostname",
		HelpLong: `Set a hostname for a container before launching the init program. If not
set, the hostname will be the VM name. The hostname can also be set by
the init program or other root process in the container.

Note: this configuration only applies to containers.
`,
		Patterns: []string{
			"vm config hostname [value]",
		},
		Call: wrapSimpleCLI(func(ns *Namespace, c *minicli.Command, r *minicli.Response) error {
			if len(c.StringArgs) == 0 {
				r.Response = ns.vmConfig.Hostname
				return nil
			}

			ns.vmConfig.Hostname = c.StringArgs["value"]

			return nil
		}),
	},
	{
		HelpShort: "configures init",
		HelpLong: `Set the init program and args to exec into upon container launch. This
will be PID 1 in the container.

Note: this configuration only applies to containers.

Default: "/init"
`,
		Patterns: []string{
			"vm config init [value]...",
		},
		Call: wrapSimpleCLI(func(ns *Namespace, c *minicli.Command, r *minicli.Response) error {
			if len(c.ListArgs) == 0 {
				if len(ns.vmConfig.Init) == 0 {
					return nil
				}

				r.Response = fmt.Sprintf("%v", ns.vmConfig.Init)
				return nil
			}

			ns.vmConfig.Init = c.ListArgs["value"]

			return nil
		}),
	},
	{
		HelpShort: "configures preinit",
		HelpLong: `Containers start in a highly restricted environment. vm config preinit
allows running processes before isolation mechanisms are enabled. This
occurs when the vm is launched and before the vm is put in the building
state. preinit processes must finish before the vm will be allowed to
start.

Specifically, the preinit command will be run after entering namespaces,
and mounting dependent filesystems, but before cgroups and root
capabilities are set, and before entering the chroot. This means that
the preinit command is run as root and can control the host.

For example, to run a script that enables ip forwarding, which is not
allowed during runtime because /proc is mounted read-only, add a preinit
script:

	vm config preinit enable_ip_forwarding.sh

Note: this configuration only applies to containers.
`,
		Patterns: []string{
			"vm config preinit [value]",
		},
		Call: wrapSimpleCLI(func(ns *Namespace, c *minicli.Command, r *minicli.Response) error {
			if len(c.StringArgs) == 0 {
				r.Response = ns.vmConfig.Preinit
				return nil
			}

			ns.vmConfig.Preinit = c.StringArgs["value"]

			return nil
		}),
	},
	{
		HelpShort: "configures fifos",
		HelpLong: `Set the number of named pipes to include in the container for
container-host communication. Named pipes will appear on the host in the
instance directory for the container as fifoN, and on the container as
/dev/fifos/fifoN.

Fifos are created using mkfifo() and have all of the same usage
constraints.

Note: this configuration only applies to containers.
`,
		Patterns: []string{
			"vm config fifos [value]",
		},
		Call: wrapSimpleCLI(func(ns *Namespace, c *minicli.Command, r *minicli.Response) error {
			if len(c.StringArgs) == 0 {
				r.Response = strconv.FormatUint(ns.vmConfig.Fifos, 10)
				return nil
			}

			i, err := strconv.ParseUint(c.StringArgs["value"], 10, 64)
			if err != nil {
				return err
			}

			ns.vmConfig.Fifos = i

			return nil
		}),
	},
	{
		HelpShort: "configures volume",
		HelpLong: `Attach one or more volumes to a container. These directories will be
mounted inside the container at the specified location.

For example, to mount /scratch/data to /data inside the container:

 vm config volume /data /scratch/data

Commands with the same <key> will overwrite previous volumes:

 vm config volume /data /scratch/data2
 vm config volume /data
 /scratch/data2

Note: this configuration only applies to containers.
`,
		Patterns: []string{
			"vm config volume",
			"vm config volume <key> [value]",
		},
		Call: wrapSimpleCLI(func(ns *Namespace, c *minicli.Command, r *minicli.Response) error {
			if c.StringArgs["key"] == "" {
				var b bytes.Buffer

				for k, v := range ns.vmConfig.VolumePaths {
					fmt.Fprintf(&b, "%v -> %v\n", k, v)
				}

				r.Response = b.String()
				return nil
			}

			if c.StringArgs["value"] == "" {
				if ns.vmConfig.VolumePaths != nil {
					r.Response = ns.vmConfig.VolumePaths[c.StringArgs["value"]]
				}
				return nil
			}

			if ns.vmConfig.VolumePaths == nil {
				ns.vmConfig.VolumePaths = make(map[string]string)
			}

			v := checkPath(c.StringArgs["value"])

			ns.vmConfig.VolumePaths[c.StringArgs["key"]] = v

			return nil
		}),
	},
	{
		HelpShort: "configures qemu",
		HelpLong: `Set the QEMU process to invoke. Relative paths are ok. When unspecified,
minimega uses "kvm" in the default path.

Note: this configuration only applies to KVM-based VMs.
`,
		Patterns: []string{
			"vm config qemu [value]",
		},
		Call: wrapSimpleCLI(func(ns *Namespace, c *minicli.Command, r *minicli.Response) error {
			if len(c.StringArgs) == 0 {
				r.Response = ns.vmConfig.QemuPath
				return nil
			}

			v := checkPath(c.StringArgs["value"])

			ns.vmConfig.QemuPath = v

			return nil
		}),
	},
	{
		HelpShort: "configures kernel",
		HelpLong: `Attach a kernel image to a VM. If set, QEMU will boot from this image
instead of any disk image.

Note: this configuration only applies to KVM-based VMs.
`,
		Patterns: []string{
			"vm config kernel [value]",
		},
		Call: wrapSimpleCLI(func(ns *Namespace, c *minicli.Command, r *minicli.Response) error {
			if len(c.StringArgs) == 0 {
				r.Response = ns.vmConfig.KernelPath
				return nil
			}

			v := checkPath(c.StringArgs["value"])

			ns.vmConfig.KernelPath = v

			return nil
		}),
	},
	{
		HelpShort: "configures initrd",
		HelpLong: `Attach an initrd image to a VM. Passed along with the kernel image at
boot time.

Note: this configuration only applies to KVM-based VMs.
`,
		Patterns: []string{
			"vm config initrd [value]",
		},
		Call: wrapSimpleCLI(func(ns *Namespace, c *minicli.Command, r *minicli.Response) error {
			if len(c.StringArgs) == 0 {
				r.Response = ns.vmConfig.InitrdPath
				return nil
			}

			v := checkPath(c.StringArgs["value"])

			ns.vmConfig.InitrdPath = v

			return nil
		}),
	},
	{
		HelpShort: "configures cdrom",
		HelpLong: `Attach a cdrom to a VM. When using a cdrom, it will automatically be set
to be the boot device.

Note: this configuration only applies to KVM-based VMs.
`,
		Patterns: []string{
			"vm config cdrom [value]",
		},
		Call: wrapSimpleCLI(func(ns *Namespace, c *minicli.Command, r *minicli.Response) error {
			if len(c.StringArgs) == 0 {
				r.Response = ns.vmConfig.CdromPath
				return nil
			}

			v := checkPath(c.StringArgs["value"])

			ns.vmConfig.CdromPath = v

			return nil
		}),
	},
	{
		HelpShort: "configures migrate",
		HelpLong: `Assign a migration image, generated by a previously saved VM to boot
with. By default, images are read from the files directory as specified
with -filepath. This can be overriden by using an absolute path.
Migration images should be booted with a kernel/initrd, disk, or cdrom.
Use 'vm migrate' to generate migration images from running VMs.

Note: this configuration only applies to KVM-based VMs.
`,
		Patterns: []string{
			"vm config migrate [value]",
		},
		Call: wrapSimpleCLI(func(ns *Namespace, c *minicli.Command, r *minicli.Response) error {
			if len(c.StringArgs) == 0 {
				r.Response = ns.vmConfig.MigratePath
				return nil
			}

			v := checkPath(c.StringArgs["value"])

			ns.vmConfig.MigratePath = v

			return nil
		}),
	},
	{
		HelpShort: "configures cpu",
		HelpLong: `Set the virtual CPU architecture.

By default, set to 'host' which matches the host architecture. See 'kvm
-cpu help' for a list of architectures available for your version of
kvm.

Note: this configuration only applies to KVM-based VMs.

Default: "host"
`,
		Patterns: []string{
			"vm config cpu [value]",
		},
		Call: wrapSimpleCLI(func(ns *Namespace, c *minicli.Command, r *minicli.Response) error {
			if len(c.StringArgs) == 0 {
				r.Response = ns.vmConfig.CPU
				return nil
			}

			ns.vmConfig.CPU = c.StringArgs["value"]

			return nil
		}),
	},
	{
		HelpShort: "configures serial-ports",
		HelpLong: `Specify the serial ports that will be created for the VM to use. Serial
ports specified will be mapped to the VM's /dev/ttySX device, where X
refers to the connected unix socket on the host at
$minimega_runtime/<vm_id>/serialX.

Examples:

To display current serial ports:
  vm config serial

To create three serial ports:
  vm config serial 3

Note: Whereas modern versions of Windows support up to 256 COM ports,
Linux typically only supports up to four serial devices. To use more,
make sure to pass "8250.n_uarts = 4" to the guest Linux kernel at boot.
Replace 4 with another number.
`,
		Patterns: []string{
			"vm config serial-ports [value]",
		},
		Call: wrapSimpleCLI(func(ns *Namespace, c *minicli.Command, r *minicli.Response) error {
			if len(c.StringArgs) == 0 {
				r.Response = strconv.FormatUint(ns.vmConfig.SerialPorts, 10)
				return nil
			}

			i, err := strconv.ParseUint(c.StringArgs["value"], 10, 64)
			if err != nil {
				return err
			}

			ns.vmConfig.SerialPorts = i

			return nil
		}),
	},
	{
		HelpShort: "configures virtio-ports",
		HelpLong: `Specify the virtio-serial ports that will be created for the VM to use.
Virtio-serial ports specified will be mapped to the VM's
/dev/virtio-port/<portname> device, where <portname> refers to the
connected unix socket on the host at
$minimega_runtime/<vm_id>/virtio-serialX.

Examples:

To display current virtio-serial ports:
  vm config virtio-serial

To create three virtio-serial ports:
  vm config virtio-serial 3
`,
		Patterns: []string{
			"vm config virtio-ports [value]",
		},
		Call: wrapSimpleCLI(func(ns *Namespace, c *minicli.Command, r *minicli.Response) error {
			if len(c.StringArgs) == 0 {
				r.Response = strconv.FormatUint(ns.vmConfig.VirtioPorts, 10)
				return nil
			}

			i, err := strconv.ParseUint(c.StringArgs["value"], 10, 64)
			if err != nil {
				return err
			}

			ns.vmConfig.VirtioPorts = i

			return nil
		}),
	},
	{
		HelpShort: "configures append",
		HelpLong: `Add an append string to a kernel set with vm kernel. Setting vm append
without using vm kernel will result in an error.

For example, to set a static IP for a linux VM:

	vm config append ip=10.0.0.5 gateway=10.0.0.1 netmask=255.255.255.0 dns=10.10.10.10

Note: this configuration only applies to KVM-based VMs.
`,
		Patterns: []string{
			"vm config append [value]...",
		},
		Call: wrapSimpleCLI(func(ns *Namespace, c *minicli.Command, r *minicli.Response) error {
			if len(c.ListArgs) == 0 {
				if len(ns.vmConfig.Append) == 0 {
					return nil
				}

				r.Response = fmt.Sprintf("%v", ns.vmConfig.Append)
				return nil
			}

			ns.vmConfig.Append = c.ListArgs["value"]

			return nil
		}),
	},
	{
		HelpShort: "configures disk",
		HelpLong: `Attach one or more disks to a vm. Any disk image supported by QEMU is a
valid parameter. Disk images launched in snapshot mode may safely be
used for multiple VMs.

Note: this configuration only applies to KVM-based VMs.
`,
		Patterns: []string{
			"vm config disk [value]...",
		},
		Call: wrapSimpleCLI(func(ns *Namespace, c *minicli.Command, r *minicli.Response) error {
			if len(c.ListArgs) == 0 {
				if len(ns.vmConfig.DiskPaths) == 0 {
					return nil
				}

				r.Response = fmt.Sprintf("%v", ns.vmConfig.DiskPaths)
				return nil
			}

			vals := c.ListArgs["value"]

			for i := range vals {
				vals[i] = checkPath(vals[i])
			}

			ns.vmConfig.DiskPaths = vals

			return nil
		}),
	},
	{
		HelpShort: "configures qemu-append",
		HelpLong: `Add additional arguments to be passed to the QEMU instance. For example:

	vm config qemu-append -serial tcp:localhost:4001

Note: this configuration only applies to KVM-based VMs.
`,
		Patterns: []string{
			"vm config qemu-append [value]...",
		},
		Call: wrapSimpleCLI(func(ns *Namespace, c *minicli.Command, r *minicli.Response) error {
			if len(c.ListArgs) == 0 {
				if len(ns.vmConfig.QemuAppend) == 0 {
					return nil
				}

				r.Response = fmt.Sprintf("%v", ns.vmConfig.QemuAppend)
				return nil
			}

			ns.vmConfig.QemuAppend = c.ListArgs["value"]

			return nil
		}),
	},
	{
		HelpShort: "configures uuid",
		HelpLong: `Configures the UUID for a virtual machine. If not set, the VM will be
given a random one when it is launched.
`,
		Patterns: []string{
			"vm config uuid [value]",
		},
		Call: wrapSimpleCLI(func(ns *Namespace, c *minicli.Command, r *minicli.Response) error {
			if len(c.StringArgs) == 0 {
				r.Response = ns.vmConfig.UUID
				return nil
			}

			ns.vmConfig.UUID = c.StringArgs["value"]

			return nil
		}),
	},
	{
		HelpShort: "configures vcpus",
		HelpLong: `Configures the number of virtual CPUs to allocate for a VM.

Default: 1
`,
		Patterns: []string{
			"vm config vcpus [value]",
		},
		Call: wrapSimpleCLI(func(ns *Namespace, c *minicli.Command, r *minicli.Response) error {
			if len(c.StringArgs) == 0 {
				r.Response = strconv.FormatUint(ns.vmConfig.VCPUs, 10)
				return nil
			}

			i, err := strconv.ParseUint(c.StringArgs["value"], 10, 64)
			if err != nil {
				return err
			}

			ns.vmConfig.VCPUs = i

			return nil
		}),
	},
	{
		HelpShort: "configures memory",
		HelpLong: `Configures the amount of physical memory to allocate (in megabytes).

Default: 2048
`,
		Patterns: []string{
			"vm config memory [value]",
		},
		Call: wrapSimpleCLI(func(ns *Namespace, c *minicli.Command, r *minicli.Response) error {
			if len(c.StringArgs) == 0 {
				r.Response = strconv.FormatUint(ns.vmConfig.Memory, 10)
				return nil
			}

			i, err := strconv.ParseUint(c.StringArgs["value"], 10, 64)
			if err != nil {
				return err
			}

			ns.vmConfig.Memory = i

			return nil
		}),
	},
	{
		HelpShort: "configures snapshot",
		HelpLong: `Enable or disable snapshot mode for disk images and container
filesystems. When enabled, disks/filesystems will be loaded in memory
when run and changes will not be saved. This allows a single
disk/filesystem to be used for many VMs.

Default: true
`,
		Patterns: []string{
			"vm config snapshot [true,false]",
		},
		Call: wrapSimpleCLI(func(ns *Namespace, c *minicli.Command, r *minicli.Response) error {
			if len(c.BoolArgs) == 0 {
				r.Response = strconv.FormatBool(ns.vmConfig.Snapshot)
				return nil
			}

			ns.vmConfig.Snapshot = c.BoolArgs["true"]

			return nil
		}),
	},
	{
		HelpShort: "configures schedule",
		HelpLong: `Set a host where the VM should be scheduled. This is only used when
launching VMs in a namespace.
`,
		Patterns: []string{
			"vm config schedule [value]",
		},
		Call: wrapSimpleCLI(func(ns *Namespace, c *minicli.Command, r *minicli.Response) error {
			if len(c.StringArgs) == 0 {
				r.Response = ns.vmConfig.Schedule
				return nil
			}

			ns.vmConfig.Schedule = c.StringArgs["value"]

			return nil
		}),
	},
	{
		HelpShort: "configures coschedule",
		HelpLong: `Set a limit on the number of VMs that should be scheduled on the same
host as the VM. A limit of zero means that the VM should be scheduled by
itself. A limit of -1 means that there is no limit. This is only used
when launching VMs in a namespace.

Default: -1
`,
		Patterns: []string{
			"vm config coschedule [value]",
		},
		Call: wrapSimpleCLI(func(ns *Namespace, c *minicli.Command, r *minicli.Response) error {
			if len(c.StringArgs) == 0 {
				r.Response = strconv.FormatInt(ns.vmConfig.Coschedule, 10)
				return nil
			}

			i, err := strconv.ParseInt(c.StringArgs["value"], 10, 64)
			if err != nil {
				return err
			}

			ns.vmConfig.Coschedule = i

			return nil
		}),
	},
	{
		HelpShort: "configures backchannel",
		HelpLong: `Enable/disable serial command and control layer for this VM.

Default: true
`,
		Patterns: []string{
			"vm config backchannel [true,false]",
		},
		Call: wrapSimpleCLI(func(ns *Namespace, c *minicli.Command, r *minicli.Response) error {
			if len(c.BoolArgs) == 0 {
				r.Response = strconv.FormatBool(ns.vmConfig.Backchannel)
				return nil
			}

			ns.vmConfig.Backchannel = c.BoolArgs["true"]

			return nil
		}),
	},
	{
		HelpShort: "configures tags",
		HelpLong: `Set tags in the same manner as "vm tag". These tags will apply to all
newly launched VMs.
`,
		Patterns: []string{
			"vm config tags",
			"vm config tags <key> [value]",
		},
		Call: wrapSimpleCLI(func(ns *Namespace, c *minicli.Command, r *minicli.Response) error {
			if c.StringArgs["key"] == "" {
				var b bytes.Buffer

				for k, v := range ns.vmConfig.Tags {
					fmt.Fprintf(&b, "%v -> %v\n", k, v)
				}

				r.Response = b.String()
				return nil
			}

			if c.StringArgs["value"] == "" {
				if ns.vmConfig.Tags != nil {
					r.Response = ns.vmConfig.Tags[c.StringArgs["value"]]
				}
				return nil
			}

			if ns.vmConfig.Tags == nil {
				ns.vmConfig.Tags = make(map[string]string)
			}

			ns.vmConfig.Tags[c.StringArgs["key"]] = c.StringArgs["value"]

			return nil
		}),
	},
	{
		HelpShort: "reset one or more configurations to default value",
		Patterns: []string{
			"clear vm config",
			"clear vm config <append,>",
			"clear vm config <backchannel,>",
			"clear vm config <cpu,>",
			"clear vm config <cdrom,>",
			"clear vm config <coschedule,>",
			"clear vm config <disk,>",
			"clear vm config <fifos,>",
			"clear vm config <filesystem,>",
			"clear vm config <hostname,>",
			"clear vm config <init,>",
			"clear vm config <initrd,>",
			"clear vm config <kernel,>",
			"clear vm config <memory,>",
			"clear vm config <migrate,>",
			"clear vm config <networks,>",
			"clear vm config <preinit,>",
			"clear vm config <qemu-append,>",
			"clear vm config <qemu-override,>",
			"clear vm config <qemu,>",
			"clear vm config <schedule,>",
			"clear vm config <serial-ports,>",
			"clear vm config <snapshot,>",
			"clear vm config <tags,>",
			"clear vm config <uuid,>",
			"clear vm config <vcpus,>",
			"clear vm config <virtio-ports,>",
			"clear vm config <volume,>",
		},
		Call: wrapSimpleCLI(func(ns *Namespace, c *minicli.Command, r *minicli.Response) error {
			// at most one key will be set in BoolArgs but we don't know what it
			// will be so we have to loop through the args and set whatever key we
			// see.
			mask := Wildcard
			for k := range c.BoolArgs {
				mask = k
			}

			ns.vmConfig.Clear(mask)

			return nil
		}),
	},
}

func (v *BaseConfig) Info(field string) (string, error) {
	if field == "uuid" {
		return v.UUID, nil
	}
	if field == "vcpus" {
		return strconv.FormatUint(v.VCPUs, 10), nil
	}
	if field == "memory" {
		return strconv.FormatUint(v.Memory, 10), nil
	}
	if field == "snapshot" {
		return strconv.FormatBool(v.Snapshot), nil
	}
	if field == "schedule" {
		return v.Schedule, nil
	}
	if field == "coschedule" {
		return fmt.Sprintf("%v", v.Coschedule), nil
	}
	if field == "backchannel" {
		return strconv.FormatBool(v.Backchannel), nil
	}
	if field == "networks" {
		return fmt.Sprintf("%v", v.Networks), nil
	}
	if field == "tags" {
		return fmt.Sprintf("%v", v.Tags), nil
	}

	return "", fmt.Errorf("invalid info field: %v", field)
}

func (v *BaseConfig) Clear(mask string) {
	if mask == Wildcard || mask == "uuid" {
		v.UUID = ""
	}
	if mask == Wildcard || mask == "vcpus" {
		v.VCPUs = 1
	}
	if mask == Wildcard || mask == "memory" {
		v.Memory = 2048
	}
	if mask == Wildcard || mask == "snapshot" {
		v.Snapshot = true
	}
	if mask == Wildcard || mask == "schedule" {
		v.Schedule = ""
	}
	if mask == Wildcard || mask == "coschedule" {
		v.Coschedule = -1
	}
	if mask == Wildcard || mask == "backchannel" {
		v.Backchannel = true
	}
	if mask == Wildcard || mask == "networks" {
		v.Networks = nil
	}
	if mask == Wildcard || mask == "tags" {
		v.Tags = nil
	}
}

func (v *ContainerConfig) Info(field string) (string, error) {
	if field == "filesystem" {
		return v.FilesystemPath, nil
	}
	if field == "hostname" {
		return v.Hostname, nil
	}
	if field == "init" {
		return fmt.Sprintf("%v", v.Init), nil
	}
	if field == "preinit" {
		return v.Preinit, nil
	}
	if field == "fifos" {
		return strconv.FormatUint(v.Fifos, 10), nil
	}
	if field == "volume" {
		return fmt.Sprintf("%v", v.VolumePaths), nil
	}

	return "", fmt.Errorf("invalid info field: %v", field)
}

func (v *ContainerConfig) Clear(mask string) {
	if mask == Wildcard || mask == "filesystem" {
		v.FilesystemPath = ""
	}
	if mask == Wildcard || mask == "hostname" {
		v.Hostname = ""
	}
	if mask == Wildcard || mask == "init" {
		v.Init = []string{"/init"}
	}
	if mask == Wildcard || mask == "preinit" {
		v.Preinit = ""
	}
	if mask == Wildcard || mask == "fifos" {
		v.Fifos = 0
	}
	if mask == Wildcard || mask == "volume" {
		v.VolumePaths = nil
	}
}

func (v *KVMConfig) Info(field string) (string, error) {
	if field == "qemu" {
		return v.QemuPath, nil
	}
	if field == "kernel" {
		return v.KernelPath, nil
	}
	if field == "initrd" {
		return v.InitrdPath, nil
	}
	if field == "cdrom" {
		return v.CdromPath, nil
	}
	if field == "migrate" {
		return v.MigratePath, nil
	}
	if field == "cpu" {
		return v.CPU, nil
	}
	if field == "serial-ports" {
		return strconv.FormatUint(v.SerialPorts, 10), nil
	}
	if field == "virtio-ports" {
		return strconv.FormatUint(v.VirtioPorts, 10), nil
	}
	if field == "append" {
		return fmt.Sprintf("%v", v.Append), nil
	}
	if field == "disk" {
		return fmt.Sprintf("%v", v.DiskPaths), nil
	}
	if field == "qemu-append" {
		return fmt.Sprintf("%v", v.QemuAppend), nil
	}
	if field == "qemu-override" {
		return fmt.Sprintf("%v", v.QemuOverride), nil
	}

	return "", fmt.Errorf("invalid info field: %v", field)
}

func (v *KVMConfig) Clear(mask string) {
	if mask == Wildcard || mask == "qemu" {
		v.QemuPath = ""
	}
	if mask == Wildcard || mask == "kernel" {
		v.KernelPath = ""
	}
	if mask == Wildcard || mask == "initrd" {
		v.InitrdPath = ""
	}
	if mask == Wildcard || mask == "cdrom" {
		v.CdromPath = ""
	}
	if mask == Wildcard || mask == "migrate" {
		v.MigratePath = ""
	}
	if mask == Wildcard || mask == "cpu" {
		v.CPU = "host"
	}
	if mask == Wildcard || mask == "serial-ports" {
		v.SerialPorts = 0
	}
	if mask == Wildcard || mask == "virtio-ports" {
		v.VirtioPorts = 0
	}
	if mask == Wildcard || mask == "append" {
		v.Append = nil
	}
	if mask == Wildcard || mask == "disk" {
		v.DiskPaths = nil
	}
	if mask == Wildcard || mask == "qemu-append" {
		v.QemuAppend = nil
	}
	if mask == Wildcard || mask == "qemu-override" {
		v.QemuOverride = nil
	}
}