package api

import (
	"context"
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
		dataBytes, _ := json.Marshal(map[string]any{"test": expectedData})

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodPost, r.Method)
			assert.Equal(t, "Bearer test-key", r.Header.Get("Authorization"))
			assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

			resp := mockResponse{
				Data: dataBytes,
			}
			err := json.NewEncoder(w).Encode(resp)
			require.NoError(t, err)
		}))
		defer server.Close()

		client, _ := NewClient("test-key")
		client.SetBaseURL(server.URL)

		var result map[string]any
		err := client.Query(context.Background(), "query { test }", nil, &result)

		require.NoError(t, err)
		assert.Equal(t, expectedData, result["test"])
	})

	t.Run("GraphQL Error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			resp := mockResponse{
				Errors: []struct {
					Message string `json:"message"`
				}{{Message: "something went wrong"}},
			}
			err := json.NewEncoder(w).Encode(resp)
			require.NoError(t, err)
		}))
		defer server.Close()

		client, _ := NewClient("test-key")
		client.SetBaseURL(server.URL)

		var result map[string]any
		err := client.Query(context.Background(), "query { test }", nil, &result)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "something went wrong")
	})

	t.Run("Missing API Key", func(t *testing.T) {
		_, err := NewClient("")
		require.ErrorIs(t, err, ErrMissingAPIKey)
	})
}
