package maputils

import (
        "fmt"
        "sync"
)


type SyncCounterMap struct {
        sync.Mutex
        m map[string]int
}

func NewSyncCounterMap() *SyncCounterMap {
        return &SyncCounterMap{
                m: make(map[string]int),
        }
}

func (sm *SyncCounterMap) Set(k string, v int) {
        sm.Mutex.Lock()
        defer sm.Mutex.Unlock()

        sm.m[k] = v
}

func (sm *SyncCounterMap) Get(k string) (int, bool) {
        sm.Mutex.Lock()
        defer sm.Mutex.Unlock()

        v, ok := sm.m[k]
        return v, ok
}

func (sm *SyncCounterMap) delete(k string) {
        sm.Mutex.Lock()
        defer sm.Mutex.Unlock()

        delete(sm.m, k)
}

func (sm *SyncCounterMap) DecreaseCounter(k string) (int, error) {
        sm.Mutex.Lock()
        defer sm.Mutex.Unlock()

        v, ok := sm.m[k]
        if !ok {
                return 0, fmt.Errorf("Fail to find counter for key %s", k)
        }

        if v > 0 {
                v--
        }

        if v == 0 {
                delete(sm.m, k)
        } else {
                sm.m[k] = v
        }

        return v, nil
}

