package utils

import "sync"

type SyncData[T any] struct {
	value T
	lock  sync.Mutex
}

type Sync[T any] interface {
	Ref() SyncData[T]
	Unlock() SyncData[T]
}

func (t *SyncData[T]) Ref() *T {
	t.lock.Lock()
	return &t.value
}

func (t *SyncData[T]) Unlock() {
	t.lock.Unlock()
}

// var TestData = SyncData[int]{value: 4}
