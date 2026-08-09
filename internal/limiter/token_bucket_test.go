package limiter

import (
	"sync"
	"testing"
	"time"
)

func TestTokenBucket_AllowsUpToCapacity(t *testing.T) {
	tb := NewTokenBucket(5, 1) // capacity 5, refill 1 token/sec

	allowed := 0
	for i := 0; i < 10; i++ {
		if tb.Allow() {
			allowed++
		}
	}

	if allowed != 5 {
		t.Errorf("expected 5 allowed requests, got %d", allowed)
	}
}

func TestTokenBucket_RefillsOverTime(t *testing.T) {
	tb := NewTokenBucket(2, 10) // capacity 2, refill 10 tokens/sec

	// drain the bucket
	tb.Allow()
	tb.Allow()

	if tb.Allow() {
		t.Error("expected bucket to be empty")
	}

	time.Sleep(150 * time.Millisecond) // should refill ~1.5 tokens

	if !tb.Allow() {
		t.Error("expected a token to be available after refill")
	}
}

func TestTokenBucket_ConcurrentAccess(t *testing.T) {
	tb := NewTokenBucket(100, 0) // no refill, fixed capacity

	var wg sync.WaitGroup
	var mu sync.Mutex
	allowed := 0

	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if tb.Allow() {
				mu.Lock()
				allowed++
				mu.Unlock()
			}
		}()
	}

	wg.Wait()

	if allowed != 100 {
		t.Errorf("expected exactly 100 allowed under race, got %d", allowed)
	}
}
