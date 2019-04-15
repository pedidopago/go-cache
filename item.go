package cache

import (
	"sync"
	"time"
)

type Item struct {
	sync.Mutex
	expiresAt  time.Time
	interval   time.Duration
	key        interface{}
	value      interface{}
	createdOn  time.Time
	lastAccess time.Time
	count      int64
}

func (i *Item) Key() interface{} {
	// key and value are immutable
	return i.key
}

func (i *Item) Value() interface{} {
	// key and value are immutable
	return i.value
}
