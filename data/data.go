package data

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
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
	maxSize     int
	deletePoint uint
	current     uint
	data        map[string]struct {
		string
		uint
	}
}

func NewCacheService(maxSize int) *CacheService {
	return &CacheService{
		deletePoint: 0,
		current:     0,
		maxSize:     maxSize,
		data: make(map[string]struct {
			string
			uint
		}),
	}
}

func (fsc CacheService) Save(key string, bytes []byte) error {
	if _, ok := fsc.data[key]; ok {
		return &ServiceException{fmt.Sprintf("A datapoint with key: %s is already there", key)}
	}
	if fsc.maxSize <= len(fsc.data) {
		defer func() {
			for fsc.maxSize <= len(fsc.data) {
				for key, value := range fsc.data {
					if value.uint == fsc.deletePoint { // this may not delete an item, see #Get
						delete(fsc.data, key)
					}
				}
				fsc.deletePoint++
			}
		}()
	}
	fsc.addUnsafe(key, string(bytes))
	return nil
}

func (fsc CacheService) addUnsafe(key string, bytes string) {
	fsc.current++
	fsc.data[key] = struct {
		string
		uint
	}{bytes, fsc.current}
}

func (fsc CacheService) Delete(key string) ([]byte, error) {
	defer func() {
		delete(fsc.data, key)
	}()
	if data, ok := fsc.data[key]; ok {
		return []byte(data.string), nil
	}
	return nil, &ServiceException{fmt.Sprintf("No such datapoint for key: %s", key)}
}

func (fsc CacheService) Get(key string) ([]byte, error) {
	tmp := fsc.data[key].string
	defer fsc.refresh(key, tmp)
	return []byte(tmp), nil
}

func (fsc CacheService) refresh(key string, bytes string) {
	_, _ = fsc.Delete(key)
	fsc.addUnsafe(key, bytes)
}

type FileService struct {
	memCache     *CacheService
	shardCharLen int
	shardLevel   int
	rootPath     string
}

func NewFileService(shardLen int, shardLevel int, rootPath string, maxMemSize int) *FileService {
	return &FileService{
		memCache: &CacheService{
			deletePoint: 0,
			current:     0,
			maxSize:     maxMemSize,
			data: make(map[string]struct {
				string
				uint
			}),
		},
		shardCharLen: shardLen,
		shardLevel:   shardLevel,
		rootPath:     rootPath,
	}
}

func (fds FileService) determinePath(key string) string {
	keys := []string{key}
	for i := 0; i <= fds.shardLevel; i++ {
		keys = strings.SplitN(keys[i], "", fds.shardCharLen)
	}
	return fds.rootPath + "/" + strings.Join(keys, "/")
}

func (fds FileService) Delete(key string) ([]byte, error) {
	path := fds.determinePath(key)
	defer func() {
		_ = os.Remove(path)
	}()
	if item, _ := fds.memCache.Delete(key); item != nil {
		return item, nil
	}
	return ioutil.ReadFile(path)
}

func (fds FileService) Get(key string) ([]byte, error) {
	if item, _ := fds.memCache.Get(key); item != nil {
		return item, nil
	}
	path := fds.determinePath(key)
	return ioutil.ReadFile(path)
}

func (fds FileService) Save(key string, bytes []byte) error {
	if err := fds.memCache.Save(key, bytes); err != nil {
		return err
	}
	path := fds.determinePath(key)
	return ioutil.WriteFile(path, bytes, os.FileMode(0770)) // read write
}
