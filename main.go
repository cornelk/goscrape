package main

import (
	"fmt"
	"strings"

	"github.com/cornelk/goscrape/scraper"
	"github.com/cornelk/gotokit/env"
	"github.com/cornelk/gotokit/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "goscrape http://website.com",
		Short: "Scrape a website and create an offline browsable version on the disk",
		RunE:  startScraper,
	}

	rootCmd.Flags().String("config", "", "config file (default is $HOME/.goscrape.yaml)")
	rootCmd.Flags().StringArrayP("include", "n", nil, "only include URLs with PERL Regular Expressions support")
	rootCmd.Flags().StringArrayP("exclude", "x", nil, "exclude URLs with PERL Regular Expressions support")
	rootCmd.Flags().StringP("output", "o", "", "output directory to write files to")
	rootCmd.Flags().IntP("imagequality", "i", 0, "image quality, 0 to disable reencoding")
	rootCmd.Flags().UintP("depth", "d", 10, "download depth, 0 for unlimited")
	rootCmd.Flags().UintP("timeout", "t", 0, "time limit in seconds for each http request to connect and read the request body")
	rootCmd.Flags().BoolP("verbose", "v", false, "verbose output")
	rootCmd.Flags().StringP("user", "u", "", "user[:password] to use for authentication")
	rootCmd.Flags().StringP("proxy", "p", "", "HTTP proxy during scraping")

	if err := rootCmd.Execute(); err != nil {
		fmt.Printf("ERROR: %v\n", err)
	}
}

func startScraper(cmd *cobra.Command, args []string) error {
	initializeViper(cmd)

	if len(args) == 0 {
		_ = cmd.Help()
		return nil
	}

	var username, password string
	userParam, _ := cmd.Flags().GetString("user")
	if userParam != "" {
		sl := strings.Split(userParam, ":")
		username = sl[0]
		if len(sl) > 1 {
			password = sl[1]
		}
	}

	includes, _ := cmd.Flags().GetStringArray("include")
	excludes, _ := cmd.Flags().GetStringArray("exclude")
	imageQuality, _ := cmd.Flags().GetInt("imagequality")
	if imageQuality < 0 || imageQuality >= 100 {
		imageQuality = 0
	}
	output, _ := cmd.Flags().GetString("output")
	depth, _ := cmd.Flags().GetUint("depth")
	timeout, _ := cmd.Flags().GetUint("timeout")
	proxy, _ := cmd.Flags().GetString("proxy")

	logger, err := createLogger()
	if err != nil {
		return fmt.Errorf("creating logger: %w", err)
	}

	cfg := scraper.Config{
		Includes:        includes,
		Excludes:        excludes,
		ImageQuality:    uint(imageQuality),
		MaxDepth:        depth,
		Timeout:         timeout,
		OutputDirectory: output,
		Username:        username,
		Password:        password,
		Proxy:           proxy,
	}

	for _, url := range args {
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
	logger = logger.Named("goscrape")
	return logger, nil
}

func initializeViper(cmd *cobra.Command) {
	configFile, err := cmd.Flags().GetString("config")
	if err == nil && configFile != "" { // enable ability to specify config file via flag
		viper.SetConfigFile(configFile)
	}

	viper.SetConfigName(".goscrape") // name of config file (without extension)
	viper.AddConfigPath("$HOME")     // adding home directory as first search path
	viper.AutomaticEnv()             // read in environment variables that match

	_ = viper.ReadInConfig()
}
