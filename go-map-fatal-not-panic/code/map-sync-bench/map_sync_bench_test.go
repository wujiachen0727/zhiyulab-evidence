package mapsyncbench

import (
	"sync"
	"testing"
)

const keySpace = 4096

var (
	sinkInt int
	sinkBool bool
	sinkAny any
)

type lockedMap struct {
	mu sync.RWMutex
	m  map[int]int
}

func newPlainMap() map[int]int {
	m := make(map[int]int, keySpace)
	for i := 0; i < keySpace; i++ {
		m[i] = i
	}
	return m
}

func newLockedMap() *lockedMap {
	return &lockedMap{m: newPlainMap()}
}

func newSyncMap() *sync.Map {
	var m sync.Map
	for i := 0; i < keySpace; i++ {
		m.Store(i, i)
	}
	return &m
}

func BenchmarkPlainMapReadMostly(b *testing.B) {
	m := newPlainMap()
	var sum int
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		key := i & (keySpace - 1)
		if i%100 == 0 {
			m[key] = i
			continue
		}
		sum += m[key]
	}

	sinkInt = sum
}

func BenchmarkLockedMapReadMostly(b *testing.B) {
	lm := newLockedMap()
	var sum int
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		key := i & (keySpace - 1)
		if i%100 == 0 {
			lm.mu.Lock()
			lm.m[key] = i
			lm.mu.Unlock()
			continue
		}
		lm.mu.RLock()
		sum += lm.m[key]
		lm.mu.RUnlock()
	}

	sinkInt = sum
}

func BenchmarkSyncMapReadMostly(b *testing.B) {
	m := newSyncMap()
	var sum int
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		key := i & (keySpace - 1)
		if i%100 == 0 {
			m.Store(key, i)
			continue
		}
		value, ok := m.Load(key)
		sinkBool = ok
		if ok {
			sum += value.(int)
		}
	}

	sinkInt = sum
}

func BenchmarkPlainMapWriteHeavy(b *testing.B) {
	m := newPlainMap()
	var sum int
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		key := i & (keySpace - 1)
		if i%2 == 0 {
			m[key] = i
		} else {
			sum += m[key]
		}
	}

	sinkInt = sum
}

func BenchmarkLockedMapWriteHeavy(b *testing.B) {
	lm := newLockedMap()
	var sum int
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		key := i & (keySpace - 1)
		if i%2 == 0 {
			lm.mu.Lock()
			lm.m[key] = i
			lm.mu.Unlock()
		} else {
			lm.mu.RLock()
			sum += lm.m[key]
			lm.mu.RUnlock()
		}
	}

	sinkInt = sum
}

func BenchmarkSyncMapWriteHeavy(b *testing.B) {
	m := newSyncMap()
	var sum int
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		key := i & (keySpace - 1)
		if i%2 == 0 {
			m.Store(key, i)
		} else {
			value, ok := m.Load(key)
			sinkBool = ok
			if ok {
				sum += value.(int)
			}
		}
	}

	sinkInt = sum
}

func BenchmarkSyncMapRangeSnapshot(b *testing.B) {
	m := newSyncMap()
	var count int
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		localCount := 0
		m.Range(func(key, value any) bool {
			sinkAny = key
			localCount++
			return localCount < 16
		})
		count += localCount
	}

	sinkInt = count
}
