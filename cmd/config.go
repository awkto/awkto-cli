package cmd

import (
	"flag"
	"fmt"
	"os"
	"sort"
	"text/tabwriter"

	"github.com/awkto/awkto-cli/internal/config"
)

func runConfig(args []string) {
	if len(args) < 1 {
		printConfigUsage()
		os.Exit(1)
	}

	switch args[0] {
	case "list":
		configListCmd()
	case "use":
		configUseCmd(args[1:])
	case "show":
		configShowCmd()
	case "add":
		configAddCmd(args[1:])
	case "remove":
		configRemoveCmd(args[1:])
	case "help", "--help", "-h":
		printConfigUsage()
	default:
		fmt.Fprintf(os.Stderr, "Unknown config action: %s\n\n", args[0])
		printConfigUsage()
		os.Exit(1)
	}
}

func printConfigUsage() {
	fmt.Print(`Usage: awkto config <action> [options]

Actions:
  list                          List all contexts
  use     <context-name>        Switch active context
  show                          Show current context details
  add     <name> [flags]        Add a new context
  remove  <name>                Remove a context

Add Flags:
  -dns-url    DNS API base URL
  -dns-token  DNS API bearer token
  -kea-url    Kea API base URL
  -kea-token  Kea API bearer token
  -subnet-id  DHCP subnet ID
`)
}

func configListCmd() {
	cf, err := config.LoadConfigFile()
	if err != nil {
		exitErr(err)
	}

	if len(cf.Contexts) == 0 {
		fmt.Println("No contexts configured.")
		fmt.Printf("Config file: %s\n", config.ConfigFilePath())
		return
	}

	names := make([]string, 0, len(cf.Contexts))
	for name := range cf.Contexts {
		names = append(names, name)
	}
	sort.Strings(names)

	w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
	fmt.Fprintln(w, "  NAME\tDNS URL\tKEA URL")
	for _, name := range names {
		ctx := cf.Contexts[name]
		marker := " "
		if name == cf.CurrentContext {
			marker = "*"
		}
		dnsURL := ctx.DNSURL
		if dnsURL == "" {
			dnsURL = "-"
		}
		keaURL := ctx.KeaURL
		if keaURL == "" {
			keaURL = "-"
		}
		fmt.Fprintf(w, "%s %s\t%s\t%s\n", marker, name, dnsURL, keaURL)
	}
	w.Flush()
}

func configUseCmd(args []string) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "Error: context name is required")
		fmt.Fprintln(os.Stderr, "Usage: awkto config use <context-name>")
		os.Exit(1)
	}

	name := args[0]

	cf, err := config.LoadConfigFile()
	if err != nil {
		exitErr(err)
	}

	if _, ok := cf.Contexts[name]; !ok {
		exitErr(fmt.Errorf("context %q not found", name))
	}

	cf.CurrentContext = name
	if err := config.SaveConfigFile(cf); err != nil {
		exitErr(err)
	}

	fmt.Printf("Switched to context %q.\n", name)
}

func configShowCmd() {
	cf, err := config.LoadConfigFile()
	if err != nil {
		exitErr(err)
	}

	if cf.CurrentContext == "" {
		fmt.Println("No current context set.")
		return
	}

	ctx, ok := cf.Contexts[cf.CurrentContext]
	if !ok {
		fmt.Printf("Current context %q not found in config file.\n", cf.CurrentContext)
		return
	}

	fmt.Printf("Current context: %s\n\n", cf.CurrentContext)
	w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
	fmt.Fprintf(w, "  dns_url:\t%s\n", valueOrDash(ctx.DNSURL))
	fmt.Fprintf(w, "  dns_token:\t%s\n", maskToken(ctx.DNSToken))
	fmt.Fprintf(w, "  kea_url:\t%s\n", valueOrDash(ctx.KeaURL))
	fmt.Fprintf(w, "  kea_token:\t%s\n", maskToken(ctx.KeaToken))
	fmt.Fprintf(w, "  subnet_id:\t%s\n", valueOrDash(ctx.SubnetID))
	w.Flush()
}

func configAddCmd(args []string) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "Error: context name is required")
		fmt.Fprintln(os.Stderr, "Usage: awkto config add <name> [flags]")
		os.Exit(1)
	}

	name := args[0]

	fs := flag.NewFlagSet("config add", flag.ExitOnError)
	dnsURL := fs.String("dns-url", "", "DNS API base URL")
	dnsToken := fs.String("dns-token", "", "DNS API bearer token")
	keaURL := fs.String("kea-url", "", "Kea API base URL")
	keaToken := fs.String("kea-token", "", "Kea API bearer token")
	subnetID := fs.String("subnet-id", "", "DHCP subnet ID")
	fs.Parse(args[1:])

	cf, err := config.LoadConfigFile()
	if err != nil {
		exitErr(err)
	}

	cf.Contexts[name] = config.Context{
		DNSURL:   *dnsURL,
		DNSToken: *dnsToken,
		KeaURL:   *keaURL,
		KeaToken: *keaToken,
		SubnetID: *subnetID,
	}

	// If this is the first context, make it the current one
	if len(cf.Contexts) == 1 {
		cf.CurrentContext = name
	}

	if err := config.SaveConfigFile(cf); err != nil {
		exitErr(err)
	}

	fmt.Printf("Added context %q.\n", name)
	if cf.CurrentContext == name {
		fmt.Printf("Set as current context.\n")
	}
}

func configRemoveCmd(args []string) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "Error: context name is required")
		fmt.Fprintln(os.Stderr, "Usage: awkto config remove <name>")
		os.Exit(1)
	}

	name := args[0]

	cf, err := config.LoadConfigFile()
	if err != nil {
		exitErr(err)
	}

	if _, ok := cf.Contexts[name]; !ok {
		exitErr(fmt.Errorf("context %q not found", name))
	}

	delete(cf.Contexts, name)

	if cf.CurrentContext == name {
		cf.CurrentContext = ""
	}

	if err := config.SaveConfigFile(cf); err != nil {
		exitErr(err)
	}

	fmt.Printf("Removed context %q.\n", name)
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
