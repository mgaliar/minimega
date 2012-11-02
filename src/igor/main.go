// The basic structure of how commands are defined and executed are taken
// from Go's "go" tool, meaning this file is essentially brought over wholesale.

package main

import (
	"bufio"
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"time"
	"text/template"
	"unicode"
	"unicode/utf8"
)

// All of these must be replaced by a config file.
const TFTPROOT = "/home/john/tftpboot/"
const PREFIX = "kn"
const START = 1
const END = 520

var Reservations map[string][]string		// maps a reservation name to a slice of node names

// A Command is an implementation of a go command
// like go build or go fix.
type Command struct {
	// Run runs the command.
	// The args are the arguments after the command name.
	Run func(cmd *Command, args []string)

	// UsageLine is the one-line usage message.
	// The first word in the line is taken to be the command name.
	UsageLine string

	// Short is the short description shown in the 'go help' output.
	Short string

	// Long is the long message shown in the 'go help <this-command>' output.
	Long string

	// Flag is a set of flags specific to this command.
	Flag flag.FlagSet

	// CustomFlags indicates that the command will do its own
	// flag parsing.
	CustomFlags bool
}

// Name returns the command's name: the first word in the usage line.
func (c *Command) Name() string {
	name := c.UsageLine
	i := strings.Index(name, " ")
	if i >= 0 {
		name = name[:i]
	}
	return name
}

func (c *Command) Usage() {
	fmt.Fprintf(os.Stderr, "usage: %s\n\n", c.UsageLine)
	fmt.Fprintf(os.Stderr, "%s\n", strings.TrimSpace(c.Long))
	os.Exit(2)
}

// Runnable reports whether the command can be run; otherwise
// it is a documentation pseudo-command such as importpath.
func (c *Command) Runnable() bool {
	return c.Run != nil
}

// Commands lists the available commands and help topics.
// The order here is the order in which they are printed by 'go help'.
var commands = []*Command{
	cmdAdd,
	cmdSub,
}

var exitStatus = 0
var exitMu sync.Mutex

func setExitStatus(n int) {
	exitMu.Lock()
	if exitStatus < n {
		exitStatus = n
	}
	exitMu.Unlock()
}

func main() {
	flag.Usage = usage
	flag.Parse()
	log.SetFlags(0)

	args := flag.Args()
	if len(args) < 1 {
		usage()
	}

	if args[0] == "help" {
		help(args[1:])
		return
	}

	// Diagnose common mistake: GOPATH==GOROOT.
	// This setting is equivalent to not setting GOPATH at all,
	// which is not what most people want when they do it.
	if gopath := os.Getenv("GOPATH"); gopath == runtime.GOROOT() {
		fmt.Fprintf(os.Stderr, "warning: GOPATH set to GOROOT (%s) has no effect\n", gopath)
	}

	for _, cmd := range commands {
		if cmd.Name() == args[0] && cmd.Run != nil {
			cmd.Flag.Usage = func() { cmd.Usage() }
			if cmd.CustomFlags {
				args = args[1:]
			} else {
				cmd.Flag.Parse(args[1:])
				args = cmd.Flag.Args()
			}
			cmd.Run(cmd, args)
			exit()
			return
		}
	}

	fmt.Fprintf(os.Stderr, "go: unknown subcommand %q\nRun 'go help' for usage.\n", args[0])
	setExitStatus(2)
	exit()
}

var usageTemplate = `Igor is a scheduler for Mega-style clusters.

Usage:

	igor command [arguments]

The commands are:
{{range .}}{{if .Runnable}}
    {{.Name | printf "%-11s"}} {{.Short}}{{end}}{{end}}

Use "igor help [command]" for more information about a command.

Additional help topics:
{{range .}}{{if not .Runnable}}
    {{.Name | printf "%-11s"}} {{.Short}}{{end}}{{end}}

Use "igor help [topic]" for more information about that topic.

`

var helpTemplate = `{{if .Runnable}}usage: igor {{.UsageLine}}

{{end}}{{.Long | trim}}
`

var documentationTemplate = `/*
{{range .}}{{if .Short}}{{.Short | capitalize}}

{{end}}{{if .Runnable}}Usage:

	igor {{.UsageLine}}

{{end}}{{.Long | trim}}


{{end}}*/
package documentation

`

// tmpl executes the given template text on data, writing the result to w.
func tmpl(w io.Writer, text string, data interface{}) {
	t := template.New("top")
	t.Funcs(template.FuncMap{"trim": strings.TrimSpace, "capitalize": capitalize})
	template.Must(t.Parse(text))
	if err := t.Execute(w, data); err != nil {
		panic(err)
	}
}

func capitalize(s string) string {
	if s == "" {
		return s
	}
	r, n := utf8.DecodeRuneInString(s)
	return string(unicode.ToTitle(r)) + s[n:]
}

func printUsage(w io.Writer) {
	tmpl(w, usageTemplate, commands)
}

func usage() {
	printUsage(os.Stderr)
	os.Exit(2)
}

// help implements the 'help' command.
func help(args []string) {
	if len(args) == 0 {
		printUsage(os.Stdout)
		// not exit 2: succeeded at 'go help'.
		return
	}
	if len(args) != 1 {
		fmt.Fprintf(os.Stderr, "usage: go help command\n\nToo many arguments given.\n")
		os.Exit(2) // failed at 'go help'
	}

	arg := args[0]

	// 'go help documentation' generates doc.go.
	if arg == "documentation" {
		buf := new(bytes.Buffer)
		printUsage(buf)
		usage := &Command{Long: buf.String()}
		tmpl(os.Stdout, documentationTemplate, append([]*Command{usage}, commands...))
		return
	}

	for _, cmd := range commands {
		if cmd.Name() == arg {
			tmpl(os.Stdout, helpTemplate, cmd)
			// not exit 2: succeeded at 'go help cmd'.
			return
		}
	}

	fmt.Fprintf(os.Stderr, "Unknown help topic %#q.  Run 'go help'.\n", arg)
	os.Exit(2) // failed at 'go help cmd'
}

var atexitFuncs []func()

func atexit(f func()) {
	atexitFuncs = append(atexitFuncs, f)
}

func exit() {
	for _, f := range atexitFuncs {
		f()
	}
	os.Exit(exitStatus)
}

func fatalf(format string, args ...interface{}) {
	errorf(format, args...)
	exit()
}

func errorf(format string, args ...interface{}) {
	log.Printf(format, args...)
	setExitStatus(1)
}

type Reservation struct {
	name	string
	pxenames	[]string  // eg C000025B
	timeleft	time.Duration
}

func readResName(filename string) (string, error) {
	tmp, err := os.Open(filename)
	if err != nil { fatalf("failed to open %v: %v", filename, err) }
	defer tmp.Close()
	filebuf := bufio.NewReader(tmp)
	line, _ := filebuf.ReadString('\n')
	line = strings.Replace(line, "\n", "", -1)
	parts := strings.Split(line, " ")
	if len(parts) != 2 || parts[0] != "default" {
		return "", errors.New("bad format")
	}
	return parts[1], nil
}

func addNode(reservations *[]Reservation, root, name string) {
	var resname string
	found := false

	// validate the filename, only continue if it's a pxe config file
	if matched, _ := regexp.Match("[0-9A-F]{8}", []byte(name)); !matched {
		return
	}

	resname, err := readResName(root+name)
	if err != nil {
		return
	}

	for _, r := range *reservations {
		if r.name == resname {
			found = true
			r.pxenames = append(r.pxenames, name)
		}
	}
	if !found {
		expiretime := time.Duration(0) * time.Second
		expirepath := TFTPROOT + "/igor/" + resname + "-expires"
		contents, err := ioutil.ReadFile(expirepath)
		cstring := string(contents)
		cstring = strings.Replace(cstring, "\n", "", -1)
		if err == nil {
			timefmt := "2006-01-02 15:04:05.999999999 -0700 MST"
			fmt.Println(timefmt)
			fmt.Println(cstring)
			expdate, err := time.Parse(timefmt, string(cstring))
			if err != nil { 
				log.Printf("couldn't parse expiration time for reservation %v\n", resname);
			} else {
				expiretime = expdate.Sub(time.Now())
			}
		}
		r := Reservation{ name: resname, pxenames: []string{ name }, timeleft: expiretime }
		*reservations = append(*reservations, r)
	}
}

func analyzeReservations() []Reservation {
	var ret []Reservation
	pxeconfig := TFTPROOT + "pxelinux.cfg/"

	f, err := os.Open(pxeconfig)
	if err != nil {
		fatalf("failed to open directory %v: %v", pxeconfig, err)
	}
	defer f.Close()

	files, err := f.Readdirnames(-1)
	if err != nil {
		fatalf("failed to read entries of %v: %v", pxeconfig, err)
	}

	for _, name := range files {
		addNode(&ret, pxeconfig, name)
	}

	return ret
}

func findReservation(node string) string {
	//ips, err := net.LookupIP(node)

	return ""
}

// Convert an IP to a PXELinux-compatible string, i.e. 192.0.2.91 -> C000025B
func toPXE(ip net.IP) string {
	s := fmt.Sprintf("%02X%02X%02X%02X", ip[0], ip[1], ip[2], ip[3])
	return s
}