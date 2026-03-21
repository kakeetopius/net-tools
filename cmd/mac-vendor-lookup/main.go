package main

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/endobit/oui"
	"github.com/pterm/pterm"
	"github.com/spf13/pflag"
)

type Result struct {
	MAC    string `json:"mac"`
	Vendor string `json:"vendor"`
	Err    string `json:"error,omitempty"`
}

type Options struct {
	PrintJSON bool
}

func main() {
	macs, opts, err := parseArgs()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return
	}

	results := lookupVendors(macs)
	err = printResults(results, opts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return
	}
}

func lookupVendors(macs []string) []Result {
	results := make([]Result, 0, len(macs))
	for _, macStr := range macs {
		mac, err := net.ParseMAC(macStr)
		var errStr string
		if err != nil {
			errStr = err.Error()
		}
		vendor := oui.VendorFromMAC(mac)
		results = append(results, Result{
			MAC:    mac.String(),
			Vendor: vendor,
			Err:    errStr,
		})
	}
	return results
}

func printResults(results []Result, opts Options) error {
	if len(results) == 0 {
		return nil
	}
	if opts.PrintJSON {
		jsonResults, err := json.MarshalIndent(results, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(jsonResults))
		return nil
	}
	data := pterm.TableData{{"MAC", "Vendor"}}

	errStr := strings.Builder{}
	for _, result := range results {
		if result.Err != "" {
			fmt.Fprintf(&errStr, "%v\n", result.Err)
			continue
		}
		vendor := "unknown"
		if result.Vendor != "" {
			vendor = result.Vendor
		}
		data = append(data, []string{result.MAC, vendor})
	}

	table := pterm.DefaultTable.WithHasHeader(true).WithBoxed(true).WithRowSeparator("-").WithHeaderRowSeparator("-").WithData(data)
	if len(data) > 1 {
		table.Render()
	}

	errors := errStr.String()
	if errors != "" {
		fmt.Println("\nErrors: ")
		fmt.Println(errors)
	}
	return nil
}

func parseArgs() ([]string, Options, error) {
	flagSet := pflag.NewFlagSet("mvl", pflag.ExitOnError)
	json := flagSet.BoolP("json", "j", false, "Output the results in json form")
	err := flagSet.Parse(os.Args[1:])
	return flagSet.Args(), Options{PrintJSON: *json}, err
}
