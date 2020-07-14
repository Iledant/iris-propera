package models

import (
	"sync"
	"time"
)

type updateKind int

const (
	paymentDemandsUpdate updateKind = iota
	paymentUpdate
	csfWeekTrendUpdate
)

var cache = make(map[updateKind]int64)

func update(kind updateKind) {
	now := time.Now().UnixNano()
	mutex := &sync.Mutex{}
	mutex.Lock()
	defer mutex.Unlock()
	cache[kind] = now
}

func needUpdate(srcKind updateKind, linkedKinds ...updateKind) bool {
	srcTime, ok := cache[srcKind]
	if !ok {
		return true
	}
	var linkedTime int64
	for _, l := range linkedKinds {
		linkedTime, ok = cache[l]
		if ok && linkedTime > srcTime {
			return true
		}
	}
	return false
}
