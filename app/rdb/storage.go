package rdb

import (
	"fmt"
	"time"

	"github.com/codecrafters-io/redis-starter-go/app/util"
)

type StoreWorker struct {
	db uint32
}

type option struct {
	expiresAt int64
}

type store struct {
	data   []byte
	option *option
}

var items = make(map[string]store)

func (p *StoreWorker) StartRDB() {
	util.DebugLog("StartRDB: Start parsing RDB")
}

func (p *StoreWorker) StartDatabase(n uint32) {
	p.db = n
}

func (p *StoreWorker) Meta(k, v []byte) {
	util.DebugLog("Meta ", fmt.Sprintf("k %v v %v", k, v))
}

func (p *StoreWorker) Set(k, v []byte, expiry int64) {
	util.DebugLog("Set ", fmt.Sprintf("k %v v %v expiry %v", k, v, expiry))
	storeRDB(string(k), v, expiry)
}

func (p *StoreWorker) EndDatabase(n uint32) {
	p.db = n
}

func (p *StoreWorker) EndRDB() {
	util.DebugLog("EndRDB: End parsing RDB")
}

func storeRDB(k string, v []byte, expiry int64) {
	if expiry == 0 {
		items[k] = store{
			data: v,
		}
	} else {
		items[k] = store{
			data: v,
			option: &option{
				expiresAt: expiry,
			},
		}
	}
}

// expiryDuration is in miliseconds
func StoreRDB(k string, v []byte, expiryDuration int64) error {
	if expiryDuration == 0 {
		items[k] = store{
			data: v,
		}
	} else {
		duration, err := time.ParseDuration(fmt.Sprintf("%dms", expiryDuration))

		if err != nil {
			return err
		}

		expiresAt := time.Now().Add(duration).UnixMilli()
		items[k] = store{
			data: v,
			option: &option{
				expiresAt: expiresAt,
			},
		}
	}

	return nil
}

func GetRDB(k string) ([]byte, bool) {
	val, ok := items[k]

	if val.option != nil {
		if time.Now().UnixMilli() > val.option.expiresAt {
			delete(items, k)
			return nil, false
		}
	}

	return val.data, ok
}
