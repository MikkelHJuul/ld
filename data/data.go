package data

import (
	"fmt"
	"os"
	"regexp"
	"syscall"
	"unsafe"
)

type Service interface {
	Delete(key string) ([]byte, error)
	Get(key string) ([]byte, error)
	Save(key string, bytes []byte) error
	GetRange(streamTo func(key string, bytes []byte) error, input ...string) error
	DeleteRange(streamTo func(key string, bytes []byte) error, input ...string) error
}

type ServiceException struct {
	message string
}

func (e ServiceException) Error() string {
	return e.message
}

type CacheService struct {
	maxMemory   string
	deletePoint uint
	current     uint
	data        memMap
}

func NewCacheService(maxMemory string) *CacheService {
	return &CacheService{
		deletePoint: 0,
		current:     0,
		maxMemory:     maxMemory,
		data: make(map[string]struct {
			*string
			uint
		}),
	}
}

func (cs CacheService) Save(key string, bytes []byte) error {
	if _, ok := cs.data[key]; ok {
		return &ServiceException{fmt.Sprintf("A datapoint with key: %s is already there", key)}
	}
	defer func() {
		for cs.maxMemory <= len(cs.data) {
			for key, value := range cs.data {
				if value.uint == cs.deletePoint { // this may not delete an item, see #Get
					delete(cs.data, key)
				}
			}
			cs.deletePoint++
		}
	}()
	stringBytes := string(bytes)
	cs.addUnsafe(key, &stringBytes)
	return nil
}

func (cs CacheService) addUnsafe(key string, bytes *string) {
	cs.current++
	cs.data[key] = struct {
		*string
		uint
	}{bytes, cs.current}
}

func (cs CacheService) Delete(key string) ([]byte, error) {
	defer func() {
		delete(cs.data, key)
	}()
	if data, ok := cs.data[key]; ok {
		return []byte(*data.string), nil
	}
	return nil, &ServiceException{fmt.Sprintf("No such datapoint for key: %s", key)}
}

func (cs CacheService) Get(key string) ([]byte, error) {
	tmp := cs.data[key].string
	defer cs.refresh(key, tmp)
	return []byte(*tmp), nil
}

func (cs CacheService) refresh(key string, bytes *string) {
	_, _ = cs.Delete(key)
	cs.addUnsafe(key, bytes)
}

func (cs CacheService) handleRange(streamMethod func(key string, bytes []byte) error, input []string) error {
	var matchMethod func(key string, value struct {
		*string
		uint
	}) error
	if len(input) == 1 {
		// assume regexp
		matcher, err := regexp.Compile(input[0])
		if err != nil {
			return err
		}
		matchMethod = func(key string, value struct {
			*string
			uint
		}) error { if matcher.MatchString(key) {
			return streamMethod(key, []byte(*value.string))
		}; return nil }
	} else if len(input) == 2 {
		from := input[0]
		to := input[1]
		matchMethod = func(key string, value struct {
			*string
			uint
		}) error { if key >= from && key <= to {
			return streamMethod(key, []byte(*value.string))
		}; return nil }
	}
	for key, value := range cs.data {
		if err := matchMethod(key, value); err != nil { // https://github.com/golangci/golangci-lint/issues/510 ?? tror den, at cs.data er en nil-slice
			return err
		}
	}
	return nil
}

func (cs CacheService) DeleteRange(streamTo func(key string, bytes []byte) error, input ...string) error {
	return cs.handleRange(func(key string, bytes []byte) error {
		defer delete(cs.data, key)
		return streamTo(key, bytes)
	}, input)
}

func (cs CacheService) GetRange(streamTo func(key string, bytes []byte) error, input ...string) error {
	return cs.handleRange(streamTo, input)
}


func NewMMapService(dataFile *os.File) *CacheService {
	return &CacheService{
		maxMemory:   "?",
		deletePoint: 0,
		current:     0,
		data:        MMap(dataFile),
	}
}

type memMap map[string]struct{string; uint}

func MMap(f *os.File) (*memMap, error) {
	var fd = f.Fd()
	fi, err := f.Stat()
	if err != nil {
		return nil, err
	}

	mmap, err := syscall.Mmap(int(fd),0, int(fi.Size()),syscall.PROT_WRITE|syscall.PROT_READ, syscall.MAP_SHARED)

	memMap := (*memMap)(unsafe.Pointer(&mmap[0]))

	return memMap, nil
}

