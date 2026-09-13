package main

import (
	"net/url"
	"os"

	"github.com/google/go-querystring/query"
	"github.com/kakeetopius/net-tools/internal/util"
	"github.com/pkg/browser"
	"github.com/spf13/pflag"
)

var googleURL = "https://google.com"

type queryParams struct {
	Query []string `url:"q,space"`
}

func main() {
	args, err := parseArgs()
	util.CheckErr(err)

	url, err := url.Parse(googleURL)
	util.CheckErr(err)

	if len(args) != 0 {
		url = url.JoinPath("search")
		params, qerr := query.Values(queryParams{Query: args})
		util.CheckErr(qerr)

		url.RawQuery = params.Encode()
	}

	err = browser.OpenURL(url.String())
	util.CheckErr(err)
}

func parseArgs() ([]string, error) {
	flagSet := pflag.NewFlagSet("server", pflag.ExitOnError)
	flagSet.Usage = util.UsageFunc("google", "queries...", flagSet.FlagUsages(), "Carry out google searches from the terminal. A browser is opened to show the results.")

	err := flagSet.Parse(os.Args[1:])
	if err != nil {
		return nil, err
	}

	return flagSet.Args(), nil
}
