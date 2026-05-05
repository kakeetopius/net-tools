package main

import (
	"fmt"
	"net"
	"os"

	"github.com/jsimonetti/rtnetlink/rtnl"
	"github.com/kakeetopius/net-tools/internal/util"
	"github.com/spf13/pflag"
)

type Options struct {
	Interface string
	Mac       string
}

func main() {
	args, err := parseArgs()
	util.CheckErr(err)

	iface, err := net.InterfaceByName(args.Interface)
	util.CheckErr(err)

	mac, err := net.ParseMAC(args.Mac)
	util.CheckErr(err)

	conn, err := rtnl.Dial(nil)
	util.CheckErr(err)
	defer conn.Close()

	fmt.Println("Current MAC: ", iface.HardwareAddr.String())

	fmt.Println("Setting interface down....")
	err = conn.LinkDown(iface)
	util.CheckErr(err)

	fmt.Println("Changing mac address for Interface ", iface.Name, " to ", args.Mac)
	err = conn.LinkSetHardwareAddr(iface, mac)
	util.CheckErr(err)

	fmt.Println("Setting interface up....")
	err = conn.LinkUp(iface)
	util.CheckErr(err)

	fmt.Println("Successful")
}

func parseArgs() (*Options, error) {
	flagSet := pflag.NewFlagSet("changemac", pflag.ContinueOnError)
	iface := flagSet.StringP("iface", "i", "", "The interface to change mac address for")
	mac := flagSet.StringP("mac", "m", "", "The Mac address to set on the interface")
	flagSet.Usage = util.UsageFunc("changemac", "", flagSet.FlagUsages(), "Change the mac address of a linux network interface")

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
