package main

import (
	"crypto/rand"
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
	Random    bool
}

func main() {
	opts, err := parseArgs()
	util.CheckErr(err)

	var mac net.HardwareAddr
	if opts.Random {
		mac, err = genRandomMac()
		util.CheckErr(err)
	} else if opts.Mac != "" {
		mac, err = net.ParseMAC(opts.Mac)
		util.CheckErr(err)
	} else {
		fmt.Fprintln(os.Stderr, "please provide a new mac address. Use changemac -h for more information.")
		return
	}

	iface, err := net.InterfaceByName(opts.Interface)
	util.CheckErr(err)

	err = changeMac(iface, mac)
	util.CheckErr(err)
}

func changeMac(iface *net.Interface, mac net.HardwareAddr) error {
	conn, err := rtnl.Dial(nil)
	if err != nil {
		return err
	}
	defer conn.Close()

	fmt.Println("Current MAC: ", iface.HardwareAddr.String())

	fmt.Println("Setting interface down....")
	err = conn.LinkDown(iface)
	if err != nil {
		return err
	}

	fmt.Println("Changing mac address for Interface ", iface.Name, " to ", mac)
	err = conn.LinkSetHardwareAddr(iface, mac)
	if err != nil {
		return err
	}

	fmt.Println("Setting interface up....")
	err = conn.LinkUp(iface)
	if err != nil {
		return err
	}

	fmt.Println("MAC successfully changed")
	return nil
}

func genRandomMac() ([]byte, error) {
	mac := make([]byte, 6)

	_, err := rand.Read(mac)
	if err != nil {
		return nil, err
	}

	// clear multicast (bit 0 - LSB)
	mac[0] &= 0xfe
	// set locally administered
	mac[0] |= 0x02

	return mac, nil
}

func parseArgs() (*Options, error) {
	opts := Options{}

	flagSet := pflag.NewFlagSet("changemac", pflag.ContinueOnError)
	flagSet.StringVarP(&opts.Interface, "iface", "i", "", "The interface to change mac address for")
	flagSet.StringVarP(&opts.Mac, "mac", "m", "", "The Mac address to set on the interface")
	flagSet.BoolVarP(&opts.Random, "random", "r", false, "Generate new random mac address.")

	flagSet.Usage = util.UsageFunc("changemac", "", flagSet.FlagUsages(), "Change the mac address of a linux network interface")

	err := flagSet.Parse(os.Args[1:])
	if err != nil {
		return nil, err
	}

	if opts.Interface == "" {
		return nil, fmt.Errorf("no interface given. Use changemac -h for help")
	}
	if opts.Mac == "" && !opts.Random {
		return nil, fmt.Errorf("please provide new mac address")
	}

	return &opts, nil
}
