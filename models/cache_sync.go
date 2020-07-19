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
	flowStockDelaysUpdate
	everyDayUpdate
	paymentRateUpdate
)

var cache = make(map[updateKind]int64)

func init() {
	update(everyDayUpdate)
	go triggerEveryDay()
}

func triggerEveryDay() error {
	now := time.Now()
	local, err := time.LoadLocation("Local")
	if err != nil {
		return nil
	}
	nextDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 10, 0, local)
	duration := time.Until(nextDay)
	time.Sleep(duration)
	for {
		update(everyDayUpdate)
		time.Sleep(24 * time.Hour)
	}
}

// update stores the current time for the given kind
func update(kind updateKind) {
	now := time.Now().UnixNano()
	mutex := &sync.Mutex{}
	mutex.Lock()
	defer mutex.Unlock()
	cache[kind] = now
}

// needUpdate checks if the srcKind needs to be calculated if the last call to
// the update function is older than one of the linkedKinds
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
