# ld
Lean database - it's a "no crap"-database; there is no database administration stuff, it's just an `rpc`-based server 
which can ingest any `value` and store it at a `key`, it was first meant for `gRPC`, where the encoded binary message 
is stored, and served with no touching the value on the server-side. To that end it is mostly a `gRPC`-cache, 
but I intend it to be a more general building block;
To empower the developer, for her to take her own steps in developing the data-storage-solution of her liking. 
Mostly this project is a learning project for myself, to learn `golang`, `rpc` and `gRPC`.

The project will probably be `golang`. It will be packaged as a `scratch`-container (linux amd64). I expect the image size to be about 15-30MB max
I will not support other ways of downloading.  (Just extract it from the docker image if you want it running in VM or bare metal,  but you shouldn't).

## API

```proto
service ld {
    ## read
    rpc Fetch(StringValue) returns (Any) 
    rpc FetchMany(stream StringValue) returns (stream Any) 
    rpc FetchRange(IdRange) returns (stream Any)
    
    ## Delete 
    rpc Delete(StringValue) returns (Any) 
    rpc DeleteMany(stream StringValue) returns (stream Any) 
    rpc DeleteRange(IdRange) returns (stream Any)
    
    ## Create
    rpc Insert(InsertObject) returns (InsertResponse)
    rpc InsertMany(stream InsertObject) return (stream InsertResponse)  
}

message IdRange {
    optional string pattern = 1  # unix style POSIX regex
    optional string from = 2
    optional string to = 3  # inclusive (required for discrete systems with discrete queries)
}

message InsertResponse {
    bool OK = 1  # false implies ID-clash
}

message InsertObject {
    required string id = 1  # [(validate.rules).string { pattern: "(?i)^[0-9a-zA-Z_-.~]+$", max_len: 64 }];  # https://tools.ietf.org/html/rfc3986#section-2.3
    required Any value = 2
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

## Keys
Indexing is fully on key level of you need secondary indexes you needed to maintain two version's of the data or actually create the index (id-id) table yourself.

### Implementation
I know very little about file systems and B/R-trees, and how to implement such a thing.  So I will stick to a simple folder structure. (and just use the file-system B+-tree.)

 With configurable folder name length `SHARD_CHAR_LENGHT`
 and number of folder levels, `SHARD_LEVEL`; 
This way you key could shard on year/month/day, by:
`SHARD_CHAR_LENGHT=2`, `SHARD_LEVEL=3`, placing
a `bytes` object  with key `YYMMDDrestofmyID` in folder structure:
`YY/MM/DD/restofmyID`.

## Caching
The application will cache successful fetch, and successful as well as failing
 insert request file inodes.

ie.
- failing inserts with an impending `Delete` operation
- multiply called similar `Fetch` operations
- a `Fetch` implies importance on the data-point.

Caching will probably be handled as a fixed length set of id, to inode pairs where items are popped(delete) and/or (re)inserted(fetch/insert) in the beginning of the set; the last item goes.

### Caching ideas
Should Implement/as configuration? Cache the full shard inodes? (Combat Hot shards, fx most recent data)

TODO is Caching only needed because I don't properly implement a B-tree, complicating this project,  when I should be focusing on the data... the API is the boss! Implementation may, and will change!
