package filters

import (
	"testing"

	"github.com/halkyon/dp/testapi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLocations_Get(t *testing.T) {
	var mq testapi.MockQuerier

	locations := NewLocations(&mq, 0)

	result, err := locations.Get(t.Context())
	require.NoError(t, err)
	assert.Equal(t, []string{"Amsterdam", "Ashburn", "Atlanta", "Berlin", "Chicago", "Dallas", "Frankfurt", "Hong Kong", "London", "Los Angeles", "Miami", "New York", "Paris", "Seattle", "Singapore", "Sydney", "Tokyo", "Toronto"}, result)
}

func TestRegions_Get(t *testing.T) {
	var mq testapi.MockQuerier

	regions := NewRegions(&mq, 0)

	result, err := regions.Get(t.Context())
	require.NoError(t, err)
	assert.Equal(t, []string{"AP", "EU", "NA"}, result)
}
