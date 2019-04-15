package cache

import (
	"testing"
	"time"
)

func TestGetSet(t *testing.T) {
	tbl := T("test")
	if tbl.delTimer == nil {
		t.Fatal("time is null")
	}
	tbl.Add("bobby", "tables", time.Now().Add(time.Second*5))
	time.Sleep((time.Second * 5) + time.Millisecond*10)
	if tbl.Exists("bobby") {
		t.Fatal("shouldn't exist", tbl.Get("bobby"))
	}
}

func TestKeepAlive(t *testing.T) {
	tbl := T("test_keepalive")
	tbl.AddKeepAlive("marco", "polo", time.Millisecond*100)
	time.Sleep(time.Millisecond * 99)
	tbl.Get("marco")
	time.Sleep(time.Millisecond * 50)
	if !tbl.Exists("marco") {
		t.Fatal("should exist")
	}
	time.Sleep(time.Millisecond * 150)
	if tbl.Exists("bobby") {
		t.Fatal("shouldn't exist", tbl.Get("marco"))
	}
}
