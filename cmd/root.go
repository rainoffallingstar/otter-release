package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/xdxtools/xdxtools-go/internal/logger"
)

var (
	cfgFile   string
	userLevel bool
	verbose   bool

	// buildVersion is the version of the binary
	buildVersion = "0.1.0"
	// buildCommit is the git commit hash
	buildCommit = "unknown"
	// buildDate is the build date
	buildDate = "unknown"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "xdxtools",
	Short: "Bioinformatics workflow management tool",
	Long: `xdxtools is a bioinformatics workflow management tool for RRBS, WGBS, RNA-seq, and PDX analysis.
It integrates with Snakemake and supports Slurm and local execution environments.`,
	Version: fmt.Sprintf("%s+%s (%s)", buildVersion, buildCommit, buildDate),
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)

	// Global flags
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file (default is xdxtools.yaml)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
	rootCmd.PersistentFlags().BoolVar(&userLevel, "user-level", false, "run at user level (for PDX)")

	// Bind flags with viper
	// viper.BindPFlag("author", rootCmd.PersistentFlags().Lookup("author"))
	// viper.BindPFlag("projectbase", rootCmd.PersistentFlags().Lookup("projectbase"))
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		// viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		// home, err := os.UserHomeDir()
		// cobra.CheckErr(err)

		// Search config in home directory with name "xdxtools" (without extension).
		// viper.AddConfigPath(home)
		// viper.SetConfigType("yaml")
		// viper.SetConfigName("xdxtools")
	}

	// viper.AutomaticEnv()

	// If a config file is found, read it in.
	// if err := viper.ReadInConfig(); err == nil {
	// 	fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	// }

	// Initialize logger
	logger.Init(verbose)
}
