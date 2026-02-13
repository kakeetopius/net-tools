package main

import (
	"errors"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/kakeetopius/net-tools/pkg/arpspoofer"
	"github.com/spf13/pflag"
)

type Options struct {
	Target        net.IP
	TargetMac     net.HardwareAddr
	Source        net.IP
	SleepDuration time.Duration
}

func main() {
	opts, err := parseArgs()
	if err != nil {
		if !errors.Is(err, pflag.ErrHelp) {
			fmt.Fprintln(os.Stderr, err)
		}
		return
	}

	err = arpspoofer.Spoof(opts.Target, opts.TargetMac, opts.Source, opts.SleepDuration)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v", err)
	}
	fmt.Println("We are spoofers!!")
}

func parseArgs() (*Options, error) {
	flagSet := pflag.NewFlagSet("arpspoofer", pflag.ExitOnError)
	target := flagSet.IPP("target", "t", nil, "The IPv4 address of the target")
	targetMac := flagSet.StringP("target-mac", "m", "", "The MAC address of the target.")
	source := flagSet.IPP("source", "s", nil, "The source address to pretend to be.")
	duration := flagSet.DurationP("sleep-duration", "d", 2*time.Second, "The period of time in milliseconds to pause for when sending ARP replies")

	err := flagSet.Parse(os.Args[1:])
	if err != nil {
		return nil, err
	}

	if !flagSet.Changed("target") {
		return nil, fmt.Errorf("please provide IP address of target. Use arpspoofer -h for more information")
	}
	if !flagSet.Changed("source") {
		return nil, fmt.Errorf("please provide source IP address. Use arpspoofer -h for more information")
	}
	if !flagSet.Changed("target-mac") {
		return nil, fmt.Errorf("please provide source target mac address. Use arpspoofer -h for more information")
	}

	mac, err := net.ParseMAC(*targetMac)
	if err != nil {
		return nil, err
	}

	return &Options{
		Target:        *target,
		TargetMac:     mac,
		Source:        *source,
		SleepDuration: *duration,
	}, nil
}
