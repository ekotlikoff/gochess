package gateway

// Credit to https://stackoverflow.com/questions/25484122/map-with-ttl-option-in-go

import (
	"errors"
	"sync"
	"time"

	matchserver "github.com/ekotlikoff/gochess/internal/server/backend/match"
)

type item struct {
	value      *matchserver.Player
	lastAccess int64
}

// TTLMap is a map with a TTL
type TTLMap struct {
	m map[string]*item
	l sync.Mutex
}

// NewTTLMap creates a new map
func NewTTLMap(ln int, maxTTL int, gcFrequencySecs int) (m *TTLMap) {
	m = &TTLMap{m: make(map[string]*item, ln)}
	go func() {
		gcFrequency := time.Tick(time.Second * time.Duration(gcFrequencySecs))
		for now := range gcFrequency {
			m.l.Lock()
			for k, v := range m.m {
				if now.Unix()-v.lastAccess > int64(maxTTL) {
					delete(m.m, k)
				}
			}
			m.l.Unlock()
		}
	}()
	return
}

// Len returns the length of the map
func (m *TTLMap) Len() int {
	m.l.Lock()
	defer m.l.Unlock()
	return len(m.m)
}

// Put puts key k and value v
func (m *TTLMap) Put(k string, v *matchserver.Player) error {
	m.l.Lock()
	defer m.l.Unlock()
	_, ok := m.m[k]
	if !ok {
		it := &item{value: v}
		it.lastAccess = time.Now().Unix()
		m.m[k] = it
	} else {
		return errors.New("failed to put key: " + k + ", value: " + v.Name())
	}
	return nil
}

// Get gets value for key k
func (m *TTLMap) Get(k string) (v *matchserver.Player, err error) {
	m.l.Lock()
	defer m.l.Unlock()
	if it, ok := m.m[k]; ok {
		v = it.value
		it.lastAccess = time.Now().Unix()
	} else {
		err = errors.New("failed to get")
	}
	return

}

// Refresh updates the key k to newk
func (m *TTLMap) Refresh(k, newk string) error {
	m.l.Lock()
	defer m.l.Unlock()
	it, ok := m.m[k]
	if ok {
		it.lastAccess = time.Now().Unix()
	}
	if _, newok := m.m[newk]; !newok && ok {
		m.m[newk] = it
		delete(m.m, k)
	} else {
		return errors.New("failed to refresh key")
	}
	return nil
}
