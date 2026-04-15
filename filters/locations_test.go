package filters

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/halkyon/dp/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLocations_Get(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]any{
			"data": map[string]any{
				"locations": []map[string]any{
					{"name": "Amsterdam", "region": "EU"},
					{"name": "New York", "region": "NA"},
					{"name": "Singapore", "region": "AP"},
				},
			},
		}
		require.NoError(t, json.NewEncoder(w).Encode(resp))
	}))
	defer server.Close()

	client, err := api.NewClient("test-key")
	require.NoError(t, err)
	client.SetBaseURL(server.URL)

	locs := NewLocations(client, time.Hour)

	t.Run("First call (cache miss)", func(t *testing.T) {
		locations, err := locs.Get(context.Background())
		require.NoError(t, err)
		assert.Equal(t, []string{"Amsterdam", "New York", "Singapore"}, locations)
	})

	t.Run("Second call (cache hit)", func(t *testing.T) {
		locations, err := locs.Get(context.Background())
		require.NoError(t, err)
		assert.Equal(t, []string{"Amsterdam", "New York", "Singapore"}, locations)
	})
}

func TestRegions_Get(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]any{
			"data": map[string]any{
				"locations": []map[string]any{
					{"name": "Amsterdam", "region": "EU"},
					{"name": "Frankfurt", "region": "EU"},
					{"name": "New York", "region": "NA"},
					{"name": "Singapore", "region": "AP"},
					{"name": "Tokyo", "region": "AP"},
				},
			},
		}
		require.NoError(t, json.NewEncoder(w).Encode(resp))
	}))
	defer server.Close()

	client, err := api.NewClient("test-key")
	require.NoError(t, err)
	client.SetBaseURL(server.URL)

	regions := NewRegions(client, time.Hour)

	t.Run("First call (cache miss)", func(t *testing.T) {
		regionList, err := regions.Get(context.Background())
		require.NoError(t, err)
		assert.Equal(t, []string{"AP", "EU", "NA"}, regionList)
	})

	t.Run("Second call (cache hit)", func(t *testing.T) {
		regionList, err := regions.Get(context.Background())
		require.NoError(t, err)
		assert.Equal(t, []string{"AP", "EU", "NA"}, regionList)
	})
}
