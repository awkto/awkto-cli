package cmd

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"text/tabwriter"

	"github.com/awkto/awkto-cli/internal/client"
)

func runReserve(args []string) {
	if len(args) < 1 {
		printReserveUsage()
		os.Exit(1)
	}

	if err := cfg.RequireKea(); err != nil {
		exitErr(err)
	}
	c := client.NewKeaClient(cfg)

	switch args[0] {
	case "list":
		reserveListCmd(c, args[1:])
	case "create":
		reserveCreateCmd(c, args[1:])
	case "delete":
		reserveDeleteCmd(c, args[1:])
	case "edit":
		reserveEditCmd(c, args[1:])
	case "help", "--help", "-h":
		printReserveUsage()
	default:
		fmt.Fprintf(os.Stderr, "Unknown reserve action: %s\n\n", args[0])
		printReserveUsage()
		os.Exit(1)
	}
}

func printReserveUsage() {
	fmt.Print(`Usage: awkto reserve <action> [options]

Actions:
  list                                     List all DHCP reservations
  create  -ip <addr> -mac <addr> -hostname <name>  Create a reservation
  delete  -ip <addr>                       Delete a reservation
  edit    -ip <addr> [-mac <addr>] [-hostname <name>]  Edit a reservation

Flags:
  -ip         IP address
  -mac        MAC address (e.g. 52:54:00:ab:cd:ef)
  -hostname   Hostname
  -subnet     Subnet ID (default: from AWKTO_SUBNET_ID or 1)
`)
}

func reserveListCmd(c *client.KeaClient, args []string) {
	fs := flag.NewFlagSet("reserve list", flag.ExitOnError)
	subnet := fs.String("subnet", cfg.SubnetID, "Subnet ID")
	fs.Parse(args)

	reservations, err := c.ListReservations(*subnet)
	if err != nil {
		exitErr(err)
	}

	if len(reservations) == 0 {
		fmt.Println("No reservations found.")
		return
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
	fmt.Fprintln(w, "IP ADDRESS\tMAC ADDRESS\tHOSTNAME\tSUBNET")
	for _, r := range reservations {
		fmt.Fprintf(w, "%s\t%s\t%s\t%d\n", r.IPAddress, r.HWAddress, r.Hostname, r.SubnetID)
	}
	w.Flush()
}

func reserveCreateCmd(c *client.KeaClient, args []string) {
	fs := flag.NewFlagSet("reserve create", flag.ExitOnError)
	ip := fs.String("ip", "", "IP address")
	mac := fs.String("mac", "", "MAC address")
	hostname := fs.String("hostname", "", "Hostname")
	subnet := fs.String("subnet", cfg.SubnetID, "Subnet ID")
	fs.Parse(args)

	if *ip == "" || *mac == "" || *hostname == "" {
		fmt.Fprintln(os.Stderr, "Error: -ip, -mac, and -hostname are all required")
		fs.Usage()
		os.Exit(1)
	}

	subnetID, _ := strconv.Atoi(*subnet)
	err := c.CreateReservation(client.ReservationCreate{
		IPAddress: *ip,
		HWAddress: *mac,
		Hostname:  *hostname,
		SubnetID:  subnetID,
	})
	if err != nil {
		exitErr(err)
	}
	fmt.Printf("Created reservation: %s (%s) -> %s\n", *ip, *mac, *hostname)
}

func reserveDeleteCmd(c *client.KeaClient, args []string) {
	fs := flag.NewFlagSet("reserve delete", flag.ExitOnError)
	ip := fs.String("ip", "", "IP address")
	fs.Parse(args)

	if *ip == "" {
		fmt.Fprintln(os.Stderr, "Error: -ip is required")
		fs.Usage()
		os.Exit(1)
	}

	if err := c.DeleteReservation(*ip); err != nil {
		exitErr(err)
	}
	fmt.Printf("Deleted reservation: %s\n", *ip)
}

func reserveEditCmd(c *client.KeaClient, args []string) {
	fs := flag.NewFlagSet("reserve edit", flag.ExitOnError)
	ip := fs.String("ip", "", "IP address of existing reservation")
	mac := fs.String("mac", "", "New MAC address")
	hostname := fs.String("hostname", "", "New hostname")
	subnet := fs.String("subnet", cfg.SubnetID, "Subnet ID")
	fs.Parse(args)

	if *ip == "" {
		fmt.Fprintln(os.Stderr, "Error: -ip is required")
		fs.Usage()
		os.Exit(1)
	}

	if *mac == "" && *hostname == "" {
		fmt.Fprintln(os.Stderr, "Error: at least -mac or -hostname must be provided")
		fs.Usage()
		os.Exit(1)
	}

	// Get current reservation to fill in unchanged fields
	reservations, err := c.ListReservations(*subnet)
	if err != nil {
		exitErr(err)
	}

	var current *client.Reservation
	for i := range reservations {
		if reservations[i].IPAddress == *ip {
			current = &reservations[i]
			break
		}
	}
	if current == nil {
		exitErr(fmt.Errorf("no reservation found for IP %s", *ip))
	}

	// Delete and recreate with updated fields
	if err := c.DeleteReservation(*ip); err != nil {
		exitErr(err)
	}

	newMAC := current.HWAddress
	if *mac != "" {
		newMAC = *mac
	}
	newHostname := current.Hostname
	if *hostname != "" {
		newHostname = *hostname
	}

	subnetID, _ := strconv.Atoi(*subnet)
	err = c.CreateReservation(client.ReservationCreate{
		IPAddress: *ip,
		HWAddress: newMAC,
		Hostname:  newHostname,
		SubnetID:  subnetID,
	})
	if err != nil {
		exitErr(err)
	}
	fmt.Printf("Updated reservation: %s (%s) -> %s\n", *ip, newMAC, newHostname)
}
