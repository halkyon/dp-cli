package completion

import (
	"fmt"
	"os"
	"strings"
	"text/template"
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

var bashCompletionTmpl = template.Must(template.New("bash").Parse(`# bash completion for {{.Name}}
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
                if aliases=($({{.Name}} aliases 2>/dev/null)); then
                    local completion_word="$cur"
                    if [[ "$cur" == *@* ]]; then
                        completion_word="${cur#*@}"
                    fi
                    COMPREPLY=($(compgen -W "${aliases[*]}" -- "$completion_word"))
                fi
                ;;
            --location)
                local -a locations
                while IFS= read -r location; do
                    locations+=("$location")
                done < <({{.Name}} locations 2>/dev/null)
                if [[ ${#locations[@]} -gt 0 ]]; then
                    local match="$cur"
                    match="${match#\"}"
                    match="${match#\'}"
                    for loc in "${locations[@]}"; do
                        if [[ "$loc" == "$match"* ]]; then
                            COMPREPLY+=("$loc")
                        fi
                    done
                fi
                ;;
            --region)
                local -a regions
                if regions=($({{.Name}} regions 2>/dev/null)); then
                    COMPREPLY=($(compgen -W "${regions[*]}" -- "$cur"))
                fi
                ;;
            --power)
                local -a power_statuses
                if power_statuses=($({{.Name}} power 2>/dev/null)); then
                    COMPREPLY=($(compgen -W "${power_statuses[*]}" -- "$cur"))
                fi
                ;;
            --status)
                local -a server_statuses
                if server_statuses=($({{.Name}} status 2>/dev/null)); then
                    COMPREPLY=($(compgen -W "${server_statuses[*]}" -- "$cur"))
                fi
                ;;
        esac
    fi
}

complete -F _dp {{.Name}}
`))

var zshCompletionTmpl = template.Must(template.New("zsh").Parse(`# zsh completion for {{.Name}}
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
                aliases=($({{.Name}} aliases 2>/dev/null))
                if [[ "$cur_word" == *@* ]]; then
                    cur_word="${cur_word#*@}"
                fi
                compstate[word]="$cur_word"
                _describe "alias" aliases
                ;;
            --location)
                local -a locations
                locations=(${(f)"$({{.Name}} locations 2>/dev/null)"})
                _describe "location" locations
                ;;
            --region)
                local -a regions
                regions=($({{.Name}} regions 2>/dev/null))
                _describe "region" regions
                ;;
            --power)
                local -a power_statuses
                power_statuses=($({{.Name}} power 2>/dev/null))
                _describe "power" power_statuses
                ;;
            --status)
                local -a server_statuses
                server_statuses=($({{.Name}} status 2>/dev/null))
                _describe "status" server_statuses
                ;;
            *)
                local -a aliases
                aliases=($({{.Name}} aliases 2>/dev/null))
                if [[ "$cur_word" == *@* ]]; then
                    cur_word="${cur_word#*@}"
                fi
                compstate[word]="$cur_word"
                _describe "alias" aliases
                ;;
        esac
    fi
}

compdef _dp {{.Name}}
`))

var fishCompletionTmpl = template.Must(template.New("fish").Parse(`# fish completion for {{.Name}}
function __fish_{{.Name}}_needs_command
    set -l cmd (commandline -opc)
    if test (count $cmd) -eq 1
        return 0
    end
    return 1
end

function __fish_{{.Name}}_get_aliases
    {{.Name}} aliases 2>/dev/null
end

function __fish_{{.Name}}_get_locations
    {{.Name}} locations 2>/dev/null
end

function __fish_{{.Name}}_get_regions
    {{.Name}} regions 2>/dev/null
end

function __fish_{{.Name}}_complete_alias
    set -l cmd (commandline -opc)
    set -l cur (commandline -ct)
    set -l word_to_complete "$cur"
    if string match -q '*@*' "$cur"
        set word_to_complete (string split '@' "$cur")[-1]
    end
    __fish_{{.Name}}_get_aliases | string match "$word_to_complete*"
end

function __fish_{{.Name}}_complete_filter
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
            __fish_{{.Name}}_get_aliases | string match "$word_to_complete*"
        case "location"
            __fish_{{.Name}}_get_locations | string match "$word_to_complete*"
        case "region"
            __fish_{{.Name}}_get_regions | string match "$word_to_complete*"
        case "power"
            {{.Name}} power 2>/dev/null | string match "$word_to_complete*"
        case "status"
            {{.Name}} status 2>/dev/null | string match "$word_to_complete*"
    end
end

complete -c {{.Name}} -f -n '__fish_{{.Name}}_needs_command'
complete -c {{.Name}} -a 'show' -d 'List servers with optional regex filter'
complete -c {{.Name}} -a 'ssh' -d 'SSH to server by alias'
complete -c {{.Name}} -a 'completion' -d 'Generate completion script'
complete -c {{.Name}} -a 'aliases' -d 'List all server aliases'
complete -c {{.Name}} -a 'locations' -d 'List all available locations'
complete -c {{.Name}} -a 'regions' -d 'List all available regions'
complete -c {{.Name}} -a 'power' -d 'List all power statuses'
complete -c {{.Name}} -a 'status' -d 'List all server statuses'
complete -c {{.Name}} -f -a '(__fish_{{.Name}}_complete_alias)'
complete -c {{.Name}} -f -a '(__fish_{{.Name}}_complete_filter)'
`))

func renderBashCompletion(name string) string {
	var buf strings.Builder
	bashCompletionTmpl.Execute(&buf, struct{ Name string }{Name: name})
	return buf.String()
}

func renderZshCompletion(name string) string {
	var buf strings.Builder
	zshCompletionTmpl.Execute(&buf, struct{ Name string }{Name: name})
	return buf.String()
}

func renderFishCompletion(name string) string {
	var buf strings.Builder
	fishCompletionTmpl.Execute(&buf, struct{ Name string }{Name: name})
	return buf.String()
}
