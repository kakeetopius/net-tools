package main

import (
	"context"
	"fmt"
	"net"
	"os"

	"github.com/spf13/pflag"
)

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

func resolve(queries []string) (map[string][]string, error) {
	resolver := net.Resolver{
		PreferGo: true,
	}
	results := make(map[string][]string)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	for _, name := range queries {
		hostNames, err := resolver.LookupHost(ctx, name)
		if err != nil {
			return nil, err
		}
		results[name] = hostNames
	}
	return results, nil
}

func printResults(results map[string][]string) {
	for name, ips := range results {
		fmt.Printf("Results for: %v\n", name)
		for _, ip := range ips {
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
