package output

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/halkyon/dp/server"
)

var QueryableFields = []string{
	"Name",
	"Alias",
	"Status",
	"Power",
	"Location",
	"IP",
	"OS",
	"CPU",
	"Memory",
	"Price",
	"Storage",
	"Tags",
	"Hostname",
	"Uptime",
	"Currency",
	"BillingPeriod",
	"Uplink",
	"RAID",
	"IPMIURL",
	"IPMIUser",
	"BGP",
	"DDOS",
	"LinkAggregation",
	"DeviceType",
}

type column struct {
	Name        string
	DisplayName string
	Width       int
}

func getFieldValue(s server.Server, field string) string {
	switch strings.ToLower(field) {
	case "name":
		return s.Name
	case "alias":
		return s.Alias
	case "status":
		return s.Status
	case "power", "powerstatus":
		return s.PowerStatus
	case "location":
		return s.Location
	case "ip":
		return s.IP
	case "os", "operatingsystem":
		return s.OperatingSystem
	case "cpu":
		return s.CPU
	case "memory":
		return s.Memory
	case "price":
		return fmt.Sprintf("%.2f", s.Price)
	case "storage":
		return s.Storage
	case "tags":
		return strings.Join(s.Tags, ", ")
	case "iptype":
		return s.IPType
	case "additionalips":
		return strings.Join(s.AdditionalIPs, ", ")
	case "hostname":
		return s.Hostname
	case "uptime":
		return s.Uptime
	case "currency":
		return s.Currency
	case "billingperiod":
		return s.BillingPeriod
	case "uplink":
		return s.Uplink
	case "raid":
		return s.RAID
	case "ipmiurl":
		return s.IPMIURL
	case "ipmiuser":
		return s.IPMIUser
	case "bgp":
		return strconv.FormatBool(s.BGP)
	case "ddos":
		return s.DDOS
	case "linkaggregation":
		return strconv.FormatBool(s.LinkAggregation)
	case "devicetype":
		return s.DeviceType
	default:
		return ""
	}
}

func buildColumns(wide bool, queryFields []string) []column {
	switch {
	case len(queryFields) > 0:
		cols := make([]column, len(queryFields))
		for i, f := range queryFields {
			cols[i] = column{Name: f, DisplayName: strings.ToUpper(f), Width: 0}
		}
		return cols
	case wide:
		return []column{
			{Name: "Name", DisplayName: "NAME", Width: 0},
			{Name: "Alias", DisplayName: "ALIAS", Width: 0},
			{Name: "Status", DisplayName: "STATUS", Width: 0},
			{Name: "Power", DisplayName: "POWER", Width: 0},
			{Name: "Location", DisplayName: "LOCATION", Width: 0},
			{Name: "IP", DisplayName: "IP", Width: 0},
			{Name: "OS", DisplayName: "OS", Width: 0},
			{Name: "CPU", DisplayName: "CPU", Width: 0},
			{Name: "Memory", DisplayName: "MEMORY", Width: 0},
			{Name: "Storage", DisplayName: "STORAGE", Width: 0},
			{Name: "Price", DisplayName: "PRICE", Width: 0},
		}
	default:
		return []column{
			{Name: "Name", DisplayName: "NAME", Width: 0},
			{Name: "Alias", DisplayName: "ALIAS", Width: 0},
			{Name: "Status", DisplayName: "STATUS", Width: 0},
			{Name: "Location", DisplayName: "LOCATION", Width: 0},
			{Name: "IP", DisplayName: "IP", Width: 0},
		}
	}
}

func PrintTable(servers []server.Server, wide bool, queryFields []string) string {
	if len(servers) == 0 {
		return "No servers found"
	}

	cols := buildColumns(wide, queryFields)

	rowValues := make([][]string, len(servers))
	for si, s := range servers {
		rowValues[si] = make([]string, len(cols))
		for i, c := range cols {
			val := getFieldValue(s, c.Name)
			rowValues[si][i] = val
			cols[i].Width = max(cols[i].Width, len(val))
		}
	}

	var formatBuilder strings.Builder
	for i := range cols {
		cols[i].Width = max(cols[i].Width, len(cols[i].DisplayName)) + 1
		fmt.Fprintf(&formatBuilder, "%%-%ds", cols[i].Width)
	}
	format := formatBuilder.String() + "\n"

	// Pre-allocate a reusable []any slice to avoid per-row allocations.
	fmtArgs := make([]any, len(cols))

	var b strings.Builder

	header := make([]string, len(cols))
	dashes := make([]string, len(cols))
	for i, c := range cols {
		header[i] = c.DisplayName
		dashes[i] = strings.Repeat("-", c.Width-1)
	}
	for i, v := range header {
		fmtArgs[i] = v
	}
	fmt.Fprintf(&b, format, fmtArgs...)
	for i, v := range dashes {
		fmtArgs[i] = v
	}
	fmt.Fprintf(&b, format, fmtArgs...)

	for _, vals := range rowValues {
		for i, v := range vals {
			fmtArgs[i] = v
		}
		fmt.Fprintf(&b, format, fmtArgs...)
	}

	return b.String()
}

func PrintCSV(w *csv.Writer, servers []server.Server, wide bool, queryFields []string) error {
	var fields []string
	switch {
	case len(queryFields) > 0:
		fields = queryFields
	case wide:
		fields = []string{"Name", "Alias", "Status", "Power", "Location", "IP", "OS", "CPU", "Memory", "Storage", "Price"}
	default:
		fields = []string{"Name", "Alias", "Status", "Location", "IP"}
	}

	if err := w.Write(fields); err != nil {
		return err
	}

	for _, s := range servers {
		var record []string
		for _, f := range fields {
			record = append(record, getFieldValue(s, f))
		}
		if err := w.Write(record); err != nil {
			return err
		}
	}

	w.Flush()
	return w.Error()
}

func PrintJSON(servers []server.Server, queryFields []string) ([]byte, error) {
	var encoded []byte
	var err error
	if len(queryFields) > 0 {
		output := make([]map[string]any, len(servers))
		for i, s := range servers {
			m := make(map[string]any)
			for _, f := range queryFields {
				m[f] = getFieldValue(s, f)
			}
			output[i] = m
		}
		encoded, err = json.MarshalIndent(output, "", "  ")
	} else {
		encoded, err = json.MarshalIndent(servers, "", "  ")
	}
	if err != nil {
		return nil, fmt.Errorf("marshaling JSON: %w", err)
	}
	return encoded, nil
}

func PrintRaw(servers []server.Server, queryFields []string) string {
	if len(servers) == 0 {
		return ""
	}
	fields := queryFields
	if len(fields) == 0 {
		fields = []string{"Name"}
	}
	var b strings.Builder
	for _, s := range servers {
		vals := make([]string, len(fields))
		for i, f := range fields {
			vals[i] = getFieldValue(s, f)
		}
		b.WriteString(strings.Join(vals, " "))
		b.WriteString("\n")
	}
	return b.String()
}
