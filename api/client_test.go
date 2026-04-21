package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockResponse struct {
	Data   json.RawMessage `json:"data"`
	Errors []struct {
		Message string `json:"message"`
	} `json:"errors"`
}

func TestClient_Query(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		expectedData := map[string]any{"test": "value"}
		dataBytes, err := json.Marshal(map[string]any{"test": expectedData})
		require.NoError(t, err)

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodPost, r.Method)
			assert.Equal(t, "Bearer test-key", r.Header.Get("Authorization"))
			assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

			resp := mockResponse{
				Data: dataBytes,
			}
			assert.NoError(t, json.NewEncoder(w).Encode(resp))
		}))
		defer server.Close()

		client, err := NewClient("test-key")
		require.NoError(t, err)
		client.SetBaseURL(server.URL)

		var result map[string]any
		require.NoError(t, client.Query(t.Context(), "query { test }", nil, &result))
		assert.Equal(t, expectedData, result["test"])
	})

	t.Run("GraphQL Error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			resp := mockResponse{
				Errors: []struct {
					Message string `json:"message"`
				}{{Message: "something went wrong"}},
			}
			assert.NoError(t, json.NewEncoder(w).Encode(resp))
		}))
		defer server.Close()

		client, err := NewClient("test-key")
		require.NoError(t, err)
		client.SetBaseURL(server.URL)

		var result map[string]any
		require.ErrorContains(t, client.Query(t.Context(), "query { test }", nil, &result), "something went wrong")
	})

	t.Run("Missing API Key", func(t *testing.T) {
		_, err := NewClient("")
		require.ErrorIs(t, err, ErrMissingAPIKey)
	})
}
