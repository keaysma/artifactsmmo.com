package utils

import "sync"

type SyncData[T any] struct {
	Value T
	lock  sync.Mutex
}

type Sync[T any] interface {
	Ref() SyncData[T]
	Unlock() SyncData[T]
}

func (t *SyncData[T]) Ref() *T {
	t.lock.Lock()
	return &t.Value
}

func (t *SyncData[T]) Unlock() {
	t.lock.Unlock()
}

func (t *SyncData[T]) With(f func(value *T) *T) {
	t.lock.Lock()
	t.Value = *f(&t.Value)
	t.lock.Unlock()
}

func (t *SyncData[T]) Set(new_value *T) {
	t.lock.Lock()
	t.Value = *new_value
	t.lock.Unlock()
}

func (t *SyncData[T]) ShallowCopy() T {
	t.lock.Lock()
	defer t.lock.Unlock()
	return t.Value
}

// var TestData = SyncData[int]{value: 4}
