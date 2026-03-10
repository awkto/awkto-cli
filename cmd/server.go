package cmd

import (
	"flag"
	"fmt"
	"os"
	"sort"
	"text/tabwriter"

	"github.com/awkto/awkto-cli/internal/config"
)

func runServer(args []string) {
	if len(args) < 1 {
		printServerUsage()
		os.Exit(1)
	}

	switch args[0] {
	case "list":
		serverListCmd("")
	case "dns":
		runServerType(args[1:], "dns")
	case "kea":
		runServerType(args[1:], "kea")
	case "add":
		serverAddCmd(args[1:])
	case "use":
		serverUseCmd(args[1:])
	case "remove":
		serverRemoveCmd(args[1:])
	case "show":
		serverShowCmd()
	case "help", "--help", "-h":
		printServerUsage()
	default:
		fmt.Fprintf(os.Stderr, "Unknown server action: %s\n\n", args[0])
		printServerUsage()
		os.Exit(1)
	}
}

func runServerType(args []string, serverType string) {
	if len(args) < 1 || args[0] == "help" || args[0] == "--help" || args[0] == "-h" {
		fmt.Printf("Usage: awkto server %s list\n\nActions:\n  list    List all %s servers\n", serverType, serverType)
		if len(args) < 1 {
			os.Exit(1)
		}
		return
	}
	switch args[0] {
	case "list":
		serverListCmd(serverType)
	default:
		fmt.Fprintf(os.Stderr, "Unknown server %s action: %s\n", serverType, args[0])
		os.Exit(1)
	}
}

func printServerUsage() {
	fmt.Print(`Usage: awkto server <action> [options]

Actions:
  list                                         List all servers
  dns list                                     List DNS servers only
  kea list                                     List KEA servers only
  add     <name> --type dns|kea --url <url> --token <token> [--subnet-id N]
                                               Add a new server
  use     <name>                               Set server as default for its type
  remove  <name>                               Remove a server
  show                                         Show current defaults with details
`)
}

func serverListCmd(filterType string) {
	cf, err := config.LoadConfigFile()
	if err != nil {
		exitErr(err)
	}

	if len(cf.Servers) == 0 {
		fmt.Println("No servers configured.")
		fmt.Printf("Config file: %s\n", config.ConfigFilePath())
		return
	}

	names := make([]string, 0, len(cf.Servers))
	for name := range cf.Servers {
		srv := cf.Servers[name]
		if filterType != "" && srv.Type != filterType {
			continue
		}
		names = append(names, name)
	}
	sort.Strings(names)

	if len(names) == 0 {
		fmt.Printf("No %s servers configured.\n", filterType)
		return
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
	fmt.Fprintln(w, "  NAME\tTYPE\tURL")
	for _, name := range names {
		srv := cf.Servers[name]
		marker := " "
		if cf.Defaults[srv.Type] == name {
			marker = "*"
		}
		fmt.Fprintf(w, "%s %s\t%s\t%s\n", marker, name, srv.Type, srv.URL)
	}
	w.Flush()
}

func serverAddCmd(args []string) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "Error: server name is required")
		fmt.Fprintln(os.Stderr, "Usage: awkto server add <name> --type dns|kea --url <url> --token <token> [--subnet-id N]")
		os.Exit(1)
	}

	name := args[0]

	fs := flag.NewFlagSet("server add", flag.ExitOnError)
	sType := fs.String("type", "", "Server type (dns or kea)")
	url := fs.String("url", "", "Server URL")
	token := fs.String("token", "", "Bearer token")
	subnetID := fs.String("subnet-id", "", "DHCP subnet ID (kea only)")
	fs.Parse(args[1:])

	if *sType == "" || *url == "" {
		fmt.Fprintln(os.Stderr, "Error: --type and --url are required")
		os.Exit(1)
	}

	if *sType != "dns" && *sType != "kea" {
		fmt.Fprintln(os.Stderr, "Error: --type must be dns or kea")
		os.Exit(1)
	}

	cf, err := config.LoadConfigFile()
	if err != nil {
		exitErr(err)
	}

	srv := config.Server{
		Type:  *sType,
		URL:   *url,
		Token: *token,
	}
	if *subnetID != "" {
		srv.SubnetID = *subnetID
	}

	cf.Servers[name] = srv

	// If this is the first server of its type, make it the default
	hasDefault := false
	if d, ok := cf.Defaults[*sType]; ok && d != "" {
		// Verify the default server still exists
		if _, exists := cf.Servers[d]; exists {
			hasDefault = true
		}
	}
	if !hasDefault {
		cf.Defaults[*sType] = name
	}

	if err := config.SaveConfigFile(cf); err != nil {
		exitErr(err)
	}

	fmt.Printf("Added %s server %q.\n", *sType, name)
	if cf.Defaults[*sType] == name {
		fmt.Printf("Set as default %s server.\n", *sType)
	}
}

func serverUseCmd(args []string) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "Error: server name is required")
		fmt.Fprintln(os.Stderr, "Usage: awkto server use <name>")
		os.Exit(1)
	}

	name := args[0]

	cf, err := config.LoadConfigFile()
	if err != nil {
		exitErr(err)
	}

	srv, ok := cf.Servers[name]
	if !ok {
		exitErr(fmt.Errorf("server %q not found", name))
	}

	cf.Defaults[srv.Type] = name

	if err := config.SaveConfigFile(cf); err != nil {
		exitErr(err)
	}

	fmt.Printf("Set %q as default %s server.\n", name, srv.Type)
}

func serverRemoveCmd(args []string) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "Error: server name is required")
		fmt.Fprintln(os.Stderr, "Usage: awkto server remove <name>")
		os.Exit(1)
	}

	name := args[0]

	cf, err := config.LoadConfigFile()
	if err != nil {
		exitErr(err)
	}

	srv, ok := cf.Servers[name]
	if !ok {
		exitErr(fmt.Errorf("server %q not found", name))
	}

	delete(cf.Servers, name)

	// Clear default if this was the default for its type
	if cf.Defaults[srv.Type] == name {
		delete(cf.Defaults, srv.Type)
	}

	if err := config.SaveConfigFile(cf); err != nil {
		exitErr(err)
	}

	fmt.Printf("Removed server %q.\n", name)
}

func serverShowCmd() {
	cf, err := config.LoadConfigFile()
	if err != nil {
		exitErr(err)
	}

	dnsDefault := cf.Defaults["dns"]
	keaDefault := cf.Defaults["kea"]

	if dnsDefault == "" && keaDefault == "" {
		fmt.Println("No default servers set.")
		return
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)

	if dnsDefault != "" {
		srv, ok := cf.Servers[dnsDefault]
		if ok {
			fmt.Printf("Default DNS server: %s\n", dnsDefault)
			fmt.Fprintf(w, "  url:\t%s\n", srv.URL)
			fmt.Fprintf(w, "  token:\t%s\n", maskToken(srv.Token))
		} else {
			fmt.Printf("Default DNS server: %s (not found in servers)\n", dnsDefault)
		}
	} else {
		fmt.Println("Default DNS server: (none)")
	}

	if keaDefault != "" {
		srv, ok := cf.Servers[keaDefault]
		if ok {
			fmt.Printf("Default KEA server: %s\n", keaDefault)
			fmt.Fprintf(w, "  url:\t%s\n", srv.URL)
			fmt.Fprintf(w, "  token:\t%s\n", maskToken(srv.Token))
			fmt.Fprintf(w, "  subnet_id:\t%s\n", valueOrDash(srv.SubnetID))
		} else {
			fmt.Printf("Default KEA server: %s (not found in servers)\n", keaDefault)
		}
	} else {
		fmt.Println("Default KEA server: (none)")
	}

	w.Flush()
}

func valueOrDash(s string) string {
	if s == "" {
		return "-"
	}
	return s
}

func maskToken(s string) string {
	if s == "" {
		return "-"
	}
	if len(s) <= 6 {
		return "***"
	}
	return s[:3] + "***" + s[len(s)-3:]
}
