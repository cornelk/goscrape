package cmd

import (
	"os"

	"github.com/cornelk/goscrape/appcontext"
	"github.com/cornelk/goscrape/scraper"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/uber-go/zap"
)

var cfgFile string

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "goscrape",
	Short: "Scrape a website and create an offline browseable version on the disk",
	Run: func(cmd *cobra.Command, args []string) {
		for _, url := range args {
			log := appcontext.Logger
			sc, err := scraper.New(url)
			if err != nil {
				log.Fatal("Error occured", zap.Error(err))
			}

			log.Info("Scraping", zap.Stringer("URL", sc.URL))
			err = sc.Start()
			if err != nil {
				log.Error("Error occured", zap.Error(err))
			}
		}
	},
}

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		os.Exit(-1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.goscrape.yaml)")

	RootCmd.Flags().UintP("depth", "d", 10, "Download depth, 0 for unlimited")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" { // enable ability to specify config file via flag
		viper.SetConfigFile(cfgFile)
	}

	viper.SetConfigName(".goscrape") // name of config file (without extension)
	viper.AddConfigPath("$HOME")     // adding home directory as first search path
	viper.AutomaticEnv()             // read in environment variables that match

	_ = viper.ReadInConfig()
}
