package main

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"runtime/debug"
	"strconv"
	"strings"

	"time"

	"github.com/halkyon/dp/api"
	"github.com/halkyon/dp/cache"
	"github.com/halkyon/dp/completion"
	"github.com/halkyon/dp/config"
	"github.com/halkyon/dp/filters"
	"github.com/halkyon/dp/server"
	"github.com/halkyon/dp/ssh"
)

var version = ""

var outputFormatFlag string
var outputWide bool
var queryFields = new(stringSlice)

var flagName = new(stringSlice)
var flagAlias = new(stringSlice)
var flagLocation = new(stringSlice)
var flagRegion = new(stringSlice)
var flagStatus = new(stringSlice)
var flagPower = new(stringSlice)
var flagTag = new(stringSlice)
var flagUser = new(stringSlice)

type stringSlice []string

func (s *stringSlice) Set(val string) error {
	*s = append(*s, val)
	return nil
}

func (s *stringSlice) String() string {
	return strings.Join(*s, ",")
}

func init() {
	flag.StringVar(&outputFormatFlag, "o", "", "Output format (shorthand): json, table, csv, raw")
	flag.StringVar(&outputFormatFlag, "output", "", "Output format: json, table, csv, raw")
	flag.BoolVar(&outputWide, "ow", false, "Show more fields (shorthand)")
	flag.Bool("output-wide", false, "Show more fields in table/csv output")
	flag.Var(queryFields, "q", "Output specific field(s) (repeatable)")
	flag.Var(queryFields, "query", "Output specific field(s) (repeatable)")

	flag.Var(flagName, "n", "Filter by name (repeatable)")
	flag.Var(flagName, "name", "Filter by name (repeatable)")
	flag.Var(flagAlias, "a", "Filter by alias (repeatable)")
	flag.Var(flagAlias, "alias", "Filter by alias (repeatable)")
	flag.Var(flagLocation, "l", "Filter by location (repeatable)")
	flag.Var(flagLocation, "location", "Filter by location (repeatable)")
	flag.Var(flagRegion, "r", "Filter by region (repeatable)")
	flag.Var(flagRegion, "region", "Filter by region (repeatable)")
	flag.Var(flagStatus, "s", "Filter by status (repeatable)")
	flag.Var(flagStatus, "status", "Filter by status (repeatable)")
	flag.Var(flagPower, "p", "Filter by power status (repeatable)")
	flag.Var(flagPower, "power", "Filter by power status (repeatable)")
	flag.Var(flagTag, "t", "Filter by tag (repeatable)")
	flag.Var(flagTag, "tag", "Filter by tag (repeatable)")

	flag.Var(flagUser, "user", "SSH user (for ssh command)")
}

func main() {
	flag.Usage = func() {
		cmd := ""
		args := os.Args[1:]
		for _, a := range args {
			if !strings.HasPrefix(a, "-") {
				cmd = a
				break
			}
		}

		fmt.Fprintf(os.Stderr, "Usage: %s [options] <command> [command options]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Commands:\n")
		fmt.Fprintf(os.Stderr, "  servers      List servers with optional filters\n")
		fmt.Fprintf(os.Stderr, "  ssh          SSH to server by alias\n")
		fmt.Fprintf(os.Stderr, "  completion   Generate shell completion script\n")
		fmt.Fprintf(os.Stderr, "  aliases      List all server aliases\n")
		fmt.Fprintf(os.Stderr, "  locations    List all available locations\n")
		fmt.Fprintf(os.Stderr, "  regions      List all available regions\n")
		fmt.Fprintf(os.Stderr, "  power        List all power statuses\n")
		fmt.Fprintf(os.Stderr, "  status       List all server statuses\n")
		fmt.Fprintf(os.Stderr, "  fields       List all queryable server fields\n\n")

		switch cmd {
		case "servers":
			fmt.Fprintf(os.Stderr, "Options for servers:\n")
			fmt.Fprintf(os.Stderr, "  -o, --output <format>   Output format: json, table, csv (default json)\n")
			fmt.Fprintf(os.Stderr, "  -ow, --output-wide      Show more fields in table/csv\n")
			fmt.Fprintf(os.Stderr, "  -q, --query <field>     Output specific field(s) (repeatable)\n")
			fmt.Fprintf(os.Stderr, "\nFilters (shorthand, long form):\n")
			fmt.Fprintf(os.Stderr, "  -n, --name <val>      Filter by name (repeatable)\n")
			fmt.Fprintf(os.Stderr, "  -a, --alias <val>     Filter by alias (repeatable)\n")
			fmt.Fprintf(os.Stderr, "  -l, --location <val>  Filter by location (repeatable)\n")
			fmt.Fprintf(os.Stderr, "  -r, --region <val>    Filter by region (repeatable)\n")
			fmt.Fprintf(os.Stderr, "  -s, --status <val>    Filter by status (repeatable)\n")
			fmt.Fprintf(os.Stderr, "  -p, --power <val>     Filter by power status (repeatable)\n")
			fmt.Fprintf(os.Stderr, "  -t, --tag <val>       Filter by tag (repeatable)\n")
		case "ssh":
			fmt.Fprintf(os.Stderr, "Options for ssh:\n")
			fmt.Fprintf(os.Stderr, "  --user <name>           SSH user (default: root on linux, admin on windows)\n")
		case "completion":
			fmt.Fprintf(os.Stderr, "Options for completion:\n")
			fmt.Fprintf(os.Stderr, "  (shell name required: bash, zsh, fish)\n")
		default:
		}
	}

	flag.Parse()

	if flag.NArg() < 1 {
		flag.Usage()
		os.Exit(1)
	}

	cmd := flag.Arg(0)
	cmdArgs := flag.Args()[1:]

	if len(cmdArgs) > 0 {
		if err := flag.CommandLine.Parse(cmdArgs); err != nil {
			fmt.Fprintln(os.Stderr, "Error parsing flags:", err)
			os.Exit(1)
		}
	}

	opts := server.Options{
		Name:     *flagName,
		Alias:    *flagAlias,
		Location: *flagLocation,
		Region:   *flagRegion,
		Status:   *flagStatus,
		Power:    *flagPower,
		Tag:      *flagTag,
	}

	if err := run(cmd, cmdArgs, opts); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}

func validateCmdFlags(cmd string) error {
	if cmd != "servers" {
		hasFilter := len(*flagName) > 0 || len(*flagAlias) > 0 || len(*flagLocation) > 0 ||
			len(*flagRegion) > 0 || len(*flagStatus) > 0 || len(*flagPower) > 0 ||
			len(*flagTag) > 0 || len(*queryFields) > 0 || outputFormatFlag != "" ||
			outputWide || flag.Lookup("output-wide").Value.String() == "true"
		if hasFilter {
			return errors.New("output and filter flags are only valid for the servers command")
		}
	}
	if cmd != "ssh" && len(*flagUser) > 0 {
		return errors.New("flag --user is only valid for the ssh command")
	}
	return nil
}

func run(cmd string, args []string, opts server.Options) error {
	if err := validateCmdFlags(cmd); err != nil {
		return err
	}

	cfg, err := config.Load()
	if err != nil {
		return err
	}
	wide := outputWide || flag.Lookup("output-wide").Value.String() == "true"
	sshUser := ""
	if len(*flagUser) > 0 {
		sshUser = (*flagUser)[0]
	}

	output := "json"
	if cfg.Output != "" {
		output = cfg.Output
	}
	if outputFormatFlag != "" {
		output = outputFormatFlag
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	switch cmd {
	case "servers":
		if len(*queryFields) > 0 {
			opts.Fields = *queryFields
		} else if output == "table" || output == "csv" {
			if wide {
				opts.Fields = []string{"Name", "Alias", "Status", "Power", "Location", "IP", "OS", "CPU", "Memory", "Storage", "Price"}
			} else {
				opts.Fields = []string{"Name", "Alias", "Status", "Location", "IP"}
			}
		}
		return runShow(ctx, output, wide, opts, cfg)
	case "ssh":
		if len(args) < 1 {
			return errors.New("usage: dp ssh <alias> [ssh flags...]")
		}
		return runSSH(ctx, opts, sshUser, args, cfg)
	case "completion":
		if len(args) < 1 {
			return errors.New("usage: dp completion <bash|zsh|fish>")
		}
		return completion.Generate(completion.Shell(args[0]))
	case "aliases":
		return runFilter(ctx, "aliases", cfg)
	case "locations":
		return runFilter(ctx, "locations", cfg)
	case "regions":
		return runFilter(ctx, "regions", cfg)
	case "power":
		return runFilter(ctx, "power", cfg)
	case "status":
		return runFilter(ctx, "status", cfg)
	case "fields":
		runFields()
		return nil
	case "version", "-V", "--version":
		fmt.Println(getVersion())
		return nil
	default:
		return fmt.Errorf("unknown command: %s", cmd)
	}
}

func getVersion() string {
	revision := "unknown"

	info, ok := debug.ReadBuildInfo()
	if ok {
		for _, setting := range info.Settings {
			if setting.Key == "vcs.revision" {
				revision = setting.Value
				break
			}
		}
	}

	if version == "" {
		return revision
	}

	if revision == version {
		return version
	}

	return fmt.Sprintf("%s (%s)", version, revision)
}

func printTable(servers []server.Server, wide bool, queryFields []string) string {
	if len(servers) == 0 {
		return "No servers found"
	}

	type col struct {
		name        string
		displayName string
		width       int
	}

	var cols []col
	switch {
	case len(queryFields) > 0:
		for _, f := range queryFields {
			cols = append(cols, col{name: f, displayName: strings.ToUpper(f), width: 0})
		}
	case wide:
		cols = []col{
			{name: "Name", displayName: "NAME", width: 0},
			{name: "Alias", displayName: "ALIAS", width: 0},
			{name: "Status", displayName: "STATUS", width: 0},
			{name: "Power", displayName: "POWER", width: 0},
			{name: "Location", displayName: "LOCATION", width: 0},
			{name: "IP", displayName: "IP", width: 0},
			{name: "OS", displayName: "OS", width: 0},
			{name: "CPU", displayName: "CPU", width: 0},
			{name: "Memory", displayName: "MEMORY", width: 0},
			{name: "Storage", displayName: "STORAGE", width: 0},
			{name: "Price", displayName: "PRICE", width: 0},
		}
	default:
		cols = []col{
			{name: "Name", displayName: "NAME", width: 0},
			{name: "Alias", displayName: "ALIAS", width: 0},
			{name: "Status", displayName: "STATUS", width: 0},
			{name: "Location", displayName: "LOCATION", width: 0},
			{name: "IP", displayName: "IP", width: 0},
		}
	}

	rowValues := make([][]string, len(servers))
	for si, s := range servers {
		rowValues[si] = make([]string, len(cols))
		for i, c := range cols {
			val := getFieldValue(s, c.name)
			rowValues[si][i] = val
			cols[i].width = max(cols[i].width, len(val))
		}
	}

	var formatBuilder strings.Builder
	for i := range cols {
		cols[i].width = max(cols[i].width, len(cols[i].displayName)) + 1
		fmt.Fprintf(&formatBuilder, "%%-%ds", cols[i].width)
	}
	format := formatBuilder.String() + "\n"

	// Pre-allocate a reusable []any slice to avoid per-row allocations.
	fmtArgs := make([]any, len(cols))

	var b strings.Builder

	header := make([]string, len(cols))
	dashes := make([]string, len(cols))
	for i, c := range cols {
		header[i] = c.displayName
		dashes[i] = strings.Repeat("-", c.width-1)
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

var queryableFields = []string{
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

func runFields() {
	for _, f := range queryableFields {
		fmt.Println(f)
	}
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

func printCSV(w *csv.Writer, servers []server.Server, wide bool, queryFields []string) error {
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

func getClient(cfg *config.Config) (api.Querier, error) {
	apiKey := cfg.APIKey
	if apiKey == "" {
		return nil, api.ErrMissingAPIKey
	}
	client, err := api.NewClient(apiKey)
	if err != nil {
		return nil, err
	}
	if cfg.APIURL != "" {
		client.SetBaseURL(cfg.APIURL)
	}
	return client, nil
}

func runShow(ctx context.Context, outputFormat string, outputWide bool, opts server.Options, cfg *config.Config) error {
	client, err := getClient(cfg)
	if err != nil {
		return err
	}

	servers, err := server.List(ctx, client, opts.ToOpts()...)
	if err != nil {
		return err
	}

	switch outputFormat {
	case "json":
		var encoded []byte
		var err error
		if len(opts.Fields) > 0 {
			output := make([]map[string]any, len(servers))
			for i, s := range servers {
				m := make(map[string]any)
				for _, f := range opts.Fields {
					m[f] = getFieldValue(s, f)
				}
				output[i] = m
			}
			encoded, err = json.MarshalIndent(output, "", "  ")
		} else {
			encoded, err = json.MarshalIndent(servers, "", "  ")
		}
		if err != nil {
			return fmt.Errorf("marshaling JSON: %w", err)
		}
		fmt.Println(string(encoded))
	case "table":
		fmt.Println(printTable(servers, outputWide, opts.Fields))
	case "csv":
		w := csv.NewWriter(os.Stdout)
		if err := printCSV(w, servers, outputWide, opts.Fields); err != nil {
			return fmt.Errorf("writing CSV: %w", err)
		}
	case "raw":
		if len(opts.Fields) == 0 {
			return errors.New("raw output requires at least one query field")
		}
		for _, s := range servers {
			var values []string
			for _, f := range opts.Fields {
				values = append(values, getFieldValue(s, f))
			}
			fmt.Println(strings.Join(values, " "))
		}
	default:
		return fmt.Errorf("unknown output format: %s (use json, table, csv, or raw)", outputFormat)
	}

	return nil
}

func runSSH(ctx context.Context, opts server.Options, sshUser string, args []string, cfg *config.Config) error {
	client, err := getClient(cfg)
	if err != nil {
		return err
	}

	if len(args) > 0 {
		alias := args[0]
		if strings.Contains(alias, "@") {
			alias = strings.SplitN(alias, "@", 2)[1]
		}
		if alias != "" {
			opts.Alias = append(opts.Alias, alias)
		}
	}

	opts.Fields = []string{"Name", "Alias", "IP", "OperatingSystem"}

	servers, err := server.List(ctx, client, opts.ToOpts()...)
	if err != nil {
		return err
	}

	return ssh.Run(ctx, servers, sshUser, args)
}

func runFilter(ctx context.Context, filterType string, cfg *config.Config) error {
	client, err := getClient(cfg)
	if err != nil {
		return err
	}

	var filter interface {
		Get(context.Context) ([]string, error)
	}

	switch filterType {
	case "aliases":
		filter = filters.NewAliases(client)
	case "locations":
		filter = filters.NewLocations(client)
	case "regions":
		filter = filters.NewRegions(client)
	case "power":
		filter = filters.NewPower()
	case "status":
		filter = filters.NewStatus()
	default:
		return fmt.Errorf("unknown filter type: %s", filterType)
	}

	var list []string
	if filterType == "power" || filterType == "status" {
		list, err = filter.Get(ctx)
	} else {
		var cacheDuration time.Duration
		switch filterType {
		case "aliases":
			cacheDuration = cfg.AliasesCache
		case "locations":
			cacheDuration = cfg.LocationsCache
		case "regions":
			cacheDuration = cfg.RegionsCache
		}

		c, err := cache.New[[]string](filterType, cacheDuration, "")
		if err != nil {
			return err
		}
		if !c.Get(&list) {
			list, err = filter.Get(ctx)
			if err != nil {
				return err
			}
			if err = c.Set(list, 0); err != nil {
				return err
			}
		}
	}
	if err != nil {
		return err
	}

	for _, item := range list {
		fmt.Println(item)
	}
	return nil
}
