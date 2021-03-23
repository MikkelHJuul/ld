# ld
Lean database - it's a simple-database, it's just an `rpc`-based server with basic Get/Set/Delete operations.
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

## Implementation
This project exposes [badgerDB](https://github.com/dgraph-io/badger). You should be able to use the badgerDB CLI-tools on the database. 

## API
Hashmap Get-Set-Delete semantics! With bidirectional streaming rpc's. No lists, because aggregation of data should be kept at a minimum.
The APIs for get and delete further implement unidirectional server-side streams for querying via `KeyRange`.

See [test](test) for client implementations, the testing package builds on the data from [DMI - Free data initiative](https://confluence.govcloud.dk/display/FDAPI) (specifically the lightning data set), 
but can easily be changed to ingest other data, ingestion and read separated into two different clients. 

(Note, loading the 9 mil datapoints for [lightning](https://confluence.govcloud.dk/pages/viewpage.action?pageId=37355752) from a downloaded unzipped line-delimited json-file takes about 2.5 mins)

### CRUD
CRUD operations must be implemented client side, use `Get -> [decision] -> Set` to implement create or update, the way you want to. fx 
```text
    Create      Get -> if {empty response} -> Set
    Update      Get -> if {non-empty} -> [map?] -> Set
```
To have done this server side would cause so much friction. All embeddable key-value databases, to my knowledge, implement Get-Set-Delete semantics, so whether you go with [bolt](https://github.com/boltdb/bolt)/[bbolt](https://github.com/etcd-io/bbolt) or badger you would always end up having this friction; so naturally you implement it without CRUD-semantics. Implementing a concurrent `GetMany`/`SetMany` ping-pong client-service feels a lot more elegant anyways.

## Configuration
via flags or environment variables:
```text
flag            ENV             default     description
------------------------------------------------------------------------------------
-port            PORT            5326        "5326" spells out "lean" in T9 keyboards
-db-location     DB_LOCATION     ld_badger   The folder location where badger stores its database-files
-in-mem          IN_MEM          false       save data in memory (or not)
-log-level       LOG_LEVEL       INFO        the logging level of the server
```

### Working with the API
The API is expandable. Because of how gRPC encoding works you can replace the `bytes` type `value` tag on the client side with whatever you want.
This way you could use it to store dynamically typed objects using `Any`. Or you can save and query the database with a fixed type.

### Comparison to [ProfaneDB](https://gitlab.com/ProfaneDB/ProfaneDB)
`ProfaneDB` uses field options to find your object's key, and can ingest a list (repeated), your key can be composite, and you don't have to think about your key. (I envy the design a bit (it's shiny), but then again I don't feel like that is the best design).

`ld` forces you to design your key, and force single-object(no-aggregation/non-repeated) thinking.

`ProfaneDB` does not support any type of non-singular key queries; you will have to build query objects with very high knowledge of your keys (specific keys). This may force you to make fewer keys, and do more work in the client. (you may end up searching for a needle in a haystack, or completely loosing a key)

`ld` supports `KeyRange`s, you can then make very specific keys, and more of them, and think about the key-design, and query that via, `From`, `To`, `Prefix` and/or `Pattern` syntax.

`ProfaneDB` uses an inbuilt extension for its `.proto`. pro: you can use their `.proto` file as is. con: google's Any-type is just like map, and requires the implementer to send type-knowledge on each object on the wire.

`ld` use the underlying protocol buffers encoding design, con: this force the implementer to edit their `.proto` file, which is an anti-pattern. pro: while the database will not know anything about the value it saves, the type will be packed binary and can be serialised.

`ld` support bulk operations (via stream methods) natively. `ProfaneDB` via a repeated nested object, Memory-wise, streaming is preferred.
