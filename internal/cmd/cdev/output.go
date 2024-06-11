package cdev

import (
	"github.com/apex/log"
	"github.com/shalb/cluster.dev/internal/config"
	"github.com/shalb/cluster.dev/internal/project"
	"github.com/spf13/cobra"
)

// planCmd represents the plan command
var outputCmd = &cobra.Command{
	Use:   "output",
	Short: "Show project outputs",
	Run: func(cmd *cobra.Command, args []string) {
		project, err := project.LoadProjectFull()
		if err != nil {
			log.Fatalf("Fatal error: outputs: %v", err.Error())
		}
		err = project.LockState()
		defer project.UnLockState()
		if err != nil {
			log.Fatalf("Fatal error: outputs: lock state: %v", err.Error())
		}
		err = project.OwnState.PrintOutputs()
		if err != nil {
			log.Fatalf("Fatal error: outputs: print %v", err.Error())
		}
	},
}

func init() {
	outputCmd.Flags().BoolVar(&config.Global.OutputJSON, "json", false, "Show outputs in JSON format.")
	rootCmd.AddCommand(outputCmd)
}
