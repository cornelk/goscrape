package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/alexflint/go-arg"
	"github.com/cornelk/goscrape/scraper"
	"github.com/cornelk/gotokit/env"
	"github.com/cornelk/gotokit/log"
)

const toolName = "goscrape"

type arguments struct {
	Exclude []string `arg:"-n,--include" help:"only include URLs with PERL Regular Expressions support"`
	Include []string `arg:"-x,--exclude" help:"exclude URLs with PERL Regular Expressions support"`
	Output  string   `arg:"-o,--output" help:"output directory to write files to"`
	URLs    []string `arg:"positional,required"`

	Depth        int64 `arg:"-d,--depth" help:"download depth, 0 for unlimited" default:"10"`
	ImageQuality int64 `arg:"-i,--imagequality" help:"image quality, 0 to disable reencoding"`
	Timeout      int64 `arg:"-t,--timeout" help:"time limit in seconds for each HTTP request to connect and read the request body"`

	Proxy string `arg:"-p,--proxy" help:"HTTP proxy to use for scraping"`
	User  string `arg:"-u,--user" help:"user[:password] to use for authentication"`

	Verbose bool `arg:"-v,--verbose" help:"verbose output"`
}

func (arguments) Description() string {
	return "Scrape a website and create an offline browsable version on the disk.\n"
}

func main() {
	var args arguments
	arg.MustParse(&args)

	if err := run(args); err != nil {
		fmt.Printf("Execution error: %s\n", err)
		os.Exit(1)
	}
}

func run(args arguments) error {
	if len(args.URLs) == 0 {
		return nil
	}

	var username, password string
	if args.User != "" {
		sl := strings.Split(args.User, ":")
		username = sl[0]
		if len(sl) > 1 {
			password = sl[1]
		}
	}

	imageQuality := args.ImageQuality
	if args.ImageQuality < 0 || args.ImageQuality >= 100 {
		imageQuality = 0
	}

	logger, err := createLogger()
	if err != nil {
		return fmt.Errorf("creating logger: %w", err)
	}

	cfg := scraper.Config{
		Includes:        args.Include,
		Excludes:        args.Exclude,
		ImageQuality:    uint(imageQuality),
		MaxDepth:        uint(args.Depth),
		Timeout:         uint(args.Timeout),
		OutputDirectory: args.Output,
		Username:        username,
		Password:        password,
		Proxy:           args.Proxy,
	}

	for _, url := range args.URLs {
		cfg.URL = url
		sc, err := scraper.New(logger, cfg)
		if err != nil {
			return fmt.Errorf("initializing scraper: %w", err)
		}

		logger.Info("Scraping", log.Stringer("URL", sc.URL))
		if err = sc.Start(); err != nil {
			return fmt.Errorf("scraping '%s': %w", sc.URL, err)
		}
	}

	return nil
}

func createLogger() (*log.Logger, error) {
	logCfg, err := log.ConfigForEnv(env.Development)
	if err != nil {
		return nil, fmt.Errorf("initializing log config: %w", err)
	}
	logCfg.JSONOutput = false
	logCfg.CallerInfo = false

	logger, err := log.NewWithConfig(logCfg)
	if err != nil {
		return nil, fmt.Errorf("initializing logger: %w", err)
	}
	logger = logger.Named(toolName)
	return logger, nil
}
