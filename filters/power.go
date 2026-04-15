package filters

import (
	"context"
)

var powerStatuses = []string{"OFF", "ON", "UNKNOWN"}

type Power struct{}

func NewPower() *Power {
	return &Power{}
}

func (*Power) Get(ctx context.Context) ([]string, error) {
	return powerStatuses, nil
}
