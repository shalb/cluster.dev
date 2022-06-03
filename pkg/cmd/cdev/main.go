package cdev

import (
	"fmt"

	"github.com/apex/log"
	"github.com/shalb/cluster.dev/pkg/config"
	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "cdev",
	Short: "See https://docs.cluster.dev/ for details.",
}

func init() {
	cobra.OnInitialize(config.InitConfig)
	rootCmd.Version = fmt.Sprintf("%v\nbuild timestamp: %v", config.Version, config.BuildTimestamp)
	rootCmd.PersistentFlags().StringVarP(&config.Global.LogLevel, "log-level", "l", "info", "Set the logging level ('debug'|'info'|'warn'|'error'|'fatal')")
	rootCmd.PersistentFlags().BoolVar(&config.Global.UseCache, "cache", false, "Use previously cached build directory")
	rootCmd.PersistentFlags().IntVar(&config.Global.MaxParallel, "parallelism", 3, "Max parallel threads for units applying")
	rootCmd.PersistentFlags().BoolVar(&config.Global.TraceLog, "trace", false, "Print functions trace info in logs")
	rootCmd.PersistentFlags().BoolVar(&config.Global.NoColor, "no-color", false, "Turn off colored output")
	rootCmd.PersistentFlags().BoolP("version", "v", false, "Print client version")
	rootCmd.PersistentFlags().BoolP("help", "h", false, "Show this help output")
	_ = rootCmd.PersistentFlags().MarkHidden("trace")
}

func Run() {

	err := rootCmd.Execute()
	if err != nil {
		log.Fatalf("Fatal error: %v", err.Error())
	}
}
