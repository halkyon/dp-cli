package aliases

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/halkyon/dp/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAliasCache_Get(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]any{
			"data": map[string]any{
				"servers": map[string]any{
					"entriesTotalCount": 3,
					"pageCount":         1,
					"isLastPage":        true,
					"entries": []map[string]any{
						{"name": "Server1", "alias": "Alias1"},
						{"name": "Server2", "alias": ""},
						{"name": "Server3", "alias": "Alias3"},
					},
				},
			},
		}
		require.NoError(t, json.NewEncoder(w).Encode(resp))
	}))
	defer server.Close()

	client, err := api.NewClient("test-key")
	require.NoError(t, err)
	client.SetBaseURL(server.URL)

	cache := New(client)

	t.Run("First call (cache miss)", func(t *testing.T) {
		aliases, err := cache.Get(t.Context())
		require.NoError(t, err)
		assert.Equal(t, []string{"Alias1", "Alias3"}, aliases)
	})

	t.Run("Second call (cache hit)", func(t *testing.T) {
		aliases, err := cache.Get(t.Context())
		require.NoError(t, err)
		assert.Equal(t, []string{"Alias1", "Alias3"}, aliases)
	})

	t.Run("Clear cache", func(t *testing.T) {
		require.NoError(t, cache.Clear())

		aliases, err := cache.Get(t.Context())
		require.NoError(t, err)
		assert.Equal(t, []string{"Alias1", "Alias3"}, aliases)
	})
}
