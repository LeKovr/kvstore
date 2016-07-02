// Package kvstore is a simple key-value in memory store
// with persistense in json file
package kvstore

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"sync"

	"github.com/LeKovr/go-base/logger"
)

// -----------------------------------------------------------------------------

// Flags is a package flags sample
// in form ready for use with github.com/jessevdk/go-flags
type Flags struct {
	StoreName string `long:"store_file" default:"store.json" description:"File to store sent codes at program exit"`
}

// -----------------------------------------------------------------------------

// StoreData interface holds sore item methods
type StoreData interface {
	//	UnmarshalJSON(b []byte) (err error)
	Init() (StoreData, error)
	Fetch(buf []byte) (StoreData, error)
}

// DataMap holds stored items
type DataMap map[string]StoreData

// http://stackoverflow.com/a/18523565

// Store holda all store attributes
type Store struct {
	data     DataMap // Data should probably not have any reference fields
	itemType StoreData
	changed  bool
	lock     *sync.RWMutex
	Log      *logger.Log
	Config   *Flags
}

// -----------------------------------------------------------------------------
// Functional options

// Config sets store config from flag var
func Config(c *Flags) func(s *Store) error {
	return func(s *Store) error {
		return s.setConfig(c)
	}
}

// -----------------------------------------------------------------------------
// Internal setters

func (s *Store) setConfig(c *Flags) error {
	s.Config = c
	return nil
}

// -----------------------------------------------------------------------------

// New - store constructor
func New(t StoreData, log *logger.Log, options ...func(*Store) error) (*Store, error) {

	d := make(DataMap)
	s := Store{changed: false, data: d, itemType: t, Log: log.WithField("in", "kvstore")}

	for _, option := range options {
		err := option(&s)
		if err != nil {
			return nil, err
		}
	}
	if s.Config == nil {
		s.Config = &Flags{}
	}
	if s.lock == nil {
		s.lock = &sync.RWMutex{}
	}
	s.Load()
	return &s, nil

}

// -----------------------------------------------------------------------------

// Destroy saves store data into file
func (s *Store) Destroy() {
	_ = s.Save()
}

// -----------------------------------------------------------------------------

// Set saves data in store and returns true if key was rewritten
func (s *Store) Set(key string, d StoreData) bool {
	s.lock.Lock()
	defer s.lock.Unlock()
	_, ok := s.data[key]
	dd, _ := d.Init()
	s.data[key] = dd
	s.changed = true
	s.Log.Debugf("in set: %+v", s)
	return ok

}

// -----------------------------------------------------------------------------

// Get returns data by key and true if it was founded
func (s Store) Get(key string) (StoreData, bool) {
	s.lock.RLock()
	defer s.lock.RUnlock()
	d, ok := s.data[key]
	s.Log.Debugf("in get %s: (%v) %+v", key, ok, d)
	return d, ok
}

// -----------------------------------------------------------------------------

// Del deletes a key and returns true if it was founded
func (s *Store) Del(key string) bool {
	s.lock.Lock()
	defer s.lock.Unlock()
	_, ok := s.data[key]
	delete(s.data, key)
	return ok
}

// -----------------------------------------------------------------------------

// Load reads json file and places it in store
func (s *Store) Load() {

	name := s.Config.StoreName

	if _, err := os.Stat(name); os.IsNotExist(err) {
		// path/to/whatever does not exist
		s.Log.Infof("Store file %s does not exists, setup empty store", name)
		return
	}
	file, err := ioutil.ReadFile(name)
	if err != nil {
		s.Log.Infof("Setup empty store for %s because %+v", name, err)
		return
	}
	var data map[string]json.RawMessage
	err = json.Unmarshal(file, &data)
	if err != nil {
		s.Log.Errorf("Parse store %s error: %+v", name, err)
	}
	s.lock.Lock()
	defer s.lock.Unlock()
	db := DataMap{}
	for k, thing := range data {
		db[k], err = s.itemType.Fetch([]byte(thing))
		if err != nil {
			s.Log.Errorf("Item %s unmarshal error: %+v", k, err)
		}
	}

	s.changed = false
	s.data = db
}

// -----------------------------------------------------------------------------

// Save saves store into json file
func (s *Store) Save() bool {
	s.Log.Debugf("in save: %+v", s.data)
	s.lock.Lock()
	defer s.lock.Unlock()
	if len(s.data) == 0 {
		s.changed = false
		return false
	} else if !s.changed {
		return false
	}
	a, _ := json.MarshalIndent(s.data, "   ", "   ")
	err := ioutil.WriteFile(s.Config.StoreName, a, 0600)
	if err != nil {
		s.Log.Errorf("Save store %s error: %+v", s.Config.StoreName, err)
	}
	s.changed = false
	return true
}
