package filters

import "time"

// Cacheable indicates that a filter's results can be cached.
type Cacheable interface {
	CacheDuration() time.Duration
	CacheKey() string
}
