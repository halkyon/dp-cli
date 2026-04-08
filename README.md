# dp-cli

CLI tool for interacting with DataPacket's GraphQL API to list and SSH into servers.

## Requirements

- Go 1.25+
- DataPacket readonly API key (generate this at https://app.datapacket.com/settings/security)

## Building

```bash
make build
```

## Configuration

```bash
export DATAPACKET_API_KEY="your-api-key"
```

Tip: put a space in front of `export ...` to ensure it doesn't get written to your shell history!

## Commands

### show - Show server info

```bash
# List all servers
dp show

# Filter by regex (matches name or alias)
dp show "dp-prod-edge-mia-.+"
```

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

