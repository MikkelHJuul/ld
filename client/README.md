# ld-client
Client for interacting with `ld`. This client is an executable and interactive shell.

The client is configured as an entrypoint for `ld` in the main-method in order to be able 
to use it as a shell-compliant container via: `mjuul/ld:<tag>-client`, which is the 
container `mjuul/ld-client` including the `ld` binary.


## Usage

```shell
>> /ld-client help

ld-client is an interactive client and executable to do non-"client-side" streaming requests

Usage:
  ld-client [command]

Commands:
  delete, del, remove, rem  get a single record
  delete-range, delran      delete a range of records, empty implies all
  get, fetch, read          get a single record
  get-range, getran         get a range of records, empty implies all
  help                      use 'help [command]' for command help
  set, add, create          set a single record
  version                   print version info

Flags:
  -h, --help                display help
      --nocolor             disable color output
  -p, --protofile string    the protofile to serialize from, if unset plain bytes are sent, 
                            and the received values are not marshalled to JSON
  -t, --target    string    the target ld server (default: localhost:5326)
```

The 5 methods (not `version` and `help`) correspond to the non-"client-side" streaming requests of [`ld.proto`](../proto/ld.proto).

An example for starting the container with docker would be: 
```shell
❯ docker run -it -v /my/proto/location/:/proto/ mjuul/ld-client -p /proto/my-proto -t some.ld.online
```
this interactive client would be able to serialise json to your `proto.Message` and then to the wire-bytes, sending it to the database, and displaying it on the way out.


### TODO
- tests
- better protofile handling
- "many"-endpoints (std-in-streaming '-'? via multiple input, via file?)