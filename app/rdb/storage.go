package rdb

import (
	"fmt"

	"github.com/codecrafters-io/redis-starter-go/app/util"
)

type Store struct {
	db int
}

func (p *Store) StartDatabase(n int) {
	p.db = n
}

func (p *Store) Meta(k, v []byte) {
	util.DebugLog("Meta ", fmt.Sprintf("k %v v %v", k, v))
}

func (p *Store) Set(k, v []byte, expiry int64) {
	util.DebugLog("Set ", fmt.Sprintf("k %v v %v expiry %v", k, v, expiry))
}
