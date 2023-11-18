package cdev

import (
	"fmt"

	"github.com/apex/log"
	"github.com/shalb/cluster.dev/pkg/config"
	"github.com/shalb/cluster.dev/pkg/profiler"
	"github.com/shalb/cluster.dev/pkg/utils"
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
	profiler.Global.MainTimeLine().Start()
	err := rootCmd.Execute()
	extendedErr, ok := err.(*CmdErrExtended)
	if !ok {
		log.Debugf("Usage stats are unavailable in current command. Ignore.")
		if err != nil {
			log.Fatalf("Fatal error: %v", err.Error())
		}
		return
	}
	profiler.Global.MainTimeLine().SetPoint(rootCmd.Name())
	p := extendedErr.ProjectPtr
	//statsExporter := utils.StatsExporter{}
	st := extendedErr.CdevUsage
	st.AbsoluteTime = profiler.Global.MainTimeLine().Duration().String()
	st.RealTime = profiler.Global.MainTimeLine().Duration().String()
	st.Operation = extendedErr.Command
	if p != nil {
		st.ProcessedUnitsCount = p.ProcessedUnitsCount
	}
	if extendedErr.Err == nil {
		st.OperationResult = "ok"
	} else {
		st.OperationResult = "fail"
	}
	if p != nil {
		st.ProjectID = p.UUID
		if len(p.Backends) == 0 || p.Backends[p.StateBackendName] == nil {
			st.BackendType = "null"
		} else {
			st.BackendType = p.Backends[p.StateBackendName].Provider()
		}
	} else {
		st.ProjectID = "null"
		st.BackendType = "null"
	}
	exporter := utils.StatsExporter{}
	_ = exporter.PushStats(st)
	if extendedErr.Err != nil {
		log.Fatalf("Fatal error: %v", err.Error())
	}
}
