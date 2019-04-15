package cache

import (
	"context"
	"sync"
	"time"
)

var (
	tables sync.Map
)

type Table struct {
	lock     sync.Mutex
	name     string
	items    sync.Map
	delTimer *time.Timer
	delTick  time.Duration
	ctx      context.Context
	cf       func()
}

func (t *Table) Init() {
	t.lock.Lock()
	defer t.lock.Unlock()
	if t.delTimer == nil {
		t.delTimer = time.NewTimer(time.Second * 10)
		t.delTick = time.Second * 10
	}
	if t.ctx == nil {
		t.ctx, t.cf = context.WithCancel(context.Background())
	}
	go t.tick()
}

func (t *Table) Stop() {
	if t.cf != nil {
		t.cf()
	}
}

func (t *Table) tick() {
	tt := t.delTimer
	ctx := t.ctx
	for ctx.Err() == nil {
		select {
		case <-tt.C:
			t.cleanup()
		case <-ctx.Done():
			t.lock.Lock()
			t.delTimer.Stop()
			t.delTimer = nil
			t.lock.Unlock()
			return
		}
	}
}

func (t *Table) cleanup() {
	t0 := time.Now()
	dlist := make([]interface{}, 0)
	sintv := time.Second * 10
	t.items.Range(func(key, val interface{}) bool {
		v := val.(*Item)
		v.Lock()
		if v.expiresAt.Before(t0) {
			dlist = append(dlist, key)
		} else {
			if n := v.expiresAt.Sub(t0); n < sintv {
				sintv = n
			}
		}
		v.Unlock()
		return true
	})
	for _, v := range dlist {
		t.items.Delete(v)
	}
	if sintv < time.Second*10 {
		t.lock.Lock()
		t.delTimer.Reset(sintv)
		t.delTick = sintv
		t.lock.Unlock()
	}
}

func T(tableName string) *Table {
	vv, loaded := tables.LoadOrStore(tableName, &Table{
		name: tableName,
	})
	v := vv.(*Table)
	if !loaded {
		v.Init()
	}
	return v
}

func (t *Table) adjustTimer(i *Item) {
	t.lock.Lock()
	defer t.lock.Unlock()
	var dur time.Duration
	if i.interval > 0 {
		dur = i.interval
	} else {
		dur = i.expiresAt.Sub(time.Now())
		if dur < 0 {
			dur = time.Microsecond
		}
	}
	if t.delTimer == nil {
		t.delTimer = time.NewTimer(dur)
		return
	}
	if t.delTick > dur {
		t.delTimer.Reset(dur)
		t.delTick = dur
	}
}

func (t *Table) Add(key, value interface{}, exp time.Time) *Item {
	item := &Item{
		createdOn: time.Now(),
		expiresAt: exp,
		key:       key,
		value:     value,
	}
	t.items.Store(key, item)
	t.adjustTimer(item)
	return item
}

func (t *Table) AddKeepAlive(key, value interface{}, interval time.Duration) *Item {
	item := &Item{
		createdOn: time.Now(),
		expiresAt: time.Now().Add(interval),
		interval:  interval,
		key:       key,
		value:     value,
	}
	t.items.Store(key, item)
	t.adjustTimer(item)
	return item
}

func (t *Table) Exists(key interface{}) bool {
	_, ok := t.items.Load(key)
	return ok
}

func (t *Table) Get(key interface{}) interface{} {
	iitem, ok := t.items.Load(key)
	if !ok {
		return nil
	}
	item := iitem.(*Item)
	item.Lock()
	item.count++
	item.lastAccess = time.Now()
	if item.interval > 0 {
		// is keepalive, update expiry
		item.expiresAt = time.Now().Add(item.interval)
	}
	item.Unlock()
	return item.value
}
