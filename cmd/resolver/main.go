package main

import (
	"context"
	"fmt"
	"net"
	"os"

	"github.com/spf13/pflag"
)

type Result struct {
	DomainName string
	Addresses  []string
	Err        error
}

func main() {
	names, err := parseArgs()
	if err != nil {
		fmt.Println("Error: ", err)
		return
	}
	results, err := resolve(names)
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
			DomainName: name,
			Addresses:  addresses,
			Err:        err,
		})
	}
	return results, nil
}

func printResults(results []Result) {
	for _, result := range results {
		fmt.Printf("Results for: %v\n", result.DomainName)
		if result.Err != nil {
			fmt.Printf("Got Error: %v\n\n", result.Err)
			continue
		}
		for _, ip := range result.Addresses {
			fmt.Println(ip)
		}
		fmt.Println()
	}
}

func parseArgs() ([]string, error) {
	flagSet := pflag.NewFlagSet("resolver", pflag.ExitOnError)
	err := flagSet.Parse(os.Args[1:])
	if err != nil {
		return nil, err
	}
	return flagSet.Args(), nil
}
