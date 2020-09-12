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
- implementing cross-server-replication (for clusters etc.) is up to a wrapping client-layer.
- implementing shard-splitting is up to a client-layer.
- implementing this, as a message-queue is up to a client (but it's easy, since delete returns the object - if one to many queuing is not that important)
- adding user security or tls is up to a proxy layer

##Configuration
via flags or environment variables:
```text
flag            ENV             default     description
port            PORT            5326        -
service-type    SERVICE_TYPE    FS          FS use files or MEM in-memory DB
fs-shard-level  FS_SHARD_LEVEL  3           requires service-type FS: how many levels of folders to put files in
fs-shard-len    FS_SHARD_LEN    3           requires service-type FS: length of folder names
fs-mem-size     FS_MEM_SIZE     1000        requires service-type FS: how many items to cache in memory
fs-root-path    FS_ROOT_PATH    /data       requires service-type FS: where to put data
mem-size        MEM_SIZE        100000      requires service-type MEM: how many items to hold in memory
```

## Keys
The database operate only on `key`-level if you need secondary indexes you needed to maintain two version's of the data or actually create the index (`id`-`id`-mapping) table yourself.

There is no Query language to clutter your code!

### Implementation
I know very little about file systems and B/R-trees, and how to implement such a thing.  So I will stick to a simple folder structure. (and just use the file-system B+-tree.)

With configurable folder name length `SHARD_CHAR_LEN`
 and number of folder levels, `SHARD_LEVEL`; 
This way you key could shard on year/month/day, by:
`SHARD_CHAR_LEN=2`, `SHARD_LEVEL=3`, placing
a `bytes` object  with key `YYMMDDrestofmyID` in folder structure:
`some/root/YY/MM/DD/restofmyID`.

## Caching
The application will cache successful fetch, and successful as well as failing
 insert requests.

ie.
- failing inserts with an impending `Delete` operation
- multiple-called similar `Fetch` operations
- a `Fetch` implies importance on the data-point.

Caching is done as a fixed length `map[string]struct{string, uint}` where the string is the value, and the `uint` is a scanning number to handle which item should be deleted.

