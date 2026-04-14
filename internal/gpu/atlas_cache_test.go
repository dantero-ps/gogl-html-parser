package gpu

import (
	"testing"
)

func TestAtlasCacheEvictionRemovesOldest(t *testing.T) {
	p := &GPUPainter{fontCache: make(map[string]*atlasCacheEntry)}

	p.fontCache["a"] = &atlasCacheEntry{lastAccess: 1}
	p.fontCache["b"] = &atlasCacheEntry{lastAccess: 2}
	p.fontCache["c"] = &atlasCacheEntry{lastAccess: 3}

	p.evictOldest()

	if _, ok := p.fontCache["a"]; ok {
		t.Error("entry 'a' (lastAccess=1) should have been evicted")
	}
	if _, ok := p.fontCache["b"]; !ok {
		t.Error("entry 'b' (lastAccess=2) should still exist")
	}
	if _, ok := p.fontCache["c"]; !ok {
		t.Error("entry 'c' (lastAccess=3) should still exist")
	}
}

func TestAtlasCacheEvictionEmptyMap(t *testing.T) {
	p := &GPUPainter{fontCache: make(map[string]*atlasCacheEntry)}

	p.evictOldest()

	if len(p.fontCache) != 0 {
		t.Error("evictOldest on empty map should be a no-op")
	}
}
