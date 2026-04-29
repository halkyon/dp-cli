package output

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
		got := PrintTable(servers, false, nil)
		assert.Contains(t, got, "DP-12345")
		assert.NotContains(t, got, "Ubuntu")
	})

	t.Run("Wide table", func(t *testing.T) {
		got := PrintTable(servers, true, nil)
		assert.Contains(t, got, "DP-12345")
		assert.Contains(t, got, "Ubuntu")
		assert.Contains(t, got, "99.99")
		assert.Contains(t, got, "POWER")
	})

	t.Run("Query fields table", func(t *testing.T) {
		queryFields := []string{"Name", "IP"}
		got := PrintTable(servers, false, queryFields)
		assert.Contains(t, got, "NAME")
		assert.Contains(t, got, "IP")
	})

	t.Run("Query Storage field", func(t *testing.T) {
		queryFields := []string{"Name", "Storage"}
		got := PrintTable(servers, false, queryFields)
		assert.Contains(t, got, "NAME")
		assert.Contains(t, got, "STORAGE")
	})

	t.Run("Case-insensitive query fields", func(t *testing.T) {
		queryFields := []string{"name", "STORAGE"}
		got := PrintTable(servers, false, queryFields)
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
		require.NoError(t, PrintCSV(w, servers, false, nil))

		records, err := csv.NewReader(buf).ReadAll()
		require.NoError(t, err)
		assert.Equal(t, []string{"Name", "Alias", "Status", "Location", "IP"}, records[0])
		assert.Equal(t, []string{"DP-12345", "Prod", "ACTIVE", "Ashburn", "1.2.3.4"}, records[1])
	})

	t.Run("CSV Wide", func(t *testing.T) {
		buf := new(bytes.Buffer)
		w := csv.NewWriter(buf)
		require.NoError(t, PrintCSV(w, servers, true, nil))

		records, err := csv.NewReader(buf).ReadAll()
		require.NoError(t, err)
		assert.Equal(t, []string{"Name", "Alias", "Status", "Power", "Location", "IP", "OS", "CPU", "Memory", "Storage", "Price"}, records[0])
	})

	t.Run("CSV Query Fields", func(t *testing.T) {
		buf := new(bytes.Buffer)
		w := csv.NewWriter(buf)
		queryFields := []string{"Name", "IP"}
		require.NoError(t, PrintCSV(w, servers, false, queryFields))

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
		encoded, err := PrintJSON(servers, nil)
		require.NoError(t, err)
		assert.Contains(t, string(encoded), `"name": "DP-12345"`)
		assert.Contains(t, string(encoded), `"ip": "1.2.3.4"`)
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

func TestPrintRaw(t *testing.T) {
	servers := []server.Server{
		{
			Name:  "DP-12345",
			Alias: "prod-web",
			IP:    "1.2.3.4",
		},
		{
			Name:  "DP-67890",
			Alias: "staging-db",
			IP:    "5.6.7.8",
		},
	}

	t.Run("Single field", func(t *testing.T) {
		got := PrintRaw(servers, []string{"Name"})
		assert.Equal(t, "DP-12345\nDP-67890\n", got)
	})

	t.Run("Multiple fields", func(t *testing.T) {
		got := PrintRaw(servers, []string{"Name", "Alias"})
		assert.Equal(t, "DP-12345 prod-web\nDP-67890 staging-db\n", got)
	})

	t.Run("Defaults to Name when no fields", func(t *testing.T) {
		got := PrintRaw(servers, nil)
		assert.Equal(t, "DP-12345\nDP-67890\n", got)
	})

	t.Run("Empty servers", func(t *testing.T) {
		got := PrintRaw([]server.Server{}, []string{"Name"})
		assert.Empty(t, got)
	})
}
