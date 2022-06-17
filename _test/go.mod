module github.com/MikkelHJuul/ld/_test

go 1.16

replace github.com/MikkelHJuul/ld/_test/client-proto/project => ./_test/client-proto/project

require (
	github.com/mmcloughlin/geohash v0.10.0
	google.golang.org/grpc v1.36.1
	google.golang.org/protobuf v1.26.0
)
