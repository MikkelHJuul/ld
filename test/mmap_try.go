package main

import (
	"fmt"
	"os"
	"reflect"
	"syscall"
	"unsafe"
)

func main() {
	file, _ := os.OpenFile("mmap.test", syscall.O_RDWR|syscall.O_APPEND, 0644)
	fd := file.Fd()
	bints, _ := mapUInt32(fd)

	//for i := range [50]int{} {
	//	bints[i] = [16]uint32{uint32(rand.Int31()),uint32(rand.Int31()),uint32(rand.Int31()),uint32(rand.Int31()),uint32(rand.Int31()),uint32(rand.Int31()),uint32(rand.Int31()),uint32(rand.Int31()),uint32(rand.Int31()),uint32(rand.Int31()),uint32(rand.Int31()),uint32(rand.Int31()),uint32(rand.Int31()),uint32(rand.Int31()),uint32(rand.Int31()),uint32(rand.Int31())}
	//}
	for j := range bints {
		fmt.Printf("%d", bints[j])
	}
	_ = syscall.Munmap(*(*[]byte)(unsafe.Pointer(&bints)))
	_ = file.Close()
}

func mapUInt32(fd uintptr) ([][16]uint32, error) {
	mmap, err := syscall.Mmap(int(fd), 0, 2048, syscall.PROT_READ|syscall.PROT_WRITE, syscall.MAP_SHARED)
	if err != nil {
		return nil, err
	}
	header := (*reflect.SliceHeader)(unsafe.Pointer(&mmap))
	// account for the size different between byte and int32
	header.Len /= 4
	header.Cap = header.Len
	return *(*[][16]uint32)(unsafe.Pointer(header)), nil
}
