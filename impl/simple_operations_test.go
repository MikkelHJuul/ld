package impl

import (
    "testing"

    pb "github.com/MikkelHJuul/ld/proto"
)

type resp struct {
    KV  *pb.KeyValue
    E   error
}

type kOp struct {
    Key *pb.Key
    Resp resp
}

type kvOp struct {
    KV  *pb.KeyValue
    Resp resp
}

type case struct {
    Set []kvOp
    Get []kOp
    Del []kOp
}

func TestSetGetDeleteSingles(t *testing.T) {
    ld := NewServer("tmp/ldsgdtest", false)
    //cases := 
}
