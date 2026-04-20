package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/kakeetopius/net-tools/internal/util"
	"github.com/pterm/pterm"
	"github.com/spf13/pflag"
)

type Result struct {
	Query string   `json:"query"`
	Addrs []string `json:"addrs"`
	MX    []string `json:"mx"`
	NS    []string `json:"ns"`
	Err   string   `json:"error,omitempty"`
}

type Options struct {
	ReverseLookup bool
	PrintJSON     bool
}

func main() {
	var results []Result
	var err error
	queries, opts, err := parseArgs()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return
	}

	if opts.ReverseLookup {
		results, err = reverseLookup(queries)
	} else {
		results, err = resolve(queries)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return
	}
	err = printResults(results, &opts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
	}
}

func resolve(queries []string) ([]Result, error) {
	resolver := net.Resolver{
		PreferGo: true,
	}
	results := make([]Result, 0, len(queries))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	spinner, err := pterm.DefaultSpinner.Start("Resolving")
	if err != nil {
		return nil, err
	}
	for _, name := range queries {
		addresses, err := resolver.LookupHost(ctx, name)
		var mxNames []string
		var nsNames []string

		if err != nil {
			results = append(results, Result{
				Query: name,
				Err:   err.Error(),
			})
			continue
		}
		mx, err := resolver.LookupMX(ctx, name)
		if err != nil {
			results = append(results, Result{
				Query: name,
				Err:   err.Error(),
			})
			continue
		}
		for _, m := range mx {
			mxNames = append(mxNames, fmt.Sprintf("%v Pref(%v)", m.Host, m.Pref))
		}
		ns, err := resolver.LookupNS(ctx, name)
		for _, n := range ns {
			nsNames = append(nsNames, fmt.Sprintf("%v ", n.Host))
		}
		if err != nil {
			results = append(results, Result{
				Query: name,
				Err:   err.Error(),
			})
			continue
		}
		results = append(results, Result{
			Query: name,
			Addrs: addresses,
			MX:    mxNames,
			NS:    nsNames,
			Err:   "",
		})
	}
	spinner.Stop()
	return results, nil
}

func reverseLookup(queries []string) ([]Result, error) {
	resolver := net.Resolver{
		PreferGo: true,
	}
	results := make([]Result, 0, len(queries))
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	for _, addr := range queries {
		names, err := resolver.LookupAddr(ctx, addr)
		var errStr string
		if err != nil {
			errStr = err.Error()
		}
		results = append(results, Result{
			Query: addr,
			Addrs: names,
			Err:   errStr,
		})
	}
	return results, nil
}

func printResults(results []Result, opts *Options) error {
	if opts.ReverseLookup {
		return printReverseLookupResults(results, opts)
	}
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

	data := pterm.TableData{{"Type", "Response(s)"}}
	errStr := strings.Builder{}
	for _, result := range results {
		if result.Err != "" {
			fmt.Fprintf(&errStr, "%v\n", result.Err)
			continue
		}

		addrString := strings.Builder{}
		mxString := strings.Builder{}
		nsString := strings.Builder{}
		for _, answer := range result.Addrs {
			fmt.Fprintf(&addrString, "%v\n", answer)
		}
		for _, mx := range result.MX {
			fmt.Fprintf(&mxString, "%v\n", mx)
		}
		for _, ns := range result.NS {
			fmt.Fprintf(&nsString, "%v\n", ns)
		}
		style := pterm.NewStyle(pterm.Bold, pterm.FgLightBlue)
		name := style.Sprint(result.Query)
		data = append(data, []string{name})
		data = append(data, []string{"A or\nAAAA", addrString.String()})
		data = append(data, []string{"MX", mxString.String()})
		data = append(data, []string{"NS", nsString.String()})
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

func printReverseLookupResults(results []Result, opts *Options) error {
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

	data := pterm.TableData{{"Query", "Response(s)"}}
	errStr := strings.Builder{}
	for _, result := range results {
		if result.Err != "" {
			fmt.Fprintf(&errStr, "%v\n", result.Err)
			continue
		}

		addrString := strings.Builder{}
		for _, answer := range result.Addrs {
			fmt.Fprintf(&addrString, "%v\n", answer)
		}
		data = append(data, []string{result.Query, addrString.String()})
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
	flagSet := pflag.NewFlagSet("resolver", pflag.ExitOnError)
	reverseLookup := flagSet.BoolP("reverse", "r", false, "Perform a reverse lookup for the given IP(s).")
	json := flagSet.BoolP("json", "j", false, "Output the results in json form")
	flagSet.Usage = util.UsageFunc("resolver", "queries...", flagSet.FlagUsages())

	err := flagSet.Parse(os.Args[1:])
	if err != nil {
		return nil, Options{}, err
	}
	return flagSet.Args(), Options{
		ReverseLookup: *reverseLookup,
		PrintJSON:     *json,
	}, nil
}
