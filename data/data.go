package data

import (
	"encoding/hex"
	"fmt"
	pb "github.com/MikkelHJuul/ld/service"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
)

type Service interface {
	Delete(chan pb.KeyValue, string) ([]byte, error)
	Get(chan pb.KeyValue, string) ([]byte, error)
	Save(chan bool, string, []byte) error
	GetRange(chan pb.KeyValue, ...string) error
	DeleteRange(chan pb.KeyValue, ...string) error
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
	data        map[string]struct {
		*string
		uint
	}
	mutex sync.RWMutex
}

func NewCacheService(maxMemory string) *CacheService {
	return &CacheService{
		deletePoint: 0,
		current:     0,
		maxMemory:   maxMemory,
		data: make(map[string]struct {
			*string
			uint
		}),
	}
}

func (cs *CacheService) Save(key string, bytes []byte) error {
	if _, ok := cs.data[key]; ok {
		return &ServiceException{fmt.Sprintf("A datapoint with key: %s is already there", key)}
	}
	defer func() {
		for len(cs.maxMemory) <= len(cs.data) {
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

func (cs *CacheService) addUnsafe(key string, bytes *string) {
	cs.current++
	cs.data[key] = struct {
		*string
		uint
	}{bytes, cs.current}
}

func (cs *CacheService) Delete(key string) ([]byte, error) {
	if data, ok := cs.data[key]; ok {
		defer func() {
			delete(cs.data, key)
		}()
		return []byte(*data.string), nil
	}
	return nil, &ServiceException{fmt.Sprintf("No such datapoint for key: %s", key)}
}

func (cs *CacheService) Get(key string) ([]byte, error) {
	tmp := cs.data[key].string
	defer cs.refresh(key, tmp)
	return []byte(*tmp), nil
}

func (cs *CacheService) refresh(key string, bytes *string) {
	_, _ = cs.Delete(key)
	cs.addUnsafe(key, bytes)
}

func (cs *CacheService) handleRange(streamMethod func(key string, bytes []byte) error, input []string) error {
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
		}) error {
			if matcher.MatchString(key) {
				return streamMethod(key, []byte(*value.string))
			}
			return nil
		}
	} else if len(input) == 2 {
		from := input[0]
		to := input[1]
		matchMethod = func(key string, value struct {
			*string
			uint
		}) error {
			if key >= from && key <= to {
				return streamMethod(key, []byte(*value.string))
			}
			return nil
		}
	}
	for key, value := range cs.data {
		if err := matchMethod(key, value); err != nil { // https://github.com/golangci/golangci-lint/issues/510 ?? tror den, at cs.data er en nil-slice
			return err
		}
	}
	return nil
}

func (cs *CacheService) DeleteRange(streamTo func(key string, bytes []byte) error, input ...string) error {
	return cs.handleRange(func(key string, bytes []byte) error {
		defer delete(cs.data, key)
		return streamTo(key, bytes)
	}, input)
}

func (cs *CacheService) GetRange(streamTo func(key string, bytes []byte) error, input ...string) error {
	return cs.handleRange(streamTo, input)
}

type VMemService struct {
	keyFile [][16]int32
	ptrFile []struct {
		uint16
		int64
	}
	dataFile *os.File
	mutex    sync.RWMutex
	holes    []struct {
		uint16
		int32
	}
	chunkSize int16
}

const kvSep string = "\x1e"
const end string = "\x19"

func NewMMapService(dataFolder *string) *VMemService {
	return &VMemService{
		data: NewDb(dataFolder),
	}
}

func (vs *VMemService) Get(ch chan pb.KeyValue, key string) {
	vs.mutex.RLock()
	var ptr = vs.getPtr(key)
	var dataLenOff = vs.ptrFile[ptr]
	out, err := ReadAt(vs.dataFile.Fd(), dataLenOff.int64, int(dataLenOff.uint16))
	vs.mutex.RUnlock()
	if err != nil {
		return
	}
	kv := strings.Split(string(out), kvSep)
	ch <- pb.KeyValue{
		Key:   &pb.Key{Key: kv[0]},
		Value: []byte(strings.Split(kv[1], end)[0]),
	}
}

func (vs *VMemService) Save(ch chan bool, key string, bytes []byte) error {
	cha := make(chan string)
	go vs.get(cha, key)
	if <-cha != "" { //verify
		ch <- false
		return &ServiceException{"already there"}
	}
	ch <- vs.add(key, bytes)
	return nil
}

func (vs *VMemService) Delete(key string) ([]byte, error) {
	if data, ok := (*vs.data)[key]; ok {
		defer func() {
			delete(*vs.data, key)
		}()
		return data, nil
	}
	return nil, &ServiceException{fmt.Sprintf("No such datapoint for key: %s", key)}
}

func (vs *VMemService) handleRange(streamMethod func(key string, bytes []byte) error, input []string) error {
	var matchMethod func(key string, value []byte) error
	if len(input) == 1 {
		// assume regexp
		matcher, err := regexp.Compile(input[0])
		if err != nil {
			return err
		}
		matchMethod = func(key string, value []byte) error {
			if matcher.MatchString(key) {
				return streamMethod(key, value)
			}
			return nil
		}
	} else if len(input) == 2 {
		from := input[0]
		to := input[1]
		matchMethod = func(key string, value []byte) error {
			if key >= from && key <= to {
				return streamMethod(key, value)
			}
			return nil
		}
	}
	for key, value := range *vs.data {
		if err := matchMethod(key, value); err != nil { // https://github.com/golangci/golangci-lint/issues/510 ?? tror den, at vs.data er en nil-slice
			return err
		}
	}
	return nil
}

func (vs *VMemService) DeleteRange(streamTo func(key string, bytes []byte) error, input ...string) error {
	return vs.handleRange(func(key string, bytes []byte) error {
		defer delete(*vs.data, key)
		return streamTo(key, bytes)
	}, input)
}

func (vs *VMemService) GetRange(streamTo func(key string, bytes []byte) error, input ...string) error {
	return vs.handleRange(streamTo, input)
}

func (vs *VMemService) add(key string, bytes []byte) bool {
	defer vs.mutex.Unlock()
	toWrite := key + kvSep + string(bytes) + end
	size := len(toWrite)%int(vs.chunkSize) + 1
	var writeOffset int32 = 0
	for item := range vs.holes {
		if int(vs.holes[item].uint16) < size {
			writeOffset = vs.holes[item].int32
			go vs.cleanUpHoles(item)
			break
		}
	}
	if writeOffset == 0 {
		stat, _ := vs.dataFile.Stat()
		writeOffset = int32(stat.Size()) // +1?
	}
	vs.mutex.Lock()
	_, _ = vs.dataFile.WriteAt([]byte(toWrite), int64(writeOffset))
	var ptrFilePtr = len(vs.ptrFile)
	vs.ptrFile[ptrFilePtr] = struct {
		uint16
		int64
	}{uint16(size), int64(writeOffset)}
	return vs.insertIntoKeyFile(key, int32(ptrFilePtr))
}

func (vs *VMemService) insertIntoKeyFile(key string, ptr int32) bool {
	var actualPtr = -ptr
	hxKey := hex.EncodeToString([]byte(key))
	sliceNum, ptrPtr, hxPos := vs.walk(hxKey, 0, 0)
	if len(hxKey) == hxPos && ptrPtr > 0 {
		vs.keyFile[ptrPtr][0] = actualPtr // full keys are stored here ( at uneven hex-buckets)
		return true
	} else if len(hxKey) == hxPos && ptrPtr < 0 {
		// have to reconstruct this bucket because an already existing value is placed here, with a longer key than this
		newPtr := len(vs.keyFile)
		otherPointer := vs.keyFile[sliceNum][HexToUint8(string(hxKey[hxPos]))]
		filePtr := vs.ptrFile[-otherPointer]
		var nextChar string
		if output, err := ReadAt(vs.dataFile.Fd(), filePtr.int64, int(filePtr.uint16)); err == nil {
			nextChar = strings.SplitAfter(string(output), kvSep)[0]
		}
		nextHex := hex.EncodeToString([]byte(nextChar))
		vs.keyFile[sliceNum][HexToUint8(string(hxKey[hxPos]))] = int32(newPtr)
		//build new bucket
		vs.keyFile[newPtr][HexToUint8(string(nextHex[0]))] = otherPointer
		vs.keyFile[newPtr][0] = actualPtr // full keys are stored here ( at uneven hex-buckets)
		return true
	} else if len(hxKey) == hxPos && ptrPtr == 0 {
		// we have to make a new bucket (or we would have to make a new bucket the next time...
		// and I don't want to have to find out that this should be placed at [0] In the next bucket)
		newPtr := len(vs.keyFile)
		vs.keyFile[sliceNum][HexToUint8(string(hxKey[hxPos]))] = int32(newPtr)
		vs.keyFile[newPtr][0] = actualPtr // full keys are stored at hex 0 in the uneven hex-buckets
		return true
	} else if len(hxKey) > hxPos && ptrPtr == 0 {
		vs.keyFile[sliceNum][HexToUint8(string(hxKey[hxPos]))] = actualPtr
		// this is the last found bucket, the given hxKey points nowhere, use this buckets empty value
		return true
	}
	fmt.Errorf("%s", "WTF this shouldn't happen")
	return false
}

func (vs *VMemService) walk(hxKey string, b int, sliceNum int32) (int32, int32, int) { // parallelize?
	var pos int64 = 0
	if len(hxKey) >= b {
		pos, _ = strconv.ParseInt(string(hxKey[b]), 16, 64)
	}
	var ptr = vs.keyFile[sliceNum][pos]
	if ptr <= 0 {
		return sliceNum, ptr, b
	}
	b++
	return vs.walk(hxKey, b, ptr)
}

func (vs *VMemService) getPtr(key string) int32 {
	_, ptr, _ := vs.walk(hex.EncodeToString([]byte(key)), 0, 0)
	if ptr > 0 {
		return 0
	}
	return -ptr
}

// clean up files...? holes etc.
