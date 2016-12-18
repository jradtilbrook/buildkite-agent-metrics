package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/buildkite/buildkite-metrics/collector"
	"gopkg.in/buildkite/go-buildkite.v2/buildkite"
)

// Version is passed in via ldflags
var Version string

func main() {
	var (
		accessToken = flag.String("token", "", "A Buildkite API Access Token")
		orgSlug     = flag.String("org", "", "A Buildkite Organization Slug")
		interval    = flag.Duration("interval", 0, "Update metrics every interval, rather than once")
		history     = flag.Duration("history", time.Hour*24, "Historical data to use for finished builds")
		debug       = flag.Bool("debug", false, "Show debug output")
		version     = flag.Bool("version", false, "Show the version")
		quiet       = flag.Bool("quiet", false, "Only print errors")
		dryRun      = flag.Bool("dry-run", false, "Whether to only print metrics")

		// filters
		queue = flag.String("queue", "", "Only include a specific queue")
	)

	flag.Parse()

	if *version {
		fmt.Printf("buildkite-metrics %s\n", Version)
		os.Exit(0)
	}

	if *accessToken == "" {
		fmt.Println("Must provide a value for -token")
		os.Exit(1)
	}

	if *orgSlug == "" {
		fmt.Println("Must provide a value for -org")
		os.Exit(1)
	}

	if *quiet {
		log.SetOutput(ioutil.Discard)
	}

	config, err := buildkite.NewTokenConfig(*accessToken, false)
	if err != nil {
		fmt.Printf("client config failed: %s\n", err)
		os.Exit(1)
	}

	client := buildkite.NewClient(config.Client())
	if *debug && os.Getenv("TRACE_HTTP") != "" {
		buildkite.SetHttpDebug(*debug)
	}

	col := collector.New(client, collector.Opts{
		OrgSlug:    *orgSlug,
		Historical: *history,
		Queue:      *queue,
		Debug:      *debug,
	})

	f := func() error {
		t := time.Now()

		res, err := col.Collect()
		if err != nil {
			return err
		}

		if !*quiet {
			res.Dump()
		}

		if !*dryRun {
			err = cloudwatchSend(res)
			if err != nil {
				return err
			}
		}

		log.Printf("Finished in %s", time.Now().Sub(t))
		return nil
	}

	if err := f(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if *interval > 0 {
		for _ = range time.NewTicker(*interval).C {
			if err := f(); err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		}
	}
}
