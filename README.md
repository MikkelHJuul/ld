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
    rpc Fetch(Id) returns (YourObjectMessage) 
    rpc FetchMany(stream Id) returns (stream YourObjectMessage) 
    rpc FetchRange(IdRange) returns (stream YourObjectMessage)
    
    ## Delete 
    rpc Delete(Id) returns (YourObjectMessage) 
    rpc DeleteMany(stream Id) returns (stream YourObjectMessage) 
    rpc DeleteRange(IdRange) returns (stream YourObjectMessage)
    
    ## Create
    rpc Insert(YourObjectMessage) returns (InsertResponse)
    rpc InsertMany(stream YourObjectMessage) return (stream InsertResponse)  
}

message Id {
    required string id = 1  # hex? At least file-compliant, TODO determine compliance
    # max length?
}

message IdRange {
    optional Id from = 1   
    optional Id to = 2  # inclusive (required for discrete systems with discrete queries)
    # .. maybe something better? regex-y? 
    # does nothing imply all?
}

message InsertResponse{
    oneof reponse {
        bool OK = 1
        YourObjectWrapper yourObject = 2 # ... returns YourObject when it fails (ID is already taken)
    }
}

message YourObjectMessage {
    required Id id = 1
    #server side:
    required bytes yourObject = 2
    #client side:
    required YourObject yourObject = 2
}

# this is not present on the server side
message YourObject {
    # ... whatever you want to package using gRPC protobuf 
}
```
- implementing upsert or InsertOrReplace functionality is up to a wrapping client-layer.
- implementing cross-server-replication (for clusters etc.) is up to a wrapping client-layer.
- implementing shard-splitting is up to a client-layer.
- implementing this, as a message-queue is up to a client (but it's easy, since delete returns the object - if one to many queuing is not that important)

## Keys
Indexing is fully on key level of you need secondary indexes you needed to maintain two version's of the data or actually create the index (id-id) table yourself.

### Implementation
I know very little about file systems and B/R-trees, and how to implement such a thing.  So I will stick to a simple folder structure.
 With configurable folder name length `SHARD_CHAR_LENGHT`
 and number of folder levels, `SHARD_LEVEL`.
This way you key could shard on year/month/day, by:
`SHARD_CHAR_LENGHT=2`, `SHARD_LEVEL=3`, placing an
a bytes object  with key `YYMMDDrestofmyID` in folder structure:
`YY/MM/DD/restofmyID`.

## Caching
The application will cache successful fetch, and - as well as failing
 insert request file inodes.
This is especially important for failing inserts where an impending `Delete` operation is waiting.
As well as for `Fetch` operations where data is typically queried multiple times or at least a `Fetch` implies importance.
Caching will probably be handled as a fixed length set of id, to inode pairs where items are popped(delete) and/or re(fetch)-inserted(insert) in the beginning of the set; the last item goes.

### Caching ideas
Should Implement/ as configuration? Cache the full shard inodes? (Combat Hot shards, fx most recent data)

