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

// runCompleteServers prints server names one per line for shell completion.
func runCompleteServers() {
	cf, err := config.LoadConfigFile()
	if err != nil {
		os.Exit(0)
	}
	names := make([]string, 0, len(cf.Servers))
	for name := range cf.Servers {
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

    local commands="dns lease reserve server version help completion"
    local dns_actions="list create edit delete"
    local lease_actions="list delete promote"
    local reserve_actions="list create edit delete"
    local server_actions="list dns kea add use remove show"
    local server_dns_actions="list"
    local server_kea_actions="list"
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
                server)
                    COMPREPLY=($(compgen -W "${server_actions}" -- "${cur}"))
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
                            COMPREPLY=($(compgen -W "-name -type -values -ttl -server" -- "${cur}"))
                            return
                            ;;
                        edit)
                            COMPREPLY=($(compgen -W "-name -type -values -ttl -server" -- "${cur}"))
                            return
                            ;;
                        delete)
                            COMPREPLY=($(compgen -W "-name -type -server" -- "${cur}"))
                            return
                            ;;
                        list)
                            COMPREPLY=($(compgen -W "-filter -server" -- "${cur}"))
                            return
                            ;;
                    esac
                    # Complete -server flag value
                    if [[ "${prev}" == "-server" ]]; then
                        local servers
                        servers=$(awkto __complete_servers 2>/dev/null)
                        COMPREPLY=($(compgen -W "${servers}" -- "${cur}"))
                        return
                    fi
                    ;;
                lease)
                    case "${words[2]}" in
                        list)
                            COMPREPLY=($(compgen -W "-subnet -server" -- "${cur}"))
                            return
                            ;;
                        delete)
                            COMPREPLY=($(compgen -W "-ip -mac -server" -- "${cur}"))
                            return
                            ;;
                        promote)
                            COMPREPLY=($(compgen -W "-ip -hostname -subnet -server" -- "${cur}"))
                            return
                            ;;
                    esac
                    if [[ "${prev}" == "-server" ]]; then
                        local servers
                        servers=$(awkto __complete_servers 2>/dev/null)
                        COMPREPLY=($(compgen -W "${servers}" -- "${cur}"))
                        return
                    fi
                    ;;
                reserve)
                    case "${words[2]}" in
                        list)
                            COMPREPLY=($(compgen -W "-subnet -server" -- "${cur}"))
                            return
                            ;;
                        create)
                            COMPREPLY=($(compgen -W "-ip -mac -hostname -subnet -server" -- "${cur}"))
                            return
                            ;;
                        edit)
                            COMPREPLY=($(compgen -W "-ip -mac -hostname -subnet -server" -- "${cur}"))
                            return
                            ;;
                        delete)
                            COMPREPLY=($(compgen -W "-ip -server" -- "${cur}"))
                            return
                            ;;
                    esac
                    if [[ "${prev}" == "-server" ]]; then
                        local servers
                        servers=$(awkto __complete_servers 2>/dev/null)
                        COMPREPLY=($(compgen -W "${servers}" -- "${cur}"))
                        return
                    fi
                    ;;
                server)
                    case "${words[2]}" in
                        add)
                            COMPREPLY=($(compgen -W "--type --url --token --subnet-id" -- "${cur}"))
                            return
                            ;;
                        use|remove)
                            local servers
                            servers=$(awkto __complete_servers 2>/dev/null)
                            COMPREPLY=($(compgen -W "${servers}" -- "${cur}"))
                            return
                            ;;
                        dns)
                            COMPREPLY=($(compgen -W "${server_dns_actions}" -- "${cur}"))
                            return
                            ;;
                        kea)
                            COMPREPLY=($(compgen -W "${server_kea_actions}" -- "${cur}"))
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
    local -a commands dns_actions lease_actions reserve_actions server_actions completion_actions server_dns_actions server_kea_actions

    commands=(
        'dns:Manage DNS records'
        'lease:Manage DHCP leases'
        'reserve:Manage DHCP reservations'
        'server:Manage server configurations'
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

    server_actions=(
        'list:List all servers'
        'dns:List DNS servers'
        'kea:List KEA servers'
        'add:Add a new server'
        'use:Set server as default for its type'
        'remove:Remove a server'
        'show:Show current defaults'
    )

    server_dns_actions=(
        'list:List DNS servers'
    )

    server_kea_actions=(
        'list:List KEA servers'
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
                            '-ttl[TTL in seconds]:ttl:' \
                            '-server[Use a specific named server]:server:{local -a servers; servers=(${(f)"$(awkto __complete_servers 2>/dev/null)"}); _describe "servers" servers}'
                        ;;
                    edit)
                        _arguments \
                            '-name[Record name]:name:' \
                            '-type[Record type]:type:(A AAAA CNAME MX TXT SRV NS PTR)' \
                            '-values[Comma-separated values]:values:' \
                            '-ttl[TTL in seconds]:ttl:' \
                            '-server[Use a specific named server]:server:{local -a servers; servers=(${(f)"$(awkto __complete_servers 2>/dev/null)"}); _describe "servers" servers}'
                        ;;
                    delete)
                        _arguments \
                            '-name[Record name]:name:' \
                            '-type[Record type]:type:(A AAAA CNAME MX TXT SRV NS PTR)' \
                            '-server[Use a specific named server]:server:{local -a servers; servers=(${(f)"$(awkto __complete_servers 2>/dev/null)"}); _describe "servers" servers}'
                        ;;
                    list)
                        _arguments \
                            '-filter[Filter by record type]:type:(A AAAA CNAME MX TXT SRV NS PTR)' \
                            '-server[Use a specific named server]:server:{local -a servers; servers=(${(f)"$(awkto __complete_servers 2>/dev/null)"}); _describe "servers" servers}'
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
                        _arguments \
                            '-subnet[Subnet ID]:subnet:' \
                            '-server[Use a specific named server]:server:{local -a servers; servers=(${(f)"$(awkto __complete_servers 2>/dev/null)"}); _describe "servers" servers}'
                        ;;
                    delete)
                        _arguments \
                            '-ip[IP address]:ip:' \
                            '-mac[MAC address]:mac:' \
                            '-server[Use a specific named server]:server:{local -a servers; servers=(${(f)"$(awkto __complete_servers 2>/dev/null)"}); _describe "servers" servers}'
                        ;;
                    promote)
                        _arguments \
                            '-ip[IP address]:ip:' \
                            '-hostname[Hostname]:hostname:' \
                            '-subnet[Subnet ID]:subnet:' \
                            '-server[Use a specific named server]:server:{local -a servers; servers=(${(f)"$(awkto __complete_servers 2>/dev/null)"}); _describe "servers" servers}'
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
                        _arguments \
                            '-subnet[Subnet ID]:subnet:' \
                            '-server[Use a specific named server]:server:{local -a servers; servers=(${(f)"$(awkto __complete_servers 2>/dev/null)"}); _describe "servers" servers}'
                        ;;
                    create)
                        _arguments \
                            '-ip[IP address]:ip:' \
                            '-mac[MAC address]:mac:' \
                            '-hostname[Hostname]:hostname:' \
                            '-subnet[Subnet ID]:subnet:' \
                            '-server[Use a specific named server]:server:{local -a servers; servers=(${(f)"$(awkto __complete_servers 2>/dev/null)"}); _describe "servers" servers}'
                        ;;
                    edit)
                        _arguments \
                            '-ip[IP address]:ip:' \
                            '-mac[MAC address]:mac:' \
                            '-hostname[Hostname]:hostname:' \
                            '-subnet[Subnet ID]:subnet:' \
                            '-server[Use a specific named server]:server:{local -a servers; servers=(${(f)"$(awkto __complete_servers 2>/dev/null)"}); _describe "servers" servers}'
                        ;;
                    delete)
                        _arguments \
                            '-ip[IP address]:ip:' \
                            '-server[Use a specific named server]:server:{local -a servers; servers=(${(f)"$(awkto __complete_servers 2>/dev/null)"}); _describe "servers" servers}'
                        ;;
                esac
            fi
            ;;
        server)
            if (( CURRENT == 3 )); then
                _describe -t server-actions 'server actions' server_actions
            else
                case "${words[3]}" in
                    add)
                        _arguments \
                            '--type[Server type]:type:(dns kea)' \
                            '--url[Server URL]:url:' \
                            '--token[Bearer token]:token:' \
                            '--subnet-id[DHCP subnet ID]:subnet:'
                        ;;
                    use|remove)
                        local -a servers
                        servers=(${(f)"$(awkto __complete_servers 2>/dev/null)"})
                        _describe -t servers 'servers' servers
                        ;;
                    dns)
                        if (( CURRENT == 4 )); then
                            _describe -t server-dns-actions 'server dns actions' server_dns_actions
                        fi
                        ;;
                    kea)
                        if (( CURRENT == 4 )); then
                            _describe -t server-kea-actions 'server kea actions' server_kea_actions
                        fi
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
