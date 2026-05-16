// Package util contains some helper functions
package util

import (
	"errors"
	"fmt"
	"net"
	"os"

	"github.com/spf13/pflag"
)

var ErrUserQuit = errors.New("user quit")

// GetIfaceByIP returns the first network interface that has an IP network
// containing the provided IP address.
func GetIfaceByIP(IPAddr net.IP) (*net.Interface, error) {
	allIfaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	for _, iface := range allIfaces {
		addrs, err := iface.Addrs()
		if err != nil {
			return nil, err
		}
		for _, addr := range addrs {
			addr, ok := addr.(*net.IPNet)
			if !ok {
				return nil, fmt.Errorf("error parsing Interface IP address")
			}
			if addr.Contains(IPAddr) {
				return &iface, nil
			}
		}
	}

	return nil, fmt.Errorf("no interface connected to that network")
}

// UsageFunc returns a function that prints usage, description, and option
// information for a command.
func UsageFunc(commandName, positionalArgsName, flagHelpOutput, description string) func() {
	return func() {
		if positionalArgsName != "" && flagHelpOutput != "" {
			fmt.Printf("Usage: %s [%s] [OPTIONS]\n", commandName, positionalArgsName)
		} else if positionalArgsName != "" {
			fmt.Printf("Usage: %s [%s]\n", commandName, positionalArgsName)
		} else if flagHelpOutput != "" {
			fmt.Printf("Usage: %s [OPTIONS]\n", commandName)
		} else {
			fmt.Printf("Usage: %s\n", commandName)
		}

		if description != "" {
			fmt.Println("\nDescription: ")
			fmt.Println("  ", description)
		}
		if flagHelpOutput != "" {
			fmt.Println("\nOptions: ")
			fmt.Println(flagHelpOutput)
		}
	}
}

// CheckErr exits the program with a non-zero status when err is not nil.
// pflag.ErrHelp is treated specially and exits cleanly without printing an error.
func CheckErr(err error) {
	if err != nil {
		returnCode := 0
		if !errors.Is(err, pflag.ErrHelp) && !errors.Is(err, ErrUserQuit) {
			// no need to print to error message for the above
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			returnCode = -1
		}
		os.Exit(returnCode)
	}
}
