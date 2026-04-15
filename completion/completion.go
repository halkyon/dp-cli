package completion

import (
	"fmt"
	"os"
	"strings"
)

type Shell string

const (
	ShellBash Shell = "bash"
	ShellZsh  Shell = "zsh"
	ShellFish Shell = "fish"
)

func Generate(shell Shell) error {
	binaryName := getBinaryName()

	switch shell {
	case ShellBash:
		fmt.Print(renderBashCompletion(binaryName))
	case ShellZsh:
		fmt.Print(renderZshCompletion(binaryName))
	case ShellFish:
		fmt.Print(renderFishCompletion(binaryName))
	default:
		return fmt.Errorf("unsupported shell: %s (use bash, zsh, or fish)", shell)
	}

	return nil
}

func getBinaryName() string {
	name := os.Args[0]
	if idx := strings.LastIndex(name, "/"); idx >= 0 {
		name = name[idx+1:]
	}
	return name
}

func renderBashCompletion(name string) string {
	return `# bash completion for ` + name + `
_dp() {
    local cur prev words cword
    _init_completion || return

    if [[ $cword -eq 1 ]]; then
        COMPREPLY=($(compgen -W "show ssh completion aliases locations regions power status" -- "$cur"))
    else
        local prev_word="$prev"
        if [[ "$prev" == --alias=* ]]; then
            prev_word="--alias"
        elif [[ "$prev" == --location=* ]]; then
            prev_word="--location"
        elif [[ "$prev" == --region=* ]]; then
            prev_word="--region"
        elif [[ "$prev" == --power=* ]]; then
            prev_word="--power"
        elif [[ "$prev" == --status=* ]]; then
            prev_word="--status"
        elif [[ "$prev" == --alias ]]; then
            prev_word="--alias"
        elif [[ "$prev" == --location ]]; then
            prev_word="--location"
        elif [[ "$prev" == --region ]]; then
            prev_word="--region"
        elif [[ "$prev" == --power ]]; then
            prev_word="--power"
        elif [[ "$prev" == --status ]]; then
            prev_word="--status"
        fi

        case "$prev_word" in
            --alias)
                local -a aliases
                if aliases=($(` + name + ` aliases 2>/dev/null)); then
                    local completion_word="$cur"
                    if [[ "$cur" == *@* ]]; then
                        completion_word="${cur#*@}"
                    fi
                    COMPREPLY=($(compgen -W "${aliases[*]}" -- "$completion_word"))
                fi
                ;;
            --location)
                local -a locations
                if locations=($(` + name + ` locations 2>/dev/null)); then
                    COMPREPLY=($(compgen -W "${locations[*]}" -- "$cur"))
                fi
                ;;
            --region)
                local -a regions
                if regions=($(` + name + ` regions 2>/dev/null)); then
                    COMPREPLY=($(compgen -W "${regions[*]}" -- "$cur"))
                fi
                ;;
            --power)
                local -a power_statuses
                if power_statuses=($(` + name + ` power 2>/dev/null)); then
                    COMPREPLY=($(compgen -W "${power_statuses[*]}" -- "$cur"))
                fi
                ;;
            --status)
                local -a server_statuses
                if server_statuses=($(` + name + ` status 2>/dev/null)); then
                    COMPREPLY=($(compgen -W "${server_statuses[*]}" -- "$cur"))
                fi
                ;;
                ;;
        esac
    fi
}

complete -F _dp ` + name + `
`
}

func renderZshCompletion(name string) string {
	return `# zsh completion for ` + name + `
_dp() {
    local -a commands
    commands=(
        "show:List servers with optional regex filter"
        "ssh:SSH to server by alias"
        "completion:Generate completion script"
        "aliases:List all server aliases"
        "locations:List all available locations"
        "regions:List all available regions"
        "power:List all power statuses"
        "status:List all server statuses"
    )

    if (( CURRENT == 2 )); then
        _describe "command" commands
    else
        local prev_word="${words[CURRENT-1]}"
        local cur_word="${words[CURRENT]}"

        if [[ "$prev_word" == --alias=* ]]; then
            prev_word="--alias"
        elif [[ "$prev_word" == --location=* ]]; then
            prev_word="--location"
        elif [[ "$prev_word" == --region=* ]]; then
            prev_word="--region"
        elif [[ "$prev_word" == --power=* ]]; then
            prev_word="--power"
        elif [[ "$prev_word" == --status=* ]]; then
            prev_word="--status"
        fi

        case "$prev_word" in
            --alias)
                local -a aliases
                aliases=($(` + name + ` aliases 2>/dev/null))
                if [[ "$cur_word" == *@* ]]; then
                    cur_word="${cur_word#*@}"
                fi
                compstate[word]="$cur_word"
                _describe "alias" aliases
                ;;
            --location)
                local -a locations
                locations=($(` + name + ` locations 2>/dev/null))
                _describe "location" locations
                ;;
            --region)
                local -a regions
                regions=($(` + name + ` regions 2>/dev/null))
                _describe "region" regions
                ;;
            --power)
                local -a power_statuses
                power_statuses=($(` + name + ` power 2>/dev/null))
                _describe "power" power_statuses
                ;;
            --status)
                local -a server_statuses
                server_statuses=($(` + name + ` status 2>/dev/null))
                _describe "status" server_statuses
                ;;
            *)
                local -a aliases
                aliases=($(` + name + ` aliases 2>/dev/null))
                if [[ "$cur_word" == *@* ]]; then
                    cur_word="${cur_word#*@}"
                fi
                compstate[word]="$cur_word"
                _describe "alias" aliases
                ;;
        esac
    fi
}

compdef _dp ` + name + `
`
}

func renderFishCompletion(name string) string {
	return `# fish completion for ` + name + `
function __fish_dp_needs_command
    set -l cmd (commandline -opc)
    if test (count $cmd) -eq 1
        return 0
    end
    return 1
end

function __fish_dp_get_aliases
    ` + name + ` aliases 2>/dev/null
end

function __fish_dp_get_locations
    ` + name + ` locations 2>/dev/null
end

function __fish_dp_get_regions
    ` + name + ` regions 2>/dev/null
end

function __fish_dp_complete_alias
    set -l cmd (commandline -opc)
    set -l cur (commandline -ct)
    set -l word_to_complete "$cur"
    if string match -q '*@*' "$cur"
        set word_to_complete (string split '@' "$cur")[-1]
    end
    __fish_dp_get_aliases | string match "$word_to_complete*"
end

function __fish_dp_complete_filter
    set -l cmd (commandline -opc)
    set -l prev (commandline -op)
    set -l cur (commandline -ct)
    set -l filter_type ""
    set -l word_to_complete "$cur"

    if string match -q '--alias=*' "$prev"
        set filter_type "alias"
        set word_to_complete (string replace '--alias=' '' "$cur")
    else if string match -q '--location=*' "$prev"
        set filter_type "location"
        set word_to_complete (string replace '--location=' '' "$cur")
    else if string match -q '--region=*' "$prev"
        set filter_type "region"
        set word_to_complete (string replace '--region=' '' "$cur")
    else if string match -q '--power=*' "$prev"
        set filter_type "power"
        set word_to_complete (string replace '--power=' '' "$cur")
    else if string match -q '--status=*' "$prev"
        set filter_type "status"
        set word_to_complete (string replace '--status=' '' "$cur")
    else if test "$cur" = "--alias"
        set filter_type "alias"
    else if test "$cur" = "--location"
        set filter_type "location"
    else if test "$cur" = "--region"
        set filter_type "region"
    else if test "$cur" = "--power"
        set filter_type "power"
    else if test "$cur" = "--status"
        set filter_type "status"
    end

    switch "$filter_type"
        case "alias"
            __fish_dp_get_aliases | string match "$word_to_complete*"
        case "location"
            __fish_dp_get_locations | string match "$word_to_complete*"
        case "region"
            __fish_dp_get_regions | string match "$word_to_complete*"
        case "power"
            ` + name + ` power 2>/dev/null | string match "$word_to_complete*"
        case "status"
            ` + name + ` status 2>/dev/null | string match "$word_to_complete*"
    end
end

complete -c ` + name + ` -f -n '__fish_dp_needs_command'
complete -c ` + name + ` -a 'show' -d 'List servers with optional regex filter'
complete -c ` + name + ` -a 'ssh' -d 'SSH to server by alias'
complete -c ` + name + ` -a 'completion' -d 'Generate completion script'
complete -c ` + name + ` -a 'aliases' -d 'List all server aliases'
complete -c ` + name + ` -a 'locations' -d 'List all available locations'
complete -c ` + name + ` -a 'regions' -d 'List all available regions'
complete -c ` + name + ` -a 'power' -d 'List all power statuses'
complete -c ` + name + ` -a 'status' -d 'List all server statuses'
complete -c ` + name + ` -f -a '(__fish_dp_complete_alias)'
complete -c ` + name + ` -f -a '(__fish_dp_complete_filter)'
`
}
