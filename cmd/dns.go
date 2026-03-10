package cmd

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/awkto/awkto-cli/internal/client"
	"github.com/awkto/awkto-cli/internal/config"
)

func runDNS(args []string) {
	if len(args) < 1 {
		printDNSUsage()
		os.Exit(1)
	}

	switch args[0] {
	case "list":
		dnsListCmd(args[1:])
	case "create":
		dnsCreateCmd(args[1:])
	case "edit":
		dnsEditCmd(args[1:])
	case "delete":
		dnsDeleteCmd(args[1:])
	case "help", "--help", "-h":
		printDNSUsage()
	default:
		fmt.Fprintf(os.Stderr, "Unknown dns action: %s\n\n", args[0])
		printDNSUsage()
		os.Exit(1)
	}
}

func printDNSUsage() {
	fmt.Print(`Usage: awkto dns <action> [options]

Actions:
  list                          List all DNS records
  create  -name -type -values [-ttl]  Create a DNS record
  edit    -name -type [-values] [-ttl] Update a DNS record
  delete  -name -type                  Delete a DNS record

Flags:
  -name     Record name (e.g. www)
  -type     Record type (A, AAAA, CNAME, MX, TXT, etc.)
  -values   Comma-separated values (e.g. "192.168.1.1" or "192.168.1.1,192.168.1.2")
  -ttl      TTL in seconds (default: 300)
  -filter   Filter list by type (e.g. -filter A)
  -server   Use a specific named server instead of the default
`)
}

func loadDNSConfig(fs *flag.FlagSet) *config.Config {
	var serverName string
	fs.StringVar(&serverName, "server", "", "Use a specific named server")
	return nil // placeholder; actual loading happens after Parse
}

func getDNSClient(fs *flag.FlagSet) *client.DNSClient {
	serverName := fs.Lookup("server").Value.String()
	cfg, err := config.LoadForDNS(serverName)
	if err != nil {
		exitErr(err)
	}
	if err := cfg.RequireDNS(); err != nil {
		exitErr(err)
	}
	return client.NewDNSClient(cfg)
}

func dnsListCmd(args []string) {
	fs := flag.NewFlagSet("dns list", flag.ExitOnError)
	filterType := fs.String("filter", "", "Filter by record type")
	fs.String("server", "", "Use a specific named server")
	fs.Parse(args)

	c := getDNSClient(fs)

	records, err := c.ListRecords()
	if err != nil {
		exitErr(err)
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
	fmt.Fprintln(w, "NAME\tTYPE\tTTL\tVALUES")
	for _, r := range records {
		if *filterType != "" && !strings.EqualFold(r.Type, *filterType) {
			continue
		}
		fmt.Fprintf(w, "%s\t%s\t%d\t%s\n", r.Name, r.Type, r.TTL, strings.Join(r.Values, ", "))
	}
	w.Flush()
}

func dnsCreateCmd(args []string) {
	fs := flag.NewFlagSet("dns create", flag.ExitOnError)
	name := fs.String("name", "", "Record name")
	rtype := fs.String("type", "A", "Record type")
	values := fs.String("values", "", "Comma-separated values")
	ttl := fs.Int("ttl", 300, "TTL in seconds")
	fs.String("server", "", "Use a specific named server")
	fs.Parse(args)

	if *name == "" || *values == "" {
		fmt.Fprintln(os.Stderr, "Error: -name and -values are required")
		fs.Usage()
		os.Exit(1)
	}

	c := getDNSClient(fs)

	vals := splitValues(*values)
	err := c.CreateRecord(client.DNSRecordCreate{
		Name:   *name,
		Type:   strings.ToUpper(*rtype),
		TTL:    *ttl,
		Values: vals,
	})
	if err != nil {
		exitErr(err)
	}
	fmt.Printf("Created %s record: %s -> %s\n", strings.ToUpper(*rtype), *name, *values)
}

func dnsEditCmd(args []string) {
	fs := flag.NewFlagSet("dns edit", flag.ExitOnError)
	name := fs.String("name", "", "Record name")
	rtype := fs.String("type", "", "Record type")
	values := fs.String("values", "", "Comma-separated values")
	ttl := fs.Int("ttl", 0, "TTL in seconds")
	fs.String("server", "", "Use a specific named server")
	fs.Parse(args)

	if *name == "" || *rtype == "" {
		fmt.Fprintln(os.Stderr, "Error: -name and -type are required")
		fs.Usage()
		os.Exit(1)
	}

	c := getDNSClient(fs)

	update := client.DNSRecordUpdate{}
	if *values != "" {
		update.Values = splitValues(*values)
	}
	if *ttl > 0 {
		update.TTL = *ttl
	}

	err := c.UpdateRecord(strings.ToUpper(*rtype), *name, update)
	if err != nil {
		exitErr(err)
	}
	fmt.Printf("Updated %s record: %s\n", strings.ToUpper(*rtype), *name)
}

func dnsDeleteCmd(args []string) {
	fs := flag.NewFlagSet("dns delete", flag.ExitOnError)
	name := fs.String("name", "", "Record name")
	rtype := fs.String("type", "", "Record type")
	fs.String("server", "", "Use a specific named server")
	fs.Parse(args)

	if *name == "" || *rtype == "" {
		fmt.Fprintln(os.Stderr, "Error: -name and -type are required")
		fs.Usage()
		os.Exit(1)
	}

	c := getDNSClient(fs)

	err := c.DeleteRecord(strings.ToUpper(*rtype), *name)
	if err != nil {
		exitErr(err)
	}
	fmt.Printf("Deleted %s record: %s\n", strings.ToUpper(*rtype), *name)
}

func splitValues(s string) []string {
	parts := strings.Split(s, ",")
	var result []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}
