package main

import (
	"fmt"
	"strings"

	"github.com/cornelk/goscrape/scraper"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "goscrape http://website.com",
		Short: "Scrape a website and create an offline browsable version on the disk",
		Run:   startScraper,
	}

	rootCmd.Flags().String("config", "", "config file (default is $HOME/.goscrape.yaml)")
	rootCmd.Flags().StringArrayP("include", "n", nil, "only include URLs with PERL Regular Expressions support")
	rootCmd.Flags().StringArrayP("exclude", "x", nil, "exclude URLs with PERL Regular Expressions support")
	rootCmd.Flags().StringP("output", "o", "", "output directory to write files to")
	rootCmd.Flags().IntP("imagequality", "i", 0, "image quality, 0 to disable reencoding")
	rootCmd.Flags().UintP("depth", "d", 10, "download depth, 0 for unlimited")
	rootCmd.Flags().BoolP("verbose", "v", false, "verbose output")
	rootCmd.Flags().StringP("user", "u", "", "user[:password] to use for authentication")

	if err := rootCmd.Execute(); err != nil {
		fmt.Printf("ERROR: %v\n", err)
	}
}

func startScraper(cmd *cobra.Command, args []string) {
	configFile, err := cmd.Flags().GetString("config")
	if err == nil && configFile != "" { // enable ability to specify config file via flag
		viper.SetConfigFile(configFile)
	}

	viper.SetConfigName(".goscrape") // name of config file (without extension)
	viper.AddConfigPath("$HOME")     // adding home directory as first search path
	viper.AutomaticEnv()             // read in environment variables that match

	_ = viper.ReadInConfig()

	if len(args) == 0 {
		_ = cmd.Help()
		return
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
	excludes, _ := cmd.Flags().GetStringArray("excludes")
	imageQuality, _ := cmd.Flags().GetInt("imagequality")
	if imageQuality < 0 || imageQuality >= 100 {
		imageQuality = 0
	}
	output, _ := cmd.Flags().GetString("output")
	depth, _ := cmd.Flags().GetUint("depth")

	logger := logger(cmd)
	for _, url := range args {
		sc, err := scraper.New(logger, url)
		if err != nil {
			logger.Fatal("Initializing scraper failed", zap.Error(err))
		}

		err = sc.SetIncludes(includes)
		if err != nil {
			logger.Fatal("Configuring includes failed", zap.Error(err))
		}

		err = sc.SetExcludes(excludes)
		if err != nil {
			logger.Fatal("Configuring excludes failed", zap.Error(err))
		}

		sc.ImageQuality = uint(imageQuality)
		sc.MaxDepth = depth
		sc.OutputDirectory = output
		sc.Username = username
		sc.Password = password

		logger.Info("Scraping", zap.Stringer("URL", sc.URL))
		err = sc.Start()
		if err != nil {
			logger.Error("Scraping failed", zap.Error(err))
		}
	}
}

func logger(cmd *cobra.Command) *zap.Logger {
	config := zap.NewDevelopmentConfig()
	config.Development = false
	config.DisableCaller = true
	config.DisableStacktrace = true

	level := config.Level
	verbose, _ := cmd.Flags().GetBool("verbose")
	if verbose {
		level.SetLevel(zap.DebugLevel)
	} else {
		level.SetLevel(zap.InfoLevel)
	}

	log, _ := config.Build()
	return log
}
