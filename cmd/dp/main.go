package main

import (
	"context"
	"encoding/csv"
	"encoding/json"
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
	flag.Bool("verbose", false, "Print verbose information")
	flag.StringVar(&outputFormatFlag, "o", "", "Output format (shorthand): json, table, csv")
	flag.StringVar(&outputFormatFlag, "output", "", "Output format: json, table, csv")
	flag.BoolVar(&outputWide, "ow", false, "Show more fields (shorthand)")
	flag.Bool("output-wide", false, "Show more fields in table/csv output")
	flag.Var(queryFields, "query", "Output specific field(s) (repeatable)")

	flag.Var(flagName, "name", "Filter by name (repeatable)")
	flag.Var(flagAlias, "alias", "Filter by alias (repeatable)")
	flag.Var(flagLocation, "location", "Filter by location (repeatable)")
	flag.Var(flagRegion, "region", "Filter by region (repeatable)")
	flag.Var(flagStatus, "status", "Filter by status (repeatable)")
	flag.Var(flagPower, "power", "Filter by power status (repeatable)")
	flag.Var(flagTag, "tag", "Filter by tag (repeatable)")

	flag.Var(flagUser, "user", "SSH user (for ssh command)")
}

func main() {
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
	case "show":
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
			return fmt.Errorf("usage: dp ssh <alias> [ssh flags...]")
		}
		return runSSH(ctx, opts, sshUser, sshArgs, verbose)
	case "completion":
		if len(args) < 1 {
			return fmt.Errorf("usage: dp completion <bash|zsh|fish>")
		}
		return completion.Generate(completion.Shell(args[0]))
	case "aliases":
		return completion.ListAliases(ctx)
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
		fieldSet[f] = true
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
		if fields[field.Name] {
			val.Field(i).Set(reflect.ValueOf(s).Field(i))
		}
	}
	return result
}

func printTable(servers []server.Server, wide bool) string {
	if len(servers) == 0 {
		return "No servers found"
	}

	type col struct {
		name  string
		width int
	}

	var cols []col
	if wide {
		cols = []col{{"NAME", 0}, {"ALIAS", 0}, {"STATUS", 0}, {"LOCATION", 0}, {"IP", 0}, {"OS", 0}, {"CPU", 0}, {"MEMORY", 0}, {"PRICE", 0}}
	} else {
		cols = []col{{"NAME", 0}, {"ALIAS", 0}, {"STATUS", 0}, {"LOCATION", 0}, {"IP", 0}}
	}

	for _, s := range servers {
		cols[0].width = max(cols[0].width, len(s.Name))
		cols[1].width = max(cols[1].width, len(s.Alias))
		cols[2].width = max(cols[2].width, len(s.Status))
		cols[3].width = max(cols[3].width, len(s.Location))
		cols[4].width = max(cols[4].width, len(s.IP))
		if wide {
			cols[5].width = max(cols[5].width, len(s.OperatingSystem))
			cols[6].width = max(cols[6].width, len(s.CPU))
			cols[7].width = max(cols[7].width, len(s.Memory))
			cols[8].width = max(cols[8].width, len(fmt.Sprintf("%.2f", s.Price)))
		}
	}

	for i := range cols {
		cols[i].width = max(cols[i].width, len(cols[i].name)) + 1
	}

	format := ""
	for _, c := range cols {
		format += fmt.Sprintf("%%-%ds", c.width)
	}
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
		header[i] = c.name
	}
	b.WriteString(fmt.Sprintf(format, toAny(header)...))

	dashes := make([]string, len(cols))
	for i, c := range cols {
		dashes[i] = strings.Repeat("-", c.width-1)
	}
	b.WriteString(fmt.Sprintf(format, toAny(dashes)...))

	for _, s := range servers {
		var values []string
		values = append(values, s.Name, s.Alias, s.Status, s.Location, s.IP)
		if wide {
			values = append(values, s.OperatingSystem, s.CPU, s.Memory, fmt.Sprintf("%.2f", s.Price))
		}
		b.WriteString(fmt.Sprintf(format, toAny(values)...))
	}

	return b.String()
}

func getFieldValue(s server.Server, field string) string {
	switch field {
	case "Name":
		return s.Name
	case "Alias":
		return s.Alias
	case "Status":
		return s.Status
	case "Location":
		return s.Location
	case "IP":
		return s.IP
	case "OS":
		return s.OperatingSystem
	case "CPU":
		return s.CPU
	case "Memory":
		return s.Memory
	case "Price":
		return fmt.Sprintf("%.2f", s.Price)
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
		fields = []string{"Name", "Alias", "Status", "Location", "IP", "OS", "CPU", "Memory", "Price"}
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

func getClient() (*api.Client, error) {
	apiKey, err := config.GetAPIKey()
	if err != nil {
		return nil, err
	}
	if apiKey == "" {
		return nil, fmt.Errorf("API key not found. Set DATAPACKET_API_KEY or api_key in ~/.config/dp/credentials")
	}
	client, err := api.NewClient(apiKey)
	if err != nil {
		return nil, err
	}
	return client, nil
}

func runShow(ctx context.Context, outputFormat string, outputWide bool, queryFields []string, opts server.Options) error {
	client, err := getClient()
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
		output, err := json.MarshalIndent(servers, "", "  ")
		if err != nil {
			return fmt.Errorf("marshaling JSON: %w", err)
		}
		fmt.Println(string(output))
	case "table":
		fmt.Println(printTable(servers, outputWide))
	case "csv":
		w := csv.NewWriter(os.Stdout)
		fmt.Println(printCSV(w, servers, outputWide, queryFields))
	default:
		return fmt.Errorf("unknown output format: %s (use json, table, or csv)", outputFormat)
	}

	return nil
}

func runSSH(ctx context.Context, opts server.Options, sshUser string, args []string, verbose bool) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	servers, err := server.List(ctx, client, opts.ToOpts()...)
	if err != nil {
		return err
	}

	return ssh.Run(ctx, servers, sshUser, args, verbose)
}
