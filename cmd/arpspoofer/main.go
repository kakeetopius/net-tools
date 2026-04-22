package main

import (
	"errors"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/kakeetopius/net-tools/internal/util"
	"github.com/kakeetopius/net-tools/pkg/arpspoofer"
	"github.com/spf13/pflag"
)

type CLIOptions struct {
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

	err = arpspoofer.Spoof(arpspoofer.SpoofOptions{
		TargetIP:      opts.Target,
		TargetMac:     opts.TargetMac,
		SourceIP:      opts.Source,
		SleepDuration: opts.SleepDuration,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v", err)
	}
}

func parseArgs() (*CLIOptions, error) {
	flagSet := pflag.NewFlagSet("arpspoofer", pflag.ExitOnError)
	target := flagSet.IPP("target", "t", nil, "The IPv4 address of the target")
	targetMac := flagSet.StringP("target-mac", "m", "", "The MAC address of the target.")
	source := flagSet.IPP("source", "s", nil, "The source address to pretend to be.")
	duration := flagSet.DurationP("sleep-duration", "d", 2*time.Second, "The period of time in milliseconds to pause for when sending ARP replies")

	flagSet.Usage = util.UsageFunc("arpspoofer", "", flagSet.FlagUsages(), "Carry out ARP spoof attacks.")
	err := flagSet.Parse(os.Args[1:])
	if err != nil {
		return nil, err
	}

	if !flagSet.Changed("target") {
		return nil, fmt.Errorf("please provide IP address of target")
	}
	if !flagSet.Changed("source") {
		return nil, fmt.Errorf("please provide source IP address you want to pretend to be")
	}
	if !flagSet.Changed("target-mac") {
		return nil, fmt.Errorf("please provide the target's mac address")
	}

	mac, err := net.ParseMAC(*targetMac)
	if err != nil {
		return nil, err
	}

	return &CLIOptions{
		Target:        *target,
		TargetMac:     mac,
		Source:        *source,
		SleepDuration: *duration,
	}, nil
}
