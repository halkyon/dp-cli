package main

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"testing"

	"github.com/halkyon/dp/server"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPrintTable(t *testing.T) {
	servers := []server.Server{
		{
			Name:            "DP-12345",
			Alias:           "Production-Web",
			Status:          "ACTIVE",
			Location:        "Ashburn",
			IP:              "1.2.3.4",
			OperatingSystem: "Ubuntu 22.04",
			CPU:             "4 Cores",
			Memory:          "16 GB",
			Price:           99.99,
		},
		{
			Name:            "DP-67890",
			Alias:           "Staging-DB",
			Status:          "PROVISIONING",
			Location:        "London",
			IP:              "5.6.7.8",
			OperatingSystem: "Debian 12",
			CPU:             "8 Cores",
			Memory:          "32 GB",
			Price:           199.99,
		},
	}

	t.Run("Narrow table", func(t *testing.T) {
		got := printTable(servers, false, nil)
		assert.Contains(t, got, "DP-12345")
		assert.NotContains(t, got, "Ubuntu")
	})

	t.Run("Wide table", func(t *testing.T) {
		got := printTable(servers, true, nil)
		assert.Contains(t, got, "DP-12345")
		assert.Contains(t, got, "Ubuntu")
		assert.Contains(t, got, "99.99")
	})

	t.Run("Query fields table", func(t *testing.T) {
		queryFields := []string{"Name", "IP"}
		got := printTable(servers, false, queryFields)
		assert.Contains(t, got, "NAME")
		assert.Contains(t, got, "IP")
	})

	t.Run("Query Storage field", func(t *testing.T) {
		queryFields := []string{"Name", "Storage"}
		got := printTable(servers, false, queryFields)
		assert.Contains(t, got, "NAME")
		assert.Contains(t, got, "STORAGE")
	})

	t.Run("Case-insensitive query fields", func(t *testing.T) {
		queryFields := []string{"name", "STORAGE"}
		got := printTable(servers, false, queryFields)
		assert.Contains(t, got, "NAME")
		assert.Contains(t, got, "STORAGE")
	})
}

func TestPrintCSV(t *testing.T) {
	servers := []server.Server{
		{
			Name:     "DP-12345",
			Alias:    "Prod",
			Status:   "ACTIVE",
			Location: "Ashburn",
			IP:       "1.2.3.4",
		},
	}

	t.Run("CSV Narrow", func(t *testing.T) {
		buf := new(bytes.Buffer)
		w := csv.NewWriter(buf)
		printCSV(w, servers, false, nil)

		records, err := csv.NewReader(buf).ReadAll()
		require.NoError(t, err)
		assert.Equal(t, []string{"Name", "Alias", "Status", "Location", "IP"}, records[0])
		assert.Equal(t, []string{"DP-12345", "Prod", "ACTIVE", "Ashburn", "1.2.3.4"}, records[1])
	})

	t.Run("CSV Wide", func(t *testing.T) {
		buf := new(bytes.Buffer)
		w := csv.NewWriter(buf)
		printCSV(w, servers, true, nil)

		records, err := csv.NewReader(buf).ReadAll()
		require.NoError(t, err)
		assert.Equal(t, []string{"Name", "Alias", "Status", "Location", "IP", "OS", "CPU", "Memory", "Storage", "Price"}, records[0])
	})

	t.Run("CSV Query Fields", func(t *testing.T) {
		buf := new(bytes.Buffer)
		w := csv.NewWriter(buf)
		queryFields := []string{"Name", "IP"}
		printCSV(w, servers, false, queryFields)

		records, err := csv.NewReader(buf).ReadAll()
		require.NoError(t, err)
		assert.Equal(t, []string{"Name", "IP"}, records[0])
		assert.Equal(t, []string{"DP-12345", "1.2.3.4"}, records[1])
	})
}

func TestJSONMarshaling(t *testing.T) {
	servers := []server.Server{
		{
			Name: "DP-12345",
			IP:   "1.2.3.4",
		},
	}

	t.Run("JSON full", func(t *testing.T) {
		output, err := json.MarshalIndent(servers, "", "  ")
		require.NoError(t, err)
		assert.Contains(t, string(output), `"name": "DP-12345"`)
		assert.Contains(t, string(output), `"ip": "1.2.3.4"`)
	})

	t.Run("JSON query fields", func(t *testing.T) {
		queryFields := []string{"Name", "IP"}

		output := make([]map[string]any, len(servers))
		for i, s := range servers {
			m := make(map[string]any)
			for _, f := range queryFields {
				m[f] = getFieldValue(s, f)
			}
			output[i] = m
		}
		encoded, err := json.MarshalIndent(output, "", "  ")
		require.NoError(t, err)
		assert.Contains(t, string(encoded), `"Name": "DP-12345"`)
		assert.Contains(t, string(encoded), `"IP": "1.2.3.4"`)
		assert.NotContains(t, string(encoded), `"Alias"`)
	})
}
