# Test ingest and fetcher client

Two simple clients for the database

I used the DMI dataset, files (zipped) can be downloaded at: `https://dmigw.govcloud.dk/v2/lightningdata/bulk/?api-key=<your-key>`.

The `api-key` requires sign up. I added the data for april 2005 in this repository, and some range-queries for it.

sign up via this guide: https://confluence.govcloud.dk/pages/viewpage.action?pageId=26476690

## How I made this
*so you could make one yourself*
1. I downloaded the data
2. built the `.proto`-file via https://json-to-proto.github.io/.
3. fixed some names and types (float >< integer)
4. I added the new type (named `Feature`) as the value-type in ld.proto, as [ld.proto](client-proto/ld.proto). 
5. I compiled the client library using this tempered `.proto`-file
   ```shell
   cd ./client-proto
   docker run -v `pwd`:/defs namely/protoc-all -l go -f ld.proto -o .
   ```
6. I built the ingestion algorithm, serializing this json-file (line-delimited) to proto-messages, with a focus on building a [key](client-proto/key.go), 
7. I added some `get` stuff to test querying.
