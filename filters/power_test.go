package filters

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPower_Get(t *testing.T) {
	p := NewPower()
	statuses, err := p.Get(t.Context())
	require.NoError(t, err)
	assert.Equal(t, []string{"OFF", "ON", "UNKNOWN"}, statuses)
}
