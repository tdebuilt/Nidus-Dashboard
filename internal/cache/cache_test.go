package cache

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestGetSetBasic(t *testing.T) {
	t.Parallel()
	c := New(5*time.Minute, time.Hour)
	defer c.Stop()

	c.Set("key1", "value1")

	val, ok := c.Get("key1")
	if !ok {
		t.Fatal("expected key1 to exist")
	}
	if val != "value1" {
		t.Fatalf("expected value1, got %v", val)
	}
}

func TestGetMiss(t *testing.T) {
	t.Parallel()
	c := New(5*time.Minute, time.Hour)
	defer c.Stop()

	val, ok := c.Get("nonexistent")
	if ok {
		t.Fatal("expected miss for nonexistent key")
	}
	if val != nil {
		t.Fatalf("expected nil, got %v", val)
	}
}

func TestTTLExpiration(t *testing.T) {
	t.Parallel()
	c := New(50*time.Millisecond, time.Hour)
	defer c.Stop()

	c.Set("key1", "value1")

	val, ok := c.Get("key1")
	if !ok || val != "value1" {
		t.Fatal("expected hit before expiry")
	}

	time.Sleep(60 * time.Millisecond)

	val, ok = c.Get("key1")
	if ok {
		t.Fatal("expected miss after TTL expiry")
	}
	if val != nil {
		t.Fatalf("expected nil after expiry, got %v", val)
	}
}

func TestCustomTTL(t *testing.T) {
	t.Parallel()
	c := New(time.Hour, time.Hour)
	defer c.Stop()

	c.SetWithTTL("short", "data", 50*time.Millisecond)
	c.Set("long", "data")

	time.Sleep(60 * time.Millisecond)

	_, ok := c.Get("short")
	if ok {
		t.Fatal("expected short TTL key to expire")
	}

	_, ok = c.Get("long")
	if !ok {
		t.Fatal("expected long TTL key to still exist")
	}
}

func TestInvalidateSingleKey(t *testing.T) {
	t.Parallel()
	c := New(5*time.Minute, time.Hour)
	defer c.Stop()

	c.Set("key1", "value1")
	c.Set("key2", "value2")

	c.Invalidate("key1")

	_, ok := c.Get("key1")
	if ok {
		t.Fatal("expected key1 to be invalidated")
	}

	val, ok := c.Get("key2")
	if !ok || val != "value2" {
		t.Fatal("expected key2 to still exist")
	}
}

func TestInvalidatePrefix(t *testing.T) {
	t.Parallel()
	c := New(5*time.Minute, time.Hour)
	defer c.Stop()

	c.Set("portainer:containers", "data1")
	c.Set("portainer:stacks", "data2")
	c.Set("proxmox:vms", "data3")

	c.InvalidatePrefix("portainer:")

	_, ok := c.Get("portainer:containers")
	if ok {
		t.Fatal("expected portainer:containers to be invalidated")
	}

	_, ok = c.Get("portainer:stacks")
	if ok {
		t.Fatal("expected portainer:stacks to be invalidated")
	}

	val, ok := c.Get("proxmox:vms")
	if !ok || val != "data3" {
		t.Fatal("expected proxmox:vms to still exist")
	}
}

func TestInvalidateAll(t *testing.T) {
	t.Parallel()
	c := New(5*time.Minute, time.Hour)
	defer c.Stop()

	c.Set("key1", "val1")
	c.Set("key2", "val2")

	c.InvalidateAll()

	if c.Len() != 0 {
		t.Fatalf("expected 0 items, got %d", c.Len())
	}
}

func TestStats(t *testing.T) {
	t.Parallel()
	c := New(5*time.Minute, time.Hour)
	defer c.Stop()
	c.ResetStats()

	c.Set("key1", "value1")

	c.Get("key1") // hit
	c.Get("key1") // hit
	c.Get("miss") // miss

	hits, misses := c.Stats()
	if hits != 2 {
		t.Fatalf("expected 2 hits, got %d", hits)
	}
	if misses != 1 {
		t.Fatalf("expected 1 miss, got %d", misses)
	}
}

func TestResetStats(t *testing.T) {
	t.Parallel()
	c := New(5*time.Minute, time.Hour)
	defer c.Stop()

	c.Set("key1", "val1")
	c.Get("key1")
	c.Get("miss")

	c.ResetStats()
	hits, misses := c.Stats()
	if hits != 0 || misses != 0 {
		t.Fatalf("expected 0/0 after reset, got %d/%d", hits, misses)
	}
}

func TestCacheDuplicateCallsOnlyOneFetch(t *testing.T) {
	t.Parallel()
	c := New(5*time.Minute, time.Hour)
	defer c.Stop()
	c.ResetStats()

	fetchCount := 0
	fetchData := func() string {
		fetchCount++
		return "fetched_data"
	}

	// First call: cache miss, fetch data
	val, ok := c.Get("service:data")
	if !ok {
		val = fetchData()
		c.Set("service:data", val)
	}
	if val != "fetched_data" {
		t.Fatalf("expected fetched_data, got %v", val)
	}

	// Second call: cache hit, no fetch
	val, ok = c.Get("service:data")
	if !ok {
		val = fetchData()
		c.Set("service:data", val)
	}
	if val != "fetched_data" {
		t.Fatalf("expected fetched_data on second call, got %v", val)
	}

	if fetchCount != 1 {
		t.Fatalf("expected 1 fetch, got %d (cache should prevent second call)", fetchCount)
	}

	hits, misses := c.Stats()
	if hits != 1 {
		t.Fatalf("expected 1 hit, got %d", hits)
	}
	if misses != 1 {
		t.Fatalf("expected 1 miss, got %d", misses)
	}
}

func TestActionInvalidatesCache(t *testing.T) {
	t.Parallel()
	c := New(5*time.Minute, time.Hour)
	defer c.Stop()

	fetchCount := 0
	fetchData := func() string {
		fetchCount++
		return fmt.Sprintf("data_v%d", fetchCount)
	}

	// Initial fetch
	val, ok := c.Get("portainer:containers")
	if !ok {
		val = fetchData()
		c.Set("portainer:containers", val)
	}
	if val != "data_v1" {
		t.Fatalf("expected data_v1, got %v", val)
	}

	// Simulate action (e.g., restart container) → invalidate
	c.InvalidatePrefix("portainer:")

	// Next fetch should miss and re-fetch
	_, ok = c.Get("portainer:containers")
	if ok {
		t.Fatal("expected cache miss after invalidation")
	}
	val = fetchData()
	c.Set("portainer:containers", val)

	if val != "data_v2" {
		t.Fatalf("expected data_v2, got %v", val)
	}
	if fetchCount != 2 {
		t.Fatalf("expected 2 fetches (invalidation forced re-fetch), got %d", fetchCount)
	}
}

func TestCleanupRemovesExpired(t *testing.T) {
	t.Parallel()
	c := New(50*time.Millisecond, 30*time.Millisecond)
	defer c.Stop()

	c.Set("key1", "val1")
	c.Set("key2", "val2")

	if c.Len() != 2 {
		t.Fatalf("expected 2 items, got %d", c.Len())
	}

	// Wait for entries to expire and cleanup to run
	time.Sleep(100 * time.Millisecond)

	if c.Len() != 0 {
		t.Fatalf("expected 0 items after cleanup, got %d", c.Len())
	}
}

func TestConcurrentAccess(t *testing.T) {
	t.Parallel()
	c := New(5*time.Minute, time.Hour)
	defer c.Stop()

	var wg sync.WaitGroup
	const goroutines = 50

	// Concurrent writes
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			key := fmt.Sprintf("key:%d", i)
			c.Set(key, i)
		}(i)
	}
	wg.Wait()

	// Concurrent reads
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			key := fmt.Sprintf("key:%d", i)
			val, ok := c.Get(key)
			if !ok {
				t.Errorf("expected key:%d to exist", i)
				return
			}
			if val != i {
				t.Errorf("expected %d, got %v", i, val)
			}
		}(i)
	}
	wg.Wait()

	// Concurrent mixed operations
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			key := fmt.Sprintf("key:%d", i)
			switch i % 3 {
			case 0:
				c.Invalidate(key)
			case 1:
				c.Get(key)
			default:
				c.Set(key, i*10)
			}
		}(i)
	}
	wg.Wait()
}

func TestLenCountsOnlyStored(t *testing.T) {
	t.Parallel()
	c := New(5*time.Minute, time.Hour)
	defer c.Stop()

	if c.Len() != 0 {
		t.Fatalf("expected 0 items in fresh cache, got %d", c.Len())
	}

	c.Set("a", 1)
	c.Set("b", 2)
	if c.Len() != 2 {
		t.Fatalf("expected 2, got %d", c.Len())
	}

	c.Invalidate("a")
	if c.Len() != 1 {
		t.Fatalf("expected 1, got %d", c.Len())
	}
}

func TestOverwriteKey(t *testing.T) {
	t.Parallel()
	c := New(5*time.Minute, time.Hour)
	defer c.Stop()

	c.Set("key1", "v1")
	c.Set("key1", "v2")

	val, ok := c.Get("key1")
	if !ok || val != "v2" {
		t.Fatalf("expected v2, got %v", val)
	}
	if c.Len() != 1 {
		t.Fatalf("overwrite should not increase count, got %d", c.Len())
	}
}
