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
	"reflect"
	"runtime/debug"
	"strings"

	"github.com/halkyon/dp/api"
	"github.com/halkyon/dp/completion"
	"github.com/halkyon/dp/config"
	"github.com/halkyon/dp/filters"
	"github.com/halkyon/dp/server"
	"github.com/halkyon/dp/ssh"
	"github.com/halkyon/dp/testapi"
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
var flagTestAPI bool

type stringSlice []string

func (s *stringSlice) Set(val string) error {
	*s = append(*s, val)
	return nil
}

func (s *stringSlice) String() string {
	return strings.Join(*s, ",")
}

func init() {
	flag.Bool("verbose", false, "Print verbose information")
	flag.StringVar(&outputFormatFlag, "o", "", "Output format (shorthand): json, table, csv")
	flag.StringVar(&outputFormatFlag, "output", "", "Output format: json, table, csv")
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
	flag.BoolVar(&flagTestAPI, "test-api", false, "Use test API server")
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
		fmt.Fprintf(os.Stderr, "  status       List all server statuses\n\n")

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
			fmt.Fprintf(os.Stderr, "  -v, --verbose           Print verbose information\n")
			fmt.Fprintf(os.Stderr, "  --user <name>           SSH user (default: root on linux, admin on windows)\n")
		case "completion":
			fmt.Fprintf(os.Stderr, "Options for completion:\n")
			fmt.Fprintf(os.Stderr, "  (shell name required: bash, zsh, fish)\n")
		case "aliases", "locations", "regions", "power", "status":
			fmt.Fprintf(os.Stderr, "Options:\n")
			fmt.Fprintf(os.Stderr, "  --test-api             Use test API server\n")
		default:
			fmt.Fprintf(os.Stderr, "Options:\n")
			fmt.Fprintf(os.Stderr, "  --test-api             Use test API server\n")
			fmt.Fprintf(os.Stderr, "  -v, --verbose          Print verbose information\n")
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

func run(cmd string, args []string, opts server.Options) error {
	wide := outputWide || flag.Lookup("output-wide").Value.String() == "true"
	verbose := flag.Lookup("verbose").Value.String() == "true"
	sshUser := ""
	if len(*flagUser) > 0 {
		sshUser = (*flagUser)[0]
	}

	output := "json"
	if configOutput, err := config.GetOutput(); err == nil && configOutput != "" {
		output = configOutput
	}
	if outputFormatFlag != "" {
		output = outputFormatFlag
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	switch cmd {
	case "servers":
		return runShow(ctx, output, wide, *queryFields, opts)
	case "ssh":
		var sshArgs []string
		for _, arg := range args {
			if strings.HasPrefix(arg, "-") {
				continue
			}
			sshArgs = append(sshArgs, arg)
		}
		if len(sshArgs) < 1 {
			return errors.New("usage: dp ssh <alias> [ssh flags...]")
		}
		return runSSH(ctx, opts, sshUser, sshArgs, verbose)
	case "completion":
		if len(args) < 1 {
			return errors.New("usage: dp completion <bash|zsh|fish>")
		}
		return completion.Generate(completion.Shell(args[0]))
	case "aliases":
		return runFilter(ctx, "aliases")
	case "locations":
		return runFilter(ctx, "locations")
	case "regions":
		return runFilter(ctx, "regions")
	case "power":
		return runFilter(ctx, "power")
	case "status":
		return runFilter(ctx, "status")
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

func filterByFields(servers []server.Server, fields []string) []server.Server {
	fieldSet := make(map[string]bool)
	for _, f := range fields {
		fieldSet[strings.ToLower(f)] = true
	}

	var filtered []server.Server
	for _, s := range servers {
		filtered = append(filtered, filterServerFields(s, fieldSet))
	}
	return filtered
}

func filterServerFields(s server.Server, fields map[string]bool) server.Server {
	if len(fields) == 0 {
		return s
	}

	var result server.Server
	typ := reflect.TypeFor[server.Server]()
	val := reflect.ValueOf(&result).Elem()

	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		if fields[strings.ToLower(field.Name)] {
			val.Field(i).Set(reflect.ValueOf(s).Field(i))
		}
	}
	return result
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

	for _, s := range servers {
		for i, c := range cols {
			val := getFieldValue(s, c.name)
			cols[i].width = max(cols[i].width, len(val))
		}
	}

	for i := range cols {
		cols[i].width = max(cols[i].width, len(cols[i].displayName)) + 1
	}

	var formatBuilder strings.Builder
	for _, c := range cols {
		fmt.Fprintf(&formatBuilder, "%%-%ds", c.width)
	}
	format := formatBuilder.String()
	format += "\n"

	toAny := func(s []string) []any {
		result := make([]any, len(s))
		for i, v := range s {
			result[i] = v
		}
		return result
	}

	var b strings.Builder

	header := make([]string, len(cols))
	for i, c := range cols {
		header[i] = c.displayName
	}
	fmt.Fprintf(&b, format, toAny(header)...)

	dashes := make([]string, len(cols))
	for i, c := range cols {
		dashes[i] = strings.Repeat("-", c.width-1)
	}
	fmt.Fprintf(&b, format, toAny(dashes)...)

	for _, s := range servers {
		var values []string
		for _, c := range cols {
			val := getFieldValue(s, c.name)
			values = append(values, val)
		}
		fmt.Fprintf(&b, format, toAny(values)...)
	}

	return b.String()
}

func getFieldValue(s server.Server, field string) string {
	switch strings.ToLower(field) {
	case "name":
		return s.Name
	case "alias":
		return s.Alias
	case "status":
		return s.Status
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
	default:
		return ""
	}
}

func printCSV(w *csv.Writer, servers []server.Server, wide bool, queryFields []string) string {
	var fields []string
	switch {
	case len(queryFields) > 0:
		fields = queryFields
	case wide:
		fields = []string{"Name", "Alias", "Status", "Location", "IP", "OS", "CPU", "Memory", "Storage", "Price"}
	default:
		fields = []string{"Name", "Alias", "Status", "Location", "IP"}
	}

	if err := w.Write(fields); err != nil {
		return err.Error()
	}

	for _, s := range servers {
		var record []string
		for _, f := range fields {
			record = append(record, getFieldValue(s, f))
		}
		if err := w.Write(record); err != nil {
			return err.Error()
		}
	}

	w.Flush()
	if err := w.Error(); err != nil {
		return err.Error()
	}
	return ""
}

func getClient(ctx context.Context) (*api.Client, error) {
	apiKey, err := config.GetAPIKey()
	if err != nil {
		return nil, err
	}
	if apiKey == "" {
		apiKey = "test-key"
	}
	client, err := api.NewClient(apiKey)
	if err != nil {
		return nil, err
	}
	if flagTestAPI || config.GetTestAPI() {
		srv, err := testapi.NewServer()
		if err != nil {
			return nil, err
		}
		go func() {
			_ = srv.Run(ctx)
		}()
		url := "http://" + srv.Addr()
		client.SetBaseURL(url)
	} else if apiURL, err := config.GetAPIURL(); err == nil && apiURL != "" {
		client.SetBaseURL(apiURL)
	}
	return client, nil
}

func runShow(ctx context.Context, outputFormat string, outputWide bool, queryFields []string, opts server.Options) error {
	client, err := getClient(ctx)
	if err != nil {
		return err
	}

	servers, err := server.List(ctx, client, opts.ToOpts()...)
	if err != nil {
		return err
	}

	if len(queryFields) > 0 {
		servers = filterByFields(servers, queryFields)
	}

	switch outputFormat {
	case "json":
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
			return fmt.Errorf("marshaling JSON: %w", err)
		}
		fmt.Println(string(encoded))
	case "table":
		fmt.Println(printTable(servers, outputWide, queryFields))
	case "csv":
		w := csv.NewWriter(os.Stdout)
		fmt.Println(printCSV(w, servers, outputWide, queryFields))
	default:
		return fmt.Errorf("unknown output format: %s (use json, table, or csv)", outputFormat)
	}

	return nil
}

func runSSH(ctx context.Context, opts server.Options, sshUser string, args []string, verbose bool) error {
	client, err := getClient(ctx)
	if err != nil {
		return err
	}

	servers, err := server.List(ctx, client, opts.ToOpts()...)
	if err != nil {
		return err
	}

	return ssh.Run(ctx, servers, sshUser, args, verbose)
}

func runFilter(ctx context.Context, filterType string) error {
	client, err := getClient(ctx)
	if err != nil {
		return err
	}

	var filter interface {
		Get(context.Context) ([]string, error)
	}

	switch filterType {
	case "aliases":
		cache, err := filters.NewAliases(client, config.GetAliasesCache(), "")
		if err != nil {
			return err
		}
		filter = cache
	case "locations":
		cache, err := filters.NewLocations(client, config.GetLocationsCache(), "")
		if err != nil {
			return err
		}
		filter = cache
	case "regions":
		cache, err := filters.NewRegions(client, config.GetRegionsCache(), "")
		if err != nil {
			return err
		}
		filter = cache
	case "power":
		filter = filters.NewPower()
	case "status":
		filter = filters.NewStatus()
	default:
		return fmt.Errorf("unknown filter type: %s", filterType)
	}

	list, err := filter.Get(ctx)
	if err != nil {
		return err
	}

	for _, item := range list {
		fmt.Println(item)
	}
	return nil
}
