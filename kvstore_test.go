package kvstore

import (
	"encoding/json"
	"reflect"
	"testing"
	"time"

	"github.com/LeKovr/go-base/logger"
)

// -----------------------------------------------------------------------------
// kvstore definition

type PhoneData struct {
	Phone string    `json:"phone"`
	Code  string    `json:"code"`
	Stamp time.Time `json:"stamp"`
}

func (pd PhoneData) Init() (StoreData, error) {
	pd.Stamp = time.Now() // TODO if add_stamp
	return pd, nil
}

func (pd PhoneData) Fetch(buf []byte) (StoreData, error) {
	v := PhoneData{}
	err := json.Unmarshal(buf, &v)
	return v, err
}

// -----------------------------------------------------------------------------
func TestKVStore(t *testing.T) {

	log, err := logger.New() // 	logger.Level("debug"))
	if err != nil {
		t.Error("Expected err to be nil, but got:", err)
	}

	store, _ := New(new(PhoneData), log) // , Config(&Flags{StoreName: "xx"}))
	defer store.Destroy()

	k := "test"
	var (
		data StoreData
		ok   bool
	)

	if data, ok = store.Get(k); ok {
		t.Error("Expected no data but got:", data)
	}

	data0 := PhoneData{Phone: "12345", Code: "code"}

	if ok = store.Set(k, data0); ok {
		t.Error("Expected new item but got overwriting")
	}
	if ok = store.Set(k, data0); !ok {
		t.Error("Expected overwriting but got new item")
	}
	k404 := "test404"
	if data, ok = store.Get(k404); ok {
		t.Error("Expected no data but got:", data)
	}
	if data, ok = store.Get(k); !ok {
		t.Error("Expected data but got nothing")
	}

	data0.Stamp = data.(PhoneData).Stamp // Stamp filled on .Set
	if !reflect.DeepEqual(data0, data) {
		t.Error("Expected equal data but got:", data0, data)
	}
	if ok = store.Del(k); !ok {
		t.Error("Expected delete but got nothing")
	}
	if data, ok = store.Get(k); ok {
		t.Error("Expected no data but got:", data)
	}

}
