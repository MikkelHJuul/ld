package data

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)
// TODO consider design of deference, possibly use goroutines? whichever is fastest and less error prone
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
		}, maxSize),
	}
}

func (cs CacheService) Save(key string, bytes []byte) error {
	if _, ok := cs.data[key]; ok {
		return &ServiceException{fmt.Sprintf("A datapoint with key: %s is already there", key)}
	}
	if cs.maxSize <= len(cs.data) {
		defer func() {
			for cs.maxSize <= len(cs.data) {
				for key, value := range cs.data {
					if value.uint == cs.deletePoint { // this may not delete an item, see #Get
						delete(cs.data, key)
					}
				}
				cs.deletePoint++
			}
		}()
	}
	cs.addUnsafe(key, string(bytes))
	return nil
}

func (cs CacheService) addUnsafe(key string, bytes string) {
	cs.current++
	cs.data[key] = struct {
		string
		uint
	}{bytes, cs.current}
}

func (cs CacheService) Delete(key string) ([]byte, error) {
	defer func() {
		delete(cs.data, key)
	}()
	if data, ok := cs.data[key]; ok {
		return []byte(data.string), nil
	}
	return nil, &ServiceException{fmt.Sprintf("No such datapoint for key: %s", key)}
}

func (cs CacheService) Get(key string) ([]byte, error) {
	tmp := cs.data[key].string
	defer cs.refresh(key, tmp)
	return []byte(tmp), nil
}

func (cs CacheService) refresh(key string, bytes string) {
	_, _ = cs.Delete(key)
	cs.addUnsafe(key, bytes)
}

func (cs CacheService) handleRange(streamMethod func(key string, bytes []byte) error, input []string) error {
	var matchMethod func(key string, value struct{string;uint}) error
	if len(input) == 1 {
		// assume regexp
		matcher, err := regexp.Compile(input[0])
		if err != nil {
			return err
		}
		matchMethod = func(key string, value struct{string;uint}) error {
			if matcher.MatchString(key) {
				return streamMethod(key, []byte(value.string))
			}
			return nil
		}
	} else if len(input) == 2 {
		from := input[0]
		to := input[1]
		matchMethod =  func(key string, value struct{string;uint}) error {
			if key >= from && key <= to {
				return streamMethod(key, []byte(value.string))
			}
			return nil
		}
	}
	for key,value := range cs.data {
		if err := matchMethod(key, value); err != nil {  // https://github.com/golangci/golangci-lint/issues/510 ?? tror den, at cs.data er en nil-slice
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

func (fs FileService) determinePath(key string) (string, string) {
	var keys []string
	for i := 0; i < fs.shardLevel; i++ {
		keyslice := []byte(key)
		keys = append(keys, string(keyslice[0:fs.shardCharLen]))
		key = string(keyslice[fs.shardCharLen:])
	}
	folderPath := fs.rootPath + "/" + strings.Join(keys, "/")
	return folderPath, folderPath + "/" + key
}

func (fs FileService) Delete(key string) ([]byte, error) {
	_, path := fs.determinePath(key)
	defer func() {
		_ = os.Remove(path)
	}()
	if item, _ := fs.memCache.Delete(key); item != nil {
		return item, nil
	}
	return ioutil.ReadFile(path)
}

func (fs FileService) Get(key string) ([]byte, error) {
	if item, _ := fs.memCache.Get(key); item != nil {
		return item, nil
	}
	_, path := fs.determinePath(key)
	return ioutil.ReadFile(path)
}

func (fs FileService) Save(key string, bytes []byte) error {
	if err := fs.memCache.Save(key, bytes); err != nil {
		return err
	}
	folderPath, path := fs.determinePath(key)
	err := os.MkdirAll(folderPath, os.FileMode(0770))
	if err != nil {
		return err
	}
	return ioutil.WriteFile(path, bytes, os.FileMode(0770)) // read write
}

func (fs FileService) handleRange(streamMethod func(path string) error, input []string) error {
//TODO test if skipDir works properly also partial compile of matcher, for faster exclusion??
	doStreamFile := func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		} else {
			return streamMethod(path)
		}
	}
	var matchMethod func(path string) bool
	var excludeCheck func(path string) bool
	if len(input) == 1 {
		// assume regexp
		matcher, err := regexp.Compile(input[0])
		if err != nil {
			return err
		}
		matchMethod = func(path string) bool {
			return matcher.MatchString(fs.extrapolateKey(path))
		}
		excludeCheck = func(path string) bool {
			//cannot determine (people can use very complex regex's...)
			return false
		}
	} else if len(input) == 2 {
		_, from := fs.determinePath(input[0])
		_, to := fs.determinePath(input[1])
		matchMethod = func(path string) bool {
			return path >= from && path <= to
		}
		excludeCheck = func(path string) bool {
			shardTot := fs.shardCharLen * fs.shardLevel
			if len(from) <= shardTot && len(from) <= shardTot {
				return true
			}
			if path >= string([]byte(from)[0:shardTot]) && path <= string([]byte(to)[0:shardTot]) {
				// there are no items in here that matches
				return true
			}
			return false
		}
	}
	walkMethod := func(path string, f os.FileInfo, err error) error {
		if f.IsDir() {
			if matchMethod(path) {
				err := filepath.Walk(path, doStreamFile); if err != nil {
					return err
				}
				return filepath.SkipDir
			}
			if excludeCheck(path) {
				return filepath.SkipDir
			}
		}
		if matchMethod(path) {
			return streamMethod(path)
		}
		return nil
	}
	return filepath.Walk(fs.rootPath, walkMethod)
}

func (fs FileService) DeleteRange(streamTo func(key string, bytes []byte) error, input ...string) error {
	var returnMethod = func (path string) error {
		if resp, err := ioutil.ReadFile(path); err != nil {
			return err
		} else {
			defer func() {
				go func() {
					if err := os.Remove(path); err != nil {
						println("Error when removing file: %s", path)
					}
				}()
			}()
			return streamTo(fs.extrapolateKey(path), resp)
		}
	}
	return fs.handleRange(returnMethod, input)
}

func (fs FileService) GetRange(streamTo func(key string, bytes []byte) error, input ...string) error {
	var returnMethod = func (path string) error {
		if resp, err := ioutil.ReadFile(path); err != nil {
			return err
		} else {
			return streamTo(fs.extrapolateKey(path), resp)
		}
	}
	return fs.handleRange(returnMethod, input)
}

func (fs FileService) extrapolateKey(path string) string {
	return strings.Replace(strings.Replace(path, fs.rootPath,"",1), "/","", fs.shardLevel + 1)
}
