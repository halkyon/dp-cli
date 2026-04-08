# dp-cli

CLI tool for interacting with DataPacket's GraphQL API to list and SSH into servers.

## Requirements

- Go 1.25+
- DataPacket readonly API key (generate this at https://app.datapacket.com/settings/security)

## Installation

Download a pre-built binary from [releases](https://github.com/halkyon/dp-cli/releases) or build from source:

```bash
make build
```

## Configuration

The API key is loaded in the following order of priority:

1. `DATAPACKET_API_KEY` environment variable
2. `~/.config/dp/credentials` config file

Example config file:

```ini
api_key = your-api-key
```

## Commands

### show - Show server info

```bash
# List all servers (sorted by price)
dp show

# Filter by regex (matches name or alias)
dp show "dp-prod-edge-mia-.+"
```

Output includes: name, alias, IP address, price, currency, status, power status, and operating system.

### ssh - SSH to server

```bash
# SSH using ubuntu user (default)
dp ssh "dp-prod-edge-mia1-11"

# SSH as different user
dp ssh "root@dp-prod-edge-mia1-11"
```

### aliases - List all server aliases

```bash
dp aliases
```

## Shell completion setup

After sourcing the completion script (see below), tab completion will work for:
- Commands (show, ssh, completion, aliases)
- Alias names when using `show` or `ssh`

### zsh

```bash
dp completion zsh > ~/.zsh/completions/_dp
```

### bash

```bash
dp completion bash > /etc/bash_completion.d/dp
```

