package main

import (
	"fmt"
	"net"
	"os"

	"github.com/jsimonetti/rtnetlink/rtnl"
	"github.com/spf13/pflag"
)

type Options struct {
	Interface string
	Mac       string
}

func main() {
	args, err := parseArgs()
	if err != nil {
		if err != pflag.ErrHelp {
			fmt.Println(err)
		}
		return
	}

	iface, err := net.InterfaceByName(args.Interface)
	if err != nil {
		fmt.Printf("Error %v\n", err)
		return
	}

	mac, err := net.ParseMAC(args.Mac)
	if err != nil {
		fmt.Printf("Error %v\n", err)
		return
	}

	conn, err := rtnl.Dial(nil)
	if err != nil {
		fmt.Printf("Error %v\n", err)
		return
	}
	defer conn.Close()

	fmt.Println("Current MAC: ", iface.HardwareAddr.String())

	fmt.Println("Setting interface down....")
	err = conn.LinkDown(iface)
	if err != nil {
		fmt.Printf("Error %v\n", err)
		return
	}

	fmt.Println("Changing mac address for Interface ", iface.Name, " to ", args.Mac)
	err = conn.LinkSetHardwareAddr(iface, mac)
	if err != nil {
		fmt.Printf("Error %v\n", err)
		return
	}

	fmt.Println("Setting interface up....")
	err = conn.LinkUp(iface)
	if err != nil {
		fmt.Printf("Error %v\n", err)
		return
	}
	fmt.Println("Successful")
}

func parseArgs() (*Options, error) {
	flagSet := pflag.NewFlagSet("changemac", pflag.ContinueOnError)
	iface := flagSet.StringP("iface", "i", "", "The interface to change mac address for")
	mac := flagSet.StringP("mac", "m", "", "The Mac address to set to the interface")

	err := flagSet.Parse(os.Args[1:])
	if err != nil {
		return nil, err
	}
	if !flagSet.Changed("iface") {
		return nil, fmt.Errorf("no interface given. Use changemac -h for help")
	}
	if !flagSet.Changed("mac") {
		return nil, fmt.Errorf("no mac address given. Use changemac -h for help")
	}

	return &Options{
		Interface: *iface,
		Mac:       *mac,
	}, nil
}
