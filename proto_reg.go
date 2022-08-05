package main

import (
	"os"

	v1 "github.com/MikkelHJuul/ld/gen/ld/v1"
)

type protoReg struct {
	protoDir string
}

const suffix = ".proto"

var _ protoRegistry = (*protoReg)(nil)

// All implements protoRegistry
func (p *protoReg) All() ([]*v1.RegistryInfo, error) {
	files, err := os.ReadDir(p.protoDir)
}

// Alter implements protoRegistry
func (*protoReg) Alter(*v1.RegistryInfo) error {
	panic("unimplemented")
}

// Create implements protoRegistry
func (*protoReg) Create(*v1.RegistryInfo) error {
	panic("unimplemented")
}

// Delete implements protoRegistry
func (*protoReg) Delete(*v1.RegistryInfo) error {
	panic("unimplemented")
}

// Get implements protoRegistry
func (p *protoReg) Get(protoName string) (*v1.RegistryInfo, error) {
	content, err := os.ReadFile(p.protoDir + protoName + suffix)
	if err != nil {
		return nil, err
	}
	return &v1.RegistryInfo{ProtoName: protoName, Methods: methodsFromByte(content[0]), ProtoFile: content[1:]}, nil
}

func methodsFromByte(b uint8) []v1.RegistryInfo_Method {
	var methods []v1.RegistryInfo_Method
	for i := uint8(1); i <= uint8(255); i = i << 1 {
		if b&i == b {
			methods = append(methods, v1.RegistryInfo_Method(i))
		}
	}
	return methods
}
