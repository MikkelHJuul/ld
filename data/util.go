package data

import (
	"os"
	"strconv"
	"syscall"
)

func ReadAt(fileInt uintptr, offset int64, len int) ([]byte, error) {
	bytes, err := syscall.Mmap(int(fileInt), offset, len, syscall.PROT_READ, syscall.MAP_PRIVATE)
	memBytes := bytes
	defer syscall.Munmap(bytes)
	return memBytes, err
}

func mapFull(file *os.File) ([]byte, error) {
	fd := file.Fd()
	if stat, err := file.Stat(); err != nil {
		return []byte(""), err
	} else {
		length := stat.Size()
		return syscall.Mmap(int(fd), 0, int(length), syscall.PROT_READ, syscall.MAP_PRIVATE)
	}
}

func WriteEnd(file *os.File, data []byte) (int64, int, error) {
	if stat, err := file.Stat(); err != nil {
		return 0, 0, err
	} else {
		offset := stat.Size()
		_, err = file.WriteAt(data, offset)
		return offset, len(data), nil
	}
}

func HexToUint8(hex string) uint8 {
	var err error
	if i, err := strconv.ParseUint(hex, 16, 64); err == nil {
		return uint8(i)
	}
	panic(err)
}
