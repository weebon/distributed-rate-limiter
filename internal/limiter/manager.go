package limiter

import (
	"sync"
)

type Manager struct {
	mu         sync.Mutex
	buckets    map[string]*TokenBucket
	capacity   float64
	refillRate float64
}

func NewManager(capacity, refillRate float64) *Manager {
	return &Manager{
		buckets:    make(map[string]*TokenBucket),
		capacity:   capacity,
		refillRate: refillRate,
	}
}

func (m *Manager) Allow(key string) bool {
	bucket := m.getBucket(key)
	return bucket.Allow()
}

func (m *Manager) getBucket(key string) *TokenBucket {
	m.mu.Lock()
	defer m.mu.Unlock()

	bucket, exists := m.buckets[key]
	if !exists {
		bucket = NewTokenBucket(m.capacity, m.refillRate)
		m.buckets[key] = bucket
	}
	return bucket
}