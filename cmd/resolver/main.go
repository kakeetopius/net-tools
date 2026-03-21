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
	Query   string   `json:"query"`
	Answers []string `json:"answers"`
	Err     string   `json:"error,omitempty"`
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

	for _, name := range queries {
		addresses, err := resolver.LookupHost(ctx, name)
		var errStr string
		if err != nil {
			errStr = err.Error()
		}
		results = append(results, Result{
			Query:   name,
			Answers: addresses,
			Err:     errStr,
		})
	}
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
			Query:   addr,
			Answers: names,
			Err:     errStr,
		})
	}
	return results, nil
}

func printResults(results []Result, opts *Options) error {
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

		answerString := strings.Builder{}
		for _, answer := range result.Answers {
			fmt.Fprintf(&answerString, "%v\n", answer)
		}
		data = append(data, []string{result.Query, answerString.String()})
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
	flagSet.Usage = util.UsageFunc("resolver", "queries", flagSet.FlagUsages())

	err := flagSet.Parse(os.Args[1:])
	if err != nil {
		return nil, Options{}, err
	}
	return flagSet.Args(), Options{
		ReverseLookup: *reverseLookup,
		PrintJSON:     *json,
	}, nil
}
