package cmd

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"text/tabwriter"

	"github.com/awkto/awkto-cli/internal/client"
)

func runLease(args []string) {
	if len(args) < 1 {
		printLeaseUsage()
		os.Exit(1)
	}

	if err := cfg.RequireKea(); err != nil {
		exitErr(err)
	}
	c := client.NewKeaClient(cfg)

	switch args[0] {
	case "list":
		leaseListCmd(c, args[1:])
	case "delete":
		leaseDeleteCmd(c, args[1:])
	case "promote":
		leasePromoteCmd(c, args[1:])
	case "help", "--help", "-h":
		printLeaseUsage()
	default:
		fmt.Fprintf(os.Stderr, "Unknown lease action: %s\n\n", args[0])
		printLeaseUsage()
		os.Exit(1)
	}
}

func printLeaseUsage() {
	fmt.Print(`Usage: awkto lease <action> [options]

Actions:
  list                          List all DHCP leases
  delete   -ip <addr> | -mac <addr>   Delete a lease by IP or MAC
  promote  -ip <addr> [-hostname <name>]  Promote a lease to a reservation

Flags:
  -ip         IP address
  -mac        MAC address
  -hostname   Hostname for promoted reservation
  -subnet     Subnet ID (default: from AWKTO_SUBNET_ID or 1)
`)
}

func leaseListCmd(c *client.KeaClient, args []string) {
	fs := flag.NewFlagSet("lease list", flag.ExitOnError)
	subnet := fs.String("subnet", cfg.SubnetID, "Subnet ID")
	fs.Parse(args)

	leases, err := c.ListLeases(*subnet)
	if err != nil {
		exitErr(err)
	}

	if len(leases) == 0 {
		fmt.Println("No leases found.")
		return
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
	fmt.Fprintln(w, "IP ADDRESS\tMAC ADDRESS\tHOSTNAME\tSTATE\tSUBNET")
	for _, l := range leases {
		state := "active"
		if l.State == 1 {
			state = "declined"
		} else if l.State == 2 {
			state = "expired"
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%d\n", l.IPAddress, l.HWAddress, l.Hostname, state, l.SubnetID)
	}
	w.Flush()
}

func leaseDeleteCmd(c *client.KeaClient, args []string) {
	fs := flag.NewFlagSet("lease delete", flag.ExitOnError)
	ip := fs.String("ip", "", "IP address")
	mac := fs.String("mac", "", "MAC address")
	fs.Parse(args)

	if *ip == "" && *mac == "" {
		fmt.Fprintln(os.Stderr, "Error: -ip or -mac is required")
		fs.Usage()
		os.Exit(1)
	}

	if *ip != "" {
		if err := c.DeleteLeaseByIP(*ip); err != nil {
			exitErr(err)
		}
		fmt.Printf("Deleted lease for IP: %s\n", *ip)
	}
	if *mac != "" {
		if err := c.DeleteLeaseByMAC(*mac); err != nil {
			exitErr(err)
		}
		fmt.Printf("Deleted lease for MAC: %s\n", *mac)
	}
}

func leasePromoteCmd(c *client.KeaClient, args []string) {
	fs := flag.NewFlagSet("lease promote", flag.ExitOnError)
	ip := fs.String("ip", "", "IP address of the lease to promote")
	hostname := fs.String("hostname", "", "Hostname for the reservation")
	subnet := fs.String("subnet", cfg.SubnetID, "Subnet ID")
	fs.Parse(args)

	if *ip == "" {
		fmt.Fprintln(os.Stderr, "Error: -ip is required")
		fs.Usage()
		os.Exit(1)
	}

	// Find the lease to get its MAC address
	leases, err := c.ListLeases(*subnet)
	if err != nil {
		exitErr(err)
	}

	var found *client.Lease
	for i := range leases {
		if leases[i].IPAddress == *ip {
			found = &leases[i]
			break
		}
	}
	if found == nil {
		exitErr(fmt.Errorf("no lease found for IP %s", *ip))
	}

	name := *hostname
	if name == "" {
		name = found.Hostname
	}
	if name == "" {
		exitErr(fmt.Errorf("no hostname found on lease and none provided with -hostname"))
	}

	subnetID, _ := strconv.Atoi(*subnet)
	err = c.CreateReservation(client.ReservationCreate{
		IPAddress: found.IPAddress,
		HWAddress: found.HWAddress,
		Hostname:  name,
		SubnetID:  subnetID,
	})
	if err != nil {
		exitErr(err)
	}

	fmt.Printf("Promoted lease to reservation: %s (%s) -> %s\n", found.IPAddress, found.HWAddress, name)
}
