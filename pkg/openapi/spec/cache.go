package spec

import (
	"sync"

	specs "github.com/go-openapi/spec"
)

type simpleCache struct {
	lock  sync.RWMutex
	store map[string]interface{}
}

func (s *simpleCache) ShallowClone() specs.ResolutionCache {
	store := make(map[string]interface{}, len(s.store))
	s.lock.RLock()
	for k, v := range s.store {
		store[k] = v
	}
	s.lock.RUnlock()

	return &simpleCache{
		store: store,
	}
}

// Get retrieves a cached URI
func (s *simpleCache) Get(uri string) (interface{}, bool) {
	s.lock.RLock()
	v, ok := s.store[uri]

	s.lock.RUnlock()
	return v, ok
}

// Set caches a URI
func (s *simpleCache) Set(uri string, data interface{}) {
	s.lock.Lock()
	s.store[uri] = data
	s.lock.Unlock()
}

var (
	// resCache is a package level cache for $ref resolution and expansion.
	// It is initialized lazily by methods that have the need for it: no
	// memory is allocated unless some composeer methods are called.
	//
	// It is initialized with JSON schema and swagger schema,
	// which do not mutate during normal operations.
	//
	// All subsequent utilizations of this cache are produced from a shallow
	// clone of this initial version.
	resCache  *simpleCache
	onceCache sync.Once

	_ specs.ResolutionCache = &simpleCache{}
)

// initResolutionCache initializes the URI resolution cache. To be wrapped in a sync.Once.Do call.
func initResolutionCache() {
	resCache = defaultResolutionCache()
}

func defaultResolutionCache() *simpleCache {
	return &simpleCache{store: map[string]interface{}{}}
}

func cacheOrDefault(cache specs.ResolutionCache) specs.ResolutionCache {
	onceCache.Do(initResolutionCache)

	if cache != nil {
		return cache
	}

	// get a shallow clone of the base cache with swagger and json schema
	return resCache.ShallowClone()
}
