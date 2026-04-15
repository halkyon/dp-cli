package filters

import (
	"context"
)

var serverStatuses = []string{"WAITING", "PROVISIONING", "MAINTENANCE", "UNREACHABLE", "ACTIVE"}

type Status struct{}

func NewStatus() *Status {
	return &Status{}
}

func (*Status) Get(ctx context.Context) ([]string, error) {
	return serverStatuses, nil
}
