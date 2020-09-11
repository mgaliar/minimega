# User Apps

This is the documentation for phenix's user apps.

## Default User Apps

- `app.go`: configures and starts user apps.
- `ntp.go`: configures a NTP server into the experiment infrastructure.
- `option.go`: ??? options in user apps ???
- `serial.go`: configures a Serial interface on a VM image.
- `startup.go`: configures minimega startup injections based on OS type.
- `user.go`: used to shell out with JSON payload to customer user apps.
- `vyatta.go`: the Vyatta user app is used to customize a vyatta router image. 
  This includes setting interfaces, ACL rules, IPSec VPN settings, etc.

!!! todo
    Add something more 

    `v1` of experiment data -- includes all experiment data in a single JSON or 
    YAML configuration file
    
    `v2` of experiment data -- includes topology, experiment, and scenario YAML
    configuation files

## Custom User Apps

Customer user apps are interacted with through `stdin` and `stdout`. The phenix
`user.app` will pass a JSON package through `stdin`. This JSON package will 
contain the experiment data based on the schema published at **{{FIXME}}**. The
phenix will block further actions and wait for `stdout` return of a JSON package
from your custom user app. In addition to the JSON package, you should return
an exit code of `0`. If you are returning log(s) or any error messages, those
should be passed via a different **{{WHAT_IS_BEST_TERM?}}**.

!!! todo
    Examples; maybe the simple test app?

Example:

```
import json, sys


def eprint(*args):
    print(*args, file=sys.stderr)


def main() :
    if len(sys.argv) != 2:
        eprint("must pass exactly one argument on the command line")
        sys.exit(1)

    spec = json.loads(sys.stdin.read())

    for n in spec['topology']['nodes']:
        for d in n['hardware']['drives']:
            d['image'] = 'm$.qc2'

    print(json.dumps(spec))
```