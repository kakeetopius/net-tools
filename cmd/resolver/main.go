package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/pterm/pterm"
	"github.com/spf13/pflag"
)

type Result struct {
	Query   string
	Answers []string
	Err     error
}

type Options struct {
	ReverseLookup bool
}

func main() {
	var results []Result
	var err error
	queries, opts, err := parseArgs()
	if err != nil {
		fmt.Println("Error: ", err)
		return
	}

	if opts.ReverseLookup {
		results, err = reverseLookup(queries)
	} else {
		results, err = resolve(queries)
	}
	if err != nil {
		fmt.Println("Error: ", err)
		return
	}
	printResults(results)
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
		results = append(results, Result{
			Query:   name,
			Answers: addresses,
			Err:     err,
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
		results = append(results, Result{
			Query:   addr,
			Answers: names,
			Err:     err,
		})
	}
	return results, nil
}

func printResults(results []Result) {
	data := pterm.TableData{{"Query", "Response(s)", "Errors"}}
	for _, result := range results {
		answerString := strings.Builder{}
		for _, answer := range result.Answers {
			fmt.Fprintf(&answerString, "%v\n", answer)
		}
		var errorStr string
		if result.Err != nil {
			errorStr = result.Err.Error()
		} else {
			errorStr = "None"
		}
		data = append(data, []string{result.Query, answerString.String(), errorStr})
	}

	pterm.DefaultTable.WithHasHeader(true).WithBoxed(true).WithRowSeparator("-").WithHeaderRowSeparator("-").WithData(data).Render()
}

func parseArgs() ([]string, Options, error) {
	flagSet := pflag.NewFlagSet("resolver", pflag.ExitOnError)
	reverseLookup := flagSet.BoolP("reverse", "r", false, "Perform a reverse lookup for the given IP(s).")
	err := flagSet.Parse(os.Args[1:])
	if err != nil {
		return nil, Options{}, err
	}
	return flagSet.Args(), Options{
		ReverseLookup: *reverseLookup,
	}, nil
}
