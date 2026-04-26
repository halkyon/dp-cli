package filters

import (
	"context"
	"time"
)

var serverStatuses = []string{"WAITING", "PROVISIONING", "MAINTENANCE", "UNREACHABLE", "ACTIVE"}

type Status struct{}

func NewStatus() *Status {
	return &Status{}
}

func (*Status) Get(ctx context.Context) ([]string, error) {
	return serverStatuses, nil
}

func (*Status) CacheDuration() time.Duration { return 0 }
func (*Status) CacheKey() string             { return "status" }
