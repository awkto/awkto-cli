package cmd

import (
	"fmt"
	"os"
	"sort"

	"github.com/awkto/awkto-cli/internal/config"
)

func runCompletion(args []string) {
	if len(args) < 1 {
		printCompletionUsage()
		os.Exit(1)
	}

	switch args[0] {
	case "bash":
		fmt.Print(bashCompletionScript)
	case "zsh":
		fmt.Print(zshCompletionScript)
	case "help", "--help", "-h":
		printCompletionUsage()
	default:
		fmt.Fprintf(os.Stderr, "Unknown shell: %s\n\n", args[0])
		printCompletionUsage()
		os.Exit(1)
	}
}

func printCompletionUsage() {
	fmt.Print(`Usage: awkto completion <shell>

Shells:
  bash    Generate bash completion script
  zsh     Generate zsh completion script

To load completions:

  Bash:
    $ source <(awkto completion bash)
    # Or persist:
    $ awkto completion bash > /etc/bash_completion.d/awkto

  Zsh:
    $ source <(awkto completion zsh)
    # Or persist:
    $ awkto completion zsh > "${fpath[1]}/_awkto"
`)
}

// runCompleteContexts prints context names one per line for shell completion.
func runCompleteContexts() {
	cf, err := config.LoadConfigFile()
	if err != nil {
		os.Exit(0)
	}
	names := make([]string, 0, len(cf.Contexts))
	for name := range cf.Contexts {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		fmt.Println(name)
	}
}

const bashCompletionScript = `# bash completion for awkto

_awkto_completions() {
    local cur prev words cword
    _init_completion || return

    local commands="dns lease reserve config version help completion"
    local dns_actions="list create edit delete"
    local lease_actions="list delete promote"
    local reserve_actions="list create edit delete"
    local config_actions="list use show add remove"
    local completion_actions="bash zsh"

    case "${cword}" in
        1)
            COMPREPLY=($(compgen -W "${commands}" -- "${cur}"))
            return
            ;;
        2)
            case "${words[1]}" in
                dns)
                    COMPREPLY=($(compgen -W "${dns_actions}" -- "${cur}"))
                    return
                    ;;
                lease)
                    COMPREPLY=($(compgen -W "${lease_actions}" -- "${cur}"))
                    return
                    ;;
                reserve)
                    COMPREPLY=($(compgen -W "${reserve_actions}" -- "${cur}"))
                    return
                    ;;
                config)
                    COMPREPLY=($(compgen -W "${config_actions}" -- "${cur}"))
                    return
                    ;;
                completion)
                    COMPREPLY=($(compgen -W "${completion_actions}" -- "${cur}"))
                    return
                    ;;
            esac
            ;;
        *)
            # Handle flags for subcommands and dynamic completion
            case "${words[1]}" in
                dns)
                    case "${words[2]}" in
                        create)
                            COMPREPLY=($(compgen -W "-name -type -values -ttl" -- "${cur}"))
                            return
                            ;;
                        edit)
                            COMPREPLY=($(compgen -W "-name -type -values -ttl" -- "${cur}"))
                            return
                            ;;
                        delete)
                            COMPREPLY=($(compgen -W "-name -type" -- "${cur}"))
                            return
                            ;;
                        list)
                            COMPREPLY=($(compgen -W "-filter" -- "${cur}"))
                            return
                            ;;
                    esac
                    ;;
                lease)
                    case "${words[2]}" in
                        list)
                            COMPREPLY=($(compgen -W "-subnet" -- "${cur}"))
                            return
                            ;;
                        delete)
                            COMPREPLY=($(compgen -W "-ip -mac" -- "${cur}"))
                            return
                            ;;
                        promote)
                            COMPREPLY=($(compgen -W "-ip -hostname -subnet" -- "${cur}"))
                            return
                            ;;
                    esac
                    ;;
                reserve)
                    case "${words[2]}" in
                        list)
                            COMPREPLY=($(compgen -W "-subnet" -- "${cur}"))
                            return
                            ;;
                        create)
                            COMPREPLY=($(compgen -W "-ip -mac -hostname -subnet" -- "${cur}"))
                            return
                            ;;
                        edit)
                            COMPREPLY=($(compgen -W "-ip -mac -hostname -subnet" -- "${cur}"))
                            return
                            ;;
                        delete)
                            COMPREPLY=($(compgen -W "-ip" -- "${cur}"))
                            return
                            ;;
                    esac
                    ;;
                config)
                    case "${words[2]}" in
                        add)
                            COMPREPLY=($(compgen -W "--dns-url --dns-token --kea-url --kea-token --subnet-id" -- "${cur}"))
                            return
                            ;;
                        use)
                            local contexts
                            contexts=$(awkto __complete_contexts 2>/dev/null)
                            COMPREPLY=($(compgen -W "${contexts}" -- "${cur}"))
                            return
                            ;;
                    esac
                    ;;
            esac
            ;;
    esac
}

complete -F _awkto_completions awkto
`

const zshCompletionScript = `#compdef awkto

_awkto() {
    local -a commands dns_actions lease_actions reserve_actions config_actions completion_actions

    commands=(
        'dns:Manage DNS records'
        'lease:Manage DHCP leases'
        'reserve:Manage DHCP reservations'
        'config:Manage CLI configuration contexts'
        'version:Print version'
        'help:Show help'
        'completion:Generate shell completion scripts'
    )

    dns_actions=(
        'list:List all DNS records'
        'create:Create a DNS record'
        'edit:Update a DNS record'
        'delete:Delete a DNS record'
    )

    lease_actions=(
        'list:List all DHCP leases'
        'delete:Delete a lease'
        'promote:Promote a lease to a reservation'
    )

    reserve_actions=(
        'list:List all DHCP reservations'
        'create:Create a reservation'
        'edit:Edit a reservation'
        'delete:Delete a reservation'
    )

    config_actions=(
        'list:List all contexts'
        'use:Switch active context'
        'show:Show current context details'
        'add:Add a new context'
        'remove:Remove a context'
    )

    completion_actions=(
        'bash:Generate bash completion script'
        'zsh:Generate zsh completion script'
    )

    if (( CURRENT == 2 )); then
        _describe -t commands 'awkto commands' commands
        return
    fi

    case "${words[2]}" in
        dns)
            if (( CURRENT == 3 )); then
                _describe -t dns-actions 'dns actions' dns_actions
            else
                case "${words[3]}" in
                    create)
                        _arguments \
                            '-name[Record name]:name:' \
                            '-type[Record type]:type:(A AAAA CNAME MX TXT SRV NS PTR)' \
                            '-values[Comma-separated values]:values:' \
                            '-ttl[TTL in seconds]:ttl:'
                        ;;
                    edit)
                        _arguments \
                            '-name[Record name]:name:' \
                            '-type[Record type]:type:(A AAAA CNAME MX TXT SRV NS PTR)' \
                            '-values[Comma-separated values]:values:' \
                            '-ttl[TTL in seconds]:ttl:'
                        ;;
                    delete)
                        _arguments \
                            '-name[Record name]:name:' \
                            '-type[Record type]:type:(A AAAA CNAME MX TXT SRV NS PTR)'
                        ;;
                    list)
                        _arguments \
                            '-filter[Filter by record type]:type:(A AAAA CNAME MX TXT SRV NS PTR)'
                        ;;
                esac
            fi
            ;;
        lease)
            if (( CURRENT == 3 )); then
                _describe -t lease-actions 'lease actions' lease_actions
            else
                case "${words[3]}" in
                    list)
                        _arguments '-subnet[Subnet ID]:subnet:'
                        ;;
                    delete)
                        _arguments \
                            '-ip[IP address]:ip:' \
                            '-mac[MAC address]:mac:'
                        ;;
                    promote)
                        _arguments \
                            '-ip[IP address]:ip:' \
                            '-hostname[Hostname]:hostname:' \
                            '-subnet[Subnet ID]:subnet:'
                        ;;
                esac
            fi
            ;;
        reserve)
            if (( CURRENT == 3 )); then
                _describe -t reserve-actions 'reserve actions' reserve_actions
            else
                case "${words[3]}" in
                    list)
                        _arguments '-subnet[Subnet ID]:subnet:'
                        ;;
                    create)
                        _arguments \
                            '-ip[IP address]:ip:' \
                            '-mac[MAC address]:mac:' \
                            '-hostname[Hostname]:hostname:' \
                            '-subnet[Subnet ID]:subnet:'
                        ;;
                    edit)
                        _arguments \
                            '-ip[IP address]:ip:' \
                            '-mac[MAC address]:mac:' \
                            '-hostname[Hostname]:hostname:' \
                            '-subnet[Subnet ID]:subnet:'
                        ;;
                    delete)
                        _arguments '-ip[IP address]:ip:'
                        ;;
                esac
            fi
            ;;
        config)
            if (( CURRENT == 3 )); then
                _describe -t config-actions 'config actions' config_actions
            else
                case "${words[3]}" in
                    add)
                        _arguments \
                            '--dns-url[DNS API base URL]:url:' \
                            '--dns-token[DNS API bearer token]:token:' \
                            '--kea-url[Kea API base URL]:url:' \
                            '--kea-token[Kea API bearer token]:token:' \
                            '--subnet-id[DHCP subnet ID]:subnet:'
                        ;;
                    use)
                        local -a contexts
                        contexts=(${(f)"$(awkto __complete_contexts 2>/dev/null)"})
                        _describe -t contexts 'config contexts' contexts
                        ;;
                esac
            fi
            ;;
        completion)
            if (( CURRENT == 3 )); then
                _describe -t completion-actions 'shell' completion_actions
            fi
            ;;
    esac
}

compdef _awkto awkto
`
