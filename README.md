# ld
[![Go Report Card](https://goreportcard.com/badge/github.com/MikkelHJuul/ld)](https://goreportcard.com/report/github.com/MikkelHJuul/ld)
[![PkgGoDev](https://pkg.go.dev/badge/github.com/MikkelHJuul/ld)](https://pkg.go.dev/github.com/MikkelHJuul/ld)
[![Maintainability](https://api.codeclimate.com/v1/badges/bd9ba9fa7fdf36eb5164/maintainability)](https://codeclimate.com/github/MikkelHJuul/ld/maintainability)
[![codecov](https://codecov.io/gh/MikkelHJuul/ld/branch/master/graph/badge.svg?token=SRXDOXVAOP)](https://codecov.io/gh/MikkelHJuul/ld)
![GitHub License](https://img.shields.io/github/license/MikkelHJuul/ld)
[![FOSSA Status](https://app.fossa.com/api/projects/git%2Bgithub.com%2FMikkelHJuul%2Fld.svg?type=shield)](https://app.fossa.com/projects/git%2Bgithub.com%2FMikkelHJuul%2Fld?ref=badge_shield)

Lean database - it's a simple-database, it's just an `rpc`-based server with the basic Get/Set/Delete operations.
The database can ingest any `value` and store it at a `key`, it only supports rpc, via `gRPC`. The encoded binary message 
is stored, and served without touching the value on the server-side. To that end it is mostly a `gRPC`-cache, 
but I intend it to be a more general building block.

The database is operating on "key"-level only. If you need secondary indexes you needed to maintain two versions of the data or actually create the index (`id'`-`id`-mapping) table yourself.
Some key-value databases offer more solutions than this; this does not, and will not, offering too many solutions most often lead to poorer solutions in general.

There is no Query language to clutter your code! I know, awesome, right?!

This project started out as a learning project for myself, to learn `golang`, `rpc` and `gRPC`.

The project is written in `golang`. It will be packaged as a `scratch`-container (`linux amd64`).
I will not support other ways of downloading. 
As always you can simply `go build`

## Docker images
images are `mjuul/ld:<tag>` and (alpine)`mjuul/ld:<tag>-client`. There is also a standalone client container `mjuul/ld-client`.

The container `mjuul/ld:<tag>` is just a scratch container with the Linux/amd64 image as entrypoint.

The container `mjuul/ld:<tag>-client` is based on the image `mjuul/ld-client` adding the binary `ld` to it and running that at startup. The client serves as an interactive shell for the database, see [client](client/README.md).

## Implementation
This project exposes [badgerDB](https://github.com/dgraph-io/badger). You should be able to use the badgerDB [CLI-tool](https://github.com/dgraph-io/badger#installing-badger-command-line-tool) on the database. 

## API
Hashmap Get-Set-Delete semantics! With bidirectional streaming rpc's. No lists, because aggregation of data should be kept at a minimum. 
The APIs for get and delete further implement unidirectional server-side streams for querying via `KeyRange`.

I consider the gRPC api to be feature complete. While the underlying implementation may change to enable better database configuration and/or usage of this code as a library. Maturity may also bring changes to the server implementation.

See [test](test) for a client implementations, the testing package builds on the data from [DMI - Free data initiative](https://confluence.govcloud.dk/display/FDAPI) (specifically the lightning data set), 
but can easily be changed to ingest other data, ingestion and read separated into two different clients. 

### Working with the API
The API is expandable. Because of how gRPC encoding works you can replace the `bytes` type `value` tag on the client side with whatever you want.
This way you could use it to store dynamically typed objects using `Any`. Or you can save and query the database with a fixed or reflected type.

The test folder holds two small programs that implements a fixed type: [my_message.proto](test/client-proto/my_message.proto).

The client uses reflection to serialize/deserialize json to a message given a `.proto`-file.

### CRUD - why not CRUD?
CRUD operations must be implemented client side, use `Get -> [decision] -> Set` to implement create or update, the way you want to. fx 
```text
    Create      Get -> if {empty response} -> Set
    Update      Get/Delete -> if {non-empty} -> [map?] -> Set
```
To have done this server side would cause so much friction. All embeddable key-value databases, to my knowledge, implement Get-Set-Delete semantics, so whether you go with [bolt](https://github.com/boltdb/bolt)/[bbolt](https://github.com/etcd-io/bbolt) or badger you would always end up having this friction; so naturally you implement it without CRUD-semantics. Implementing a concurrent `GetMany`/`SetMany` ping-pong client-service feels a lot more elegant anyways.

## Configuration
via flags or environment variables:
```text
flag            ENV             default     description
------------------------------------------------------------------------------------
-port            PORT            5326        "5326" spells out "lean" in T9 keyboards
-db-location     DB_LOCATION     ld_badger   The folder location where badger stores its database-files
-in-mem          IN_MEM          false       save data in memory (or not) setting this to true ignores db-location.
-log-level       LOG_LEVEL       INFO        the logging level of the server
```
The container `mjuul/ld:<tag>-client` does not support flags for `ld`, use environment variables. (Since it is `ld-client` that is the entrypoint)


### Comparison to [ProfaneDB](https://gitlab.com/ProfaneDB/ProfaneDB)
`ProfaneDB` uses field options to find your object's key, and can ingest a list (repeated), your key can be composite, and you don't have to think about your key. (I envy the design a bit (it's shiny), but then again I don't feel like that is the best design).

`ld` forces you to design your key, and force single-object(no-aggregation/non-repeated) thinking.

`ProfaneDB` does not support any type of non-singular key queries; you will have to build query objects with very high knowledge of your keys (specific keys). This may force you to make fewer keys, and do more work in the client. (you may end up searching for a needle in a haystack, or completely loosing a key)

`ld` supports `KeyRange`s, you can then make very specific keys, and more of them, and think about the key-design, and query that via, `From`, `To`, `Prefix` and/or `Pattern` syntax.

`ProfaneDB` uses an inbuilt extension for its `.proto`. pro: you can use their `.proto` file as is. con: google's Any-type is just like map, and requires the implementer to send type-knowledge on each object on the wire.

`ld` use the underlying protocol buffers encoding design, con: this force the implementer to edit their `.proto` file, which is an anti-pattern. pro: while the database will not know anything about the value it saves, the type will be packed binary and can be serialised.

`ld` support bulk operations (via stream methods) natively. `ProfaneDB` via a repeated nested object, Memory-wise, streaming is preferred.


## License
[![FOSSA Status](https://app.fossa.com/api/projects/git%2Bgithub.com%2FMikkelHJuul%2Fld.svg?type=large)](https://app.fossa.com/projects/git%2Bgithub.com%2FMikkelHJuul%2Fld?ref=badge_large)

### TODO
- benchmarks
