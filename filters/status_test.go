package filters

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStatus_Get(t *testing.T) {
	s := NewStatus()
	statuses, err := s.Get(context.Background())
	require.NoError(t, err)
	assert.Equal(t, []string{"WAITING", "PROVISIONING", "MAINTENANCE", "UNREACHABLE", "ACTIVE"}, statuses)
}
