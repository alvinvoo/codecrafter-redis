package rdb

import (
	"fmt"
	"time"

	"github.com/codecrafters-io/redis-starter-go/app/util"
)

type StoreWorker struct {
	db int
}

type option struct {
	storedAt  time.Time
	expiresIn time.Duration
}

type store struct {
	data   []byte
	option *option
}

var items = make(map[string]store)

func (p *StoreWorker) StartDatabase(n int) {
	p.db = n
}

func (p *StoreWorker) Meta(k, v []byte) {
	util.DebugLog("Meta ", fmt.Sprintf("k %v v %v", k, v))
}

func (p *StoreWorker) Set(k, v []byte, expiry int64) {
	util.DebugLog("Set ", fmt.Sprintf("k %v v %v expiry %v", k, v, expiry))
}

// expiry is in miliseconds
func StoreRDB(k string, v []byte, expiry int64) error {
	if expiry == 0 {
		items[k] = store{
			data: v,
		}
	} else {
		duration, err := time.ParseDuration(fmt.Sprintf("%dms", expiry))

		if err != nil {
			return err
		}

		items[k] = store{
			data: v,
			option: &option{
				storedAt:  time.Now(),
				expiresIn: duration,
			},
		}
	}

	return nil
}

func GetRDB(k string) (data []byte, ok bool) {
	val, ok := items[k]

	if val.option != nil {
		if (val.option.storedAt.Add(val.option.expiresIn)).Before(time.Now()) {
			delete(items, k)
			return nil, false
		}
	}

	return val.data, ok
}
