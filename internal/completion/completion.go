package completion

import (
	"context"
	"fmt"

	"dp/internal/aliases"
	"dp/internal/api"
	"dp/internal/config"
)

type Shell string

const (
	ShellBash Shell = "bash"
	ShellZsh  Shell = "zsh"
	ShellFish Shell = "fish"
)

func Generate(shell Shell) error {
	switch shell {
	case ShellBash:
		fmt.Print(bashCompletionScript)
	case ShellZsh:
		fmt.Print(zshCompletionScript)
	case ShellFish:
		fmt.Print(fishCompletionScript)
	default:
		return fmt.Errorf("unsupported shell: %s (use bash, zsh, or fish)", shell)
	}

	return nil
}

func ListAliases(ctx context.Context) error {
	apiKey, err := config.GetAPIKey()
	if err != nil {
		return err
	}
	if apiKey == "" {
		return api.ErrMissingAPIKey
	}

	client, err := api.NewClient(apiKey)
	if err != nil {
		return err
	}

	cache := aliases.New(client)
	list, err := cache.Get(ctx)
	if err != nil {
		return err
	}

	for _, alias := range list {
		fmt.Println(alias)
	}

	return nil
}

var bashCompletionScript = `# bash completion for dp
_dp_servers() {
    local cur prev words cword
    _init_completion || return

    if [[ $cword -eq 1 ]]; then
        COMPREPLY=($(compgen -W "show ssh completion aliases" -- "$cur"))
    elif [[ $cword -eq 2 ]]; then
        case "${words[1]}" in
            ssh|show)
                local -a aliases
                if aliases=($(dp aliases 2>/dev/null)); then
                    COMPREPLY=($(compgen -W "${aliases[*]}" -- "$cur"))
                fi
                ;;
            completion)
                COMPREPLY=($(compgen -W "bash zsh fish" -- "$cur"))
                ;;
        esac
    fi
}

complete -F _dp_servers dp
`

var zshCompletionScript = `# zsh completion for dp
_dp_servers() {
    local -a commands
    commands=(
        "show:List servers with optional regex filter"
        "ssh:SSH to server by alias"
        "completion:Generate completion script"
        "aliases:List all server aliases"
    )

    if (( CURRENT == 2 )); then
        _describe "command" commands
    elif (( CURRENT == 3 )); then
        case "${words[2]}" in
            ssh|show)
                local -a aliases
                aliases=($(dp aliases 2>/dev/null))
                _describe "alias" aliases
                ;;
            completion)
                _describe "shell" "bash bash completion" "zsh zsh completion" "fish fish completion"
                ;;
        esac
    fi
}

compdef _dp_servers dp
`

var fishCompletionScript = `# fish completion for dp
function __fish_dp_servers_needs_command
    set -l cmd (commandline -opc)
    if test (count $cmd) -eq 1
        return 0
    end
    return 1
end

function __fish_dp_servers_using_command
    set -l cmd (commandline -opc)
    if test (count $cmd) -gt 1
        if test "$cmd[2]" = "$argv[1]"
            return 0
        end
    end
    return 1
end

function __fish_dp_servers_get_aliases
    dp aliases 2>/dev/null
end

complete -c dp -f -n '__fish_dp_servers_needs_command'
complete -c dp -a 'show' -d 'List servers with optional regex filter'
complete -c dp -a 'ssh' -d 'SSH to server by alias'
complete -c dp -a 'completion' -d 'Generate completion script'
complete -c dp -a 'aliases' -d 'List all server aliases'

complete -c dp -f -n '__fish_dp_servers_using_command show' -a '(__fish_dp_servers_get_aliases)'
complete -c dp -f -n '__fish_dp_servers_using_command ssh' -a '(__fish_dp_servers_get_aliases)'
`
