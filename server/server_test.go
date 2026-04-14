package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/halkyon/dp/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestServer_List(t *testing.T) {
	expectedServers := []serverNode{
		{
			Name:        "DP-12345",
			Alias:       "Prod-Web",
			Hostname:    "dp-12345.example.com",
			Uptime:      172800,
			StatusV2:    "ACTIVE",
			PowerStatus: "ON",
			Location:    locationInfo{Name: "Ashburn", Region: "NA"},
			Network: networkInfo{
				IPAddresses:    []ipAddress{{IP: "1.2.3.4", Type: "IPV4", IsPrimary: true}},
				UplinkCapacity: 10,
			},
			Billing: billingInfo{
				SubscriptionItem: subscriptionItem{
					Price:    99.99,
					Currency: "USD",
				},
			},
			System: systemInfo{
				OperatingSystem: operatingSystemInfo{Name: "Ubuntu 22.04"},
				Raid:            "NONE",
			},
			Hardware: hardwareInfo{
				CPUs:    []cpuInfo{{Name: "Intel Xeon", Cores: 4}},
				RAMs:    []ramInfo{{Size: 16}},
				Storage: []storageInfo{{Size: 100, Type: "SSD"}},
			},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]any{
			"data": map[string]any{
				"servers": map[string]any{
					"entriesTotalCount": len(expectedServers),
					"pageCount":         1,
					"isLastPage":        true,
					"entries":           expectedServers,
				},
			},
		}
		assert.NoError(t, json.NewEncoder(w).Encode(resp))
	}))
	defer server.Close()

	client, err := api.NewClient("test-key")
	require.NoError(t, err)
	client.SetBaseURL(server.URL)

	t.Run("List servers", func(t *testing.T) {
		servers, err := List(t.Context(), client)
		assert.NoError(t, err)
		assert.Len(t, servers, 1)
		assert.Equal(t, "DP-12345", servers[0].Name)
		assert.Equal(t, "Prod-Web", servers[0].Alias)
		assert.Equal(t, "1.2.3.4", servers[0].IP)
		assert.Equal(t, "Ubuntu 22.04", servers[0].OperatingSystem)
		assert.Equal(t, "Intel Xeon", servers[0].CPU)
		assert.Equal(t, "16 GB", servers[0].Memory)
		assert.Equal(t, "100 GB SSD", servers[0].Storage)
		assert.Equal(t, 99.99, servers[0].Price)
	})
}
