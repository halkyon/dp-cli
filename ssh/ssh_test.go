package ssh

import (
	"testing"

	"github.com/halkyon/dp/server"
	"github.com/stretchr/testify/assert"
)

func TestRun(t *testing.T) {
	servers := []server.Server{
		{
			Name:            "Server1",
			Alias:           "S1",
			IP:              "127.0.0.1",
			OperatingSystem: "Linux",
		},
	}

	t.Run("No server found", func(t *testing.T) {
		assert.ErrorContains(t, Run(t.Context(), servers, "user", []string{"ls"}), "no server found")
	})

	t.Run("Server has no IP", func(t *testing.T) {
		serversNoIP := []server.Server{
			{Name: "ServerNoIP", Alias: "SNoIP", IP: ""},
		}
		assert.ErrorContains(t, Run(t.Context(), serversNoIP, "user", []string{"SNoIP", "ls"}), "has no IP address")
	})
}
