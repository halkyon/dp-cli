package filters

import (
	"context"
	"time"
)

var powerStatuses = []string{"OFF", "ON", "UNKNOWN"}

type Power struct{}

func NewPower() *Power {
	return &Power{}
}

func (*Power) Get(ctx context.Context) ([]string, error) {
	return powerStatuses, nil
}

func (*Power) CacheDuration() time.Duration { return 0 }
func (*Power) CacheKey() string             { return "power" }
