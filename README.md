# ld
Lean database - it's a "no crap"-database; there is no database administration stuff, it's just an `rpc`-based server 
which can ingest any `value` and store it at a `key`, it only support rpc, via `gRPC`. The encoded binary message 
is stored, and served without touching the value on the server-side. To that end it is mostly a `gRPC`-cache, 
but I intend it to be a more general building block;
To empower the developer, for her to take her own steps in developing the data-storage-solution of her liking. 

Mostly this project is a learning project for myself, to learn `golang`, `rpc` and `gRPC`.

The project is written in `golang`. It will be packaged as a `scratch`-container (`linux amd64`). atm. it is an `11 MB` scratch container. 
I will not support other ways of downloading.  (Just extract it from the docker image if you want it running in VM or bare metal, why would you?).

## API

```proto
syntax = "proto3";

option go_package = "github.com/MikkelHJuul/ld/service";

package service;

service ld {
    //// read
    rpc Fetch(Key) returns (KeyValue);
    rpc FetchMany(stream Key) returns (stream KeyValue);
    rpc FetchRange(KeyRange) returns (stream KeyValue);

    //// Delete
    rpc Delete(Key) returns (KeyValue);
    rpc DeleteMany(stream Key) returns (stream KeyValue);
    rpc DeleteRange(KeyRange) returns (stream KeyValue);

    //// Create
    rpc Insert(KeyValue) returns (InsertResponse);
    rpc InsertMany(stream KeyValue) returns (stream InsertResponse);
}

message KeyRange {
    string pattern = 1;  // unix style POSIX regex or RE2 tbd https://github.com/google/re2/wiki/Syntax
    string from = 2;
    string to = 3;  // inclusive (required for discrete systems with discrete queries)
}

message InsertResponse {
    bool OK = 1;  // false implies ID-clash
}

message Key {
    string key = 1;  // [(validate.rules).string { pattern: "(?i)^[0-9a-zA-Z_-.~]+$", max_len: 64 }];  // https://tools.ietf.org/html/rfc3986//section-2.3
}

message KeyValue {
    Key key = 1;
    bytes value = 2;
}

# this is not present on the server side
#message YourObject {
#     ... whatever you want to package using gRPC protobuf 
#}
```
- implementing upsert or InsertOrReplace functionality is up to a wrapping client-layer.
- implementing cross-server-replication (for clusters etc.) is up to a wrapping client-server-layer.
- implementing shard-splitting is up to a client-server-layer.
- implementing this, as a message-queue is up to a client (but it's easy, since delete returns the object - if one to many queuing is not that important and transactions and replication etc.)
- adding user security or tls is up to a proxy layer


see [test](test) for client implementations

##Configuration
via flags or environment variables:
```text
flag            ENV             default     description
------------------------------------------------------------------------------------
port            PORT            5326        "5326" spells out "lean" in T9 keyboards
service-type    SERVICE_TYPE    FS          MMAP use virtual memory or MEM in-memory DB
mmap-file       MMAP_FILE       /data       requires service-type MMAP: where to put data, file or folder
mem-size        MEM_SIZE        5G          requires service-type MEM: how much memory to occupy for data in memory
```

## Keys
The database operate only on `key`-level if you need secondary indexes you needed to maintain two version's of the data or actually create the index (`id'`-`id`-mapping) table yourself.

There is no Query language to clutter your code! I know, awesome, right?!

## Inmemory reinsertion
The application will reinsert successful fetch data (otherwise it may be removed because of cache clean up).

## Virtual memory
I make use of `syscall.Mmap(...)`
