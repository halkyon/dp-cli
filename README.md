# dp

CLI tool for interacting with DataPacket's GraphQL API to list and SSH into servers.

## Requirements

- Go 1.25+ (if building from source)
- DataPacket readonly API key (generate this at https://app.datapacket.com/settings/security)

## Installation

Download a pre-built binary from [releases](https://github.com/halkyon/dp-cli/releases) or build from source:

```bash
make build
```

## Configuration

The API key is loaded in the following order of priority:

1. `DATAPACKET_API_KEY` environment variable
2. `api_key` in config file

The output format is loaded in the following order of priority:

1. `-o` / `--output` flag
2. `output` in config file
3. default `json`

Cache durations and API URL can be configured in the config file:

```ini
api_key = your-api-key
output = table
api_url = https://api.example.com/graphql
aliases_cache = 24h
locations_cache = 168h
regions_cache = 168h
```

## Commands

### show - List servers

```bash
# List all servers (sorted by price)
dp show

# Output formats: json (default), table, csv
dp -o table show
dp -o csv show

# Combine with filters (flags work before or after command)
dp -o csv show --alias=dp-select-storage-slo-4
dp -o table show --location=London

# Show wide table/csv with more fields
dp -o table -ow show
dp -o csv --output-wide show

# Filter by name (repeatable)
dp show --name DP-12345 --name DP-67890

# Filter by alias (repeatable)
dp show --alias my-server --alias prod-server

# Filter by location (repeatable)
dp show --location Amsterdam --location Berlin

# Filter by region (repeatable)
dp show --region EU --region NA

# Filter by status (repeatable)
dp show --status ACTIVE --status MAINTENANCE

# Filter by power status (repeatable)
dp show --power ON

# Filter by tag (repeatable)
dp show --tag env=prod --tag team=backend

# Output specific fields (comma-separated)
dp show -q Name,Alias,IP

# Fuzzy filter using jq
dp show | jq '[.[] | select(.alias) | select(.alias | test("dp-prod-edge-fra01-[0-9]"))]'
```

### ssh - SSH to server

```bash
# SSH using default user (based on OS)
dp ssh my-server

# SSH as different user
dp ssh root@my-server

# SSH with verbose output
dp -v ssh my-server

# SSH with specific user
dp ssh --user=admin my-server
```

## Global Options

```bash
-o, --output <format>  Output format: json, table, csv (default "json")
    --output-wide      Show more fields in table/csv output
-q, --query <fields>   Output specific field(s) (comma-separated)
-v, --verbose          Print verbose information
```

**Note:** Flags can come before or after the command (e.g., `dp -o csv show --name=foo` or `dp show -o csv`).

### ssh - SSH to server

```bash
# SSH using default user (based on OS)
dp ssh my-server

# SSH as different user
dp ssh root@my-server

# SSH with verbose output
dp -v ssh my-server

# SSH with specific user (flag after command)
dp ssh --user=admin my-server
```

## Shell completion setup

After sourcing the completion script (see below), tab completion will work for:
- Commands (show, ssh, completion)
- Alias names when using `show` or `ssh`

### zsh

```bash
dp completion zsh > ~/.zsh/completions/_dp
```

### bash

```bash
dp completion bash > /etc/bash_completion.d/dp
```

## SDK usage

The dp packages can be imported for programmatic access to the DataPacket API.

### API client

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/halkyon/dp/api"
    "github.com/halkyon/dp/config"
    "github.com/halkyon/dp/server"
)

func main() {
    ctx := context.Background()

    apiKey, err := config.GetAPIKey()
    if err != nil {
        log.Fatal(err)
    }
    if apiKey == "" {
        log.Fatal(api.ErrMissingAPIKey)
    }

    client, err := api.NewClient(apiKey)
    if err != nil {
        log.Fatal(err)
    }

    servers, err := server.List(ctx, client)
    if err != nil {
        log.Fatal(err)
    }

    for _, s := range servers {
        fmt.Printf("%s (%s) - %s\n", s.Name, s.Alias, s.IP)
    }
}
```

### List with filters

```go
servers, err := server.List(ctx, client,
    server.WithName("DP-12345", "DP-67890"),
    server.WithAlias("my-server", "prod-server"),
    server.WithLocation("Amsterdam", "Berlin"),
    server.WithRegion("EU", "NA"),
    server.WithStatus("ACTIVE", "MAINTENANCE"),
    server.WithPower("ON"),
    server.WithTag("env=prod", "team=backend"),
)
```

