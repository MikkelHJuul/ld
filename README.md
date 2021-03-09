# ld
Lean database - it's a simple-database, it's just an `rpc`-based server with basic CRUD.
which can ingest any `value` and store it at a `key`, it only support rpc, via `gRPC`. The encoded binary message 
is stored, and served without touching the value on the server-side. To that end it is mostly a `gRPC`-cache, 
but I intend it to be a more general building block;

The database operate only on `key`-level if you need secondary indexes you needed to maintain two version's of the data or actually create the index (`id'`-`id`-mapping) table yourself.

There is no Query language to clutter your code! I know, awesome, right?!

This project started out as a learning project for myself, to learn `golang`, `rpc` and `gRPC`.

The project is written in `golang`. It will be packaged as a `scratch`-container (`linux amd64`).
I will not support other ways of downloading. 
As always you can simply `go build`

## Implementation
This project exposes badgerDB. You should be able to use the badgerDB CLI-tools on the database. 

## API
CRUD! With bidirectional streaming rpc's. No lists, because aggregation of data should be kept at a minimum.
Update is implemented as upsert.
The APIs for read and delete further implement unidirectional server-side streams for querying via `KeyRange`.

see [test](test) for client implementations

##Configuration
via flags or environment variables:
```text
flag            ENV             default     description
------------------------------------------------------------------------------------
port            PORT            5326        "5326" spells out "lean" in T9 keyboards
in-mem          IN_MEM          false       save data in memory (or not)
```

### Working with the API
The API is expandable. Because of how gRPC encoding works you can replace the `bytes` type `value` tag on the client side with whatever you want.
This way you could use it to store dynamically typed objects using `Any`. Or you can save and query the database with a fixed type.

## Testing keys
I have exposed the primary mechanism for matching the key via the structure `data.KeyRangeWrapper.Match(string)`
