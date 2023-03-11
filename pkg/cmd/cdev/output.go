package cdev

import (
	"github.com/apex/log"
	"github.com/shalb/cluster.dev/pkg/config"
	"github.com/shalb/cluster.dev/pkg/project"
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
		if err != nil {
			log.Fatalf("Fatal error: outputs: lock state: %v", err.Error())
		}
		stProject, err := project.LoadState()
		if err != nil {
			project.UnLockState()
			log.Fatalf("Fatal error: outputs: load state: %v", err.Error())
		}
		err = stProject.PrintOutputs()
		if err != nil {
			log.Fatalf("Fatal error: outputs: print %v", err.Error())
		}
		project.UnLockState()
	},
}

func init() {
  outputCmd.Flags().BoolVar(&config.Global.OutputJSON, "json", false, "Show outputs in JSON format.")
	rootCmd.AddCommand(outputCmd)
}
