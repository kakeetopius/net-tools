package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/spf13/pflag"
)

type Options struct {
	Directory string
	Address   string
	Port      int
}

func main() {
	opts, err := parseArgs()
	if err != nil {
		fmt.Println(err)
		return
	}

	log.Printf("Serving files in directory: %v", opts.Directory)
	log.Printf("Starting Server on %v:%v", opts.Address, opts.Port)

	fileServer := http.FileServer(http.Dir(opts.Directory))
	http.Handle("/", fileServer)

	address := fmt.Sprintf("%v:%v", opts.Address, opts.Port)
	err = http.ListenAndServe(address, nil)
	if err != nil {
		log.Println(err)
	}
}

func parseArgs() (*Options, error) {
	flagSet := pflag.NewFlagSet("server", pflag.ExitOnError)
	dir := flagSet.StringP("dir", "d", ".", "The directory to serve http files from.")
	port := flagSet.IntP("port", "p", 4020, "The port to listen on.")
	address := flagSet.IPP("address", "a", nil, "The address to listen on. (Default is all address)")

	fmt.Println(len(os.Args))
	err := flagSet.Parse(os.Args[1:])
	if err != nil {
		return nil, err
	}
	var addStr string
	if *address != nil {
		addStr = address.String()
	}

	return &Options{
		Directory: *dir,
		Port:      *port,
		Address:   addStr,
	}, nil
}
