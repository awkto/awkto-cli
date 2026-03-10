package cmd

import (
	"fmt"
	"os"
)

var Version = "dev"

func Execute() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "dns":
		runDNS(os.Args[2:])
	case "lease":
		runLease(os.Args[2:])
	case "reserve":
		runReserve(os.Args[2:])
	case "server":
		runServer(os.Args[2:])
	case "completion":
		runCompletion(os.Args[2:])
	case "__complete_servers":
		runCompleteServers()
	case "version":
		fmt.Printf("awkto %s\n", Version)
	case "help", "--help", "-h":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Printf(`awkto %s - CLI for managing Kea DHCP and DNS records

Usage:
  awkto <command> <action> [options]

Commands:
  dns       Manage DNS records (create, list, delete, edit)
  lease     Manage DHCP leases (list, delete, promote)
  reserve   Manage DHCP reservations (list, create, delete, edit)
  server    Manage server configurations
  completion  Generate shell completion scripts (bash, zsh)
  version   Print version

Configuration:
  Config is loaded from ~/.awkto/config.yaml (override with AWKTO_CONFIG env var).
  Use 'awkto server add <name>' to create servers and 'awkto server use <name>'
  to set defaults. See 'awkto server --help' for details.

Environment Variables (override config file values):
  AWKTO_KEA_URL      Kea API base URL (e.g. https://kea.example.com:8080)
  AWKTO_KEA_TOKEN    Kea API bearer token
  AWKTO_DNS_URL      DNS API base URL (e.g. https://dns.example.com)
  AWKTO_DNS_TOKEN    DNS API bearer token
  AWKTO_SUBNET_ID    DHCP subnet ID (default: 1)
  AWKTO_CONFIG       Path to config file (default: ~/.awkto/config.yaml)

Run 'awkto <command> --help' for more information on a command.
`, Version)
}

func requireArgs(args []string, min int, usage string) {
	if len(args) < min {
		fmt.Fprintf(os.Stderr, "Error: missing required arguments\nUsage: %s\n", usage)
		os.Exit(1)
	}
}

func exitErr(err error) {
	fmt.Fprintf(os.Stderr, "Error: %v\n", err)
	os.Exit(1)
}
