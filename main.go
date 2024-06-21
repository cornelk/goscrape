package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/alexflint/go-arg"
	"github.com/cornelk/goscrape/scraper"
	"github.com/cornelk/gotokit/app"
	"github.com/cornelk/gotokit/buildinfo"
	"github.com/cornelk/gotokit/env"
	"github.com/cornelk/gotokit/log"
)

var (
	version = "dev"
	commit  = ""
	date    = ""
)

type arguments struct {
	Exclude []string `arg:"-n,--include" help:"only include URLs with PERL Regular Expressions support"`
	Include []string `arg:"-x,--exclude" help:"exclude URLs with PERL Regular Expressions support"`
	Output  string   `arg:"-o,--output" help:"output directory to write files to"`
	URLs    []string `arg:"positional"`

	Depth        int64 `arg:"-d,--depth" help:"download depth, 0 for unlimited" default:"10"`
	ImageQuality int64 `arg:"-i,--imagequality" help:"image quality, 0 to disable reencoding"`
	Timeout      int64 `arg:"-t,--timeout" help:"time limit in seconds for each HTTP request to connect and read the request body"`

	Serve      string `arg:"-s,--serve" help:"serve the website using a webserver"`
	ServerPort int16  `arg:"-r,--serverport" help:"port to use for the webserver" default:"8080"`

	Headers   []string `arg:"-h,--header" help:"HTTP header to use for scraping"`
	Proxy     string   `arg:"-p,--proxy" help:"HTTP proxy to use for scraping"`
	User      string   `arg:"-u,--user" help:"user[:password] to use for authentication"`
	UserAgent string   `arg:"-a,--useragent" help:"user agent to use for scraping"`

	Verbose bool `arg:"-v,--verbose" help:"verbose output"`
}

func (arguments) Description() string {
	return "Scrape a website and create an offline browsable version on the disk.\n"
}

func (arguments) Version() string {
	return fmt.Sprintf("Version: %s\n", buildinfo.Version(version, commit, date))
}

func main() {
	args, err := readArguments()
	if err != nil {
		fmt.Printf("Reading arguments failed: %s\n", err)
		os.Exit(1)
	}

	ctx := app.Context()

	if args.Verbose {
		log.SetDefaultLevel(log.DebugLevel)
	}
	logger, err := createLogger()
	if err != nil {
		fmt.Printf("Creating logger failed: %s\n", err)
		os.Exit(1)
	}

	if args.Serve != "" {
		if err := runServer(ctx, args, logger); err != nil {
			fmt.Printf("Server execution error: %s\n", err)
			os.Exit(1)
		}
		return
	}

	if err := runScraper(ctx, args, logger); err != nil {
		fmt.Printf("Scraping execution error: %s\n", err)
		os.Exit(1)
	}
}

func readArguments() (arguments, error) {
	var args arguments
	parser, err := arg.NewParser(arg.Config{}, &args)
	if err != nil {
		return arguments{}, fmt.Errorf("creating argument parser: %w", err)
	}

	if err = parser.Parse(os.Args[1:]); err != nil {
		switch {
		case errors.Is(err, arg.ErrHelp):
			parser.WriteHelp(os.Stdout)
			os.Exit(0)
		case errors.Is(err, arg.ErrVersion):
			fmt.Println(args.Version())
			os.Exit(0)
		}

		return arguments{}, fmt.Errorf("parsing arguments: %w", err)
	}

	if len(args.URLs) == 0 && args.Serve == "" {
		parser.WriteHelp(os.Stdout)
		os.Exit(0)
	}

	return args, nil
}

// nolint: funlen
func runScraper(ctx context.Context, args arguments, logger *log.Logger) error {
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

	cfg := scraper.Config{
		Includes: args.Include,
		Excludes: args.Exclude,

		ImageQuality: uint(imageQuality),
		MaxDepth:     uint(args.Depth),
		Timeout:      uint(args.Timeout),

		OutputDirectory: args.Output,
		Username:        username,
		Password:        password,

		Header:    scraper.Headers(args.Headers),
		Proxy:     args.Proxy,
		UserAgent: args.UserAgent,
	}

	for _, url := range args.URLs {
		cfg.URL = url
		sc, err := scraper.New(logger, cfg)
		if err != nil {
			return fmt.Errorf("initializing scraper: %w", err)
		}

		logger.Info("Scraping", log.String("url", sc.URL.String()))
		if err = sc.Start(ctx); err != nil {
			if errors.Is(err, context.Canceled) {
				os.Exit(0)
			}

			return fmt.Errorf("scraping '%s': %w", sc.URL, err)
		}
	}

	return nil
}

func runServer(ctx context.Context, args arguments, logger *log.Logger) error {
	if err := scraper.ServeDirectory(ctx, args.Serve, args.ServerPort, logger); err != nil {
		return fmt.Errorf("serving directory: %w", err)
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
	return logger, nil
}
