# ldb - should it be ld? (even leaner)
Lean database - it's a "no crap"-database; there is no database administration stuff, it's just an `rpc`-based server 
which can ingest any `value` and store it at a `key`, it was first meant for `gRPC`, where the encoded binary message 
is stored, and served with no touching the value on the server-side. To that end it is mostly a `gRPC`-cache, 
but I intend it to be a more general building block;
To empower the developer, for her to take her own steps in developing the data-storage-solution of her liking. 
Mostly this project is a learning project for myself, to learn `golang`, `rpc` and `gRPC`.


##API

```gotemplate
service {
    ## read
    rpc Fetch(Id) returns (YourObject) 
    rpc FetchMany(stream Id) returns (stream YourObject) 
    rpc FetchRange(IdRange) returns (stream YourObject)
    
    ## Delete 
    rpc Delete(Id) returns (YourObject) 
    rpc DeleteMany(stream Id) returns (stream YourObject) 
    rpc DeleteRange(IdRange) returns (stream YourObject)
    
    ## Create
    rpc Add(YourObject) returns (AddResponse)  # you have to query /{id} or add header X-LDB-id: {id}
    rpc AddMany(stream YourObject) return (stream AddResponse)  
    # dunno if this is even possible when the id has to be sent
}
struct Id {
    id char[]/string/whichever native works best maybe hex?   
}
struct IdRange {
    from char[]/string/whichever native works best maybe hex?   
    to char[]/string/whichever native works best maybe hex?   
    # .. maybe something better?
}

struct AddResponse{
    OK bool  # false means the ID is taken.
    # ... more? return YourObject?
}

struct YourObject {
    # ... whatever you want to package using gRPC protobuf 
}
```
- implementing upsert or InsertOrReplace functionality is up to a wrapping client-layer.
- implementing cross-server-replication (for clusters etc.) is up to a wrapping client-layer.
- implementing shard-splitting is up to a client-layer.
- implementing this, as a message-queue is up to a client (but it's easy, since delete returns the object, if one to many queuing is not that important)

