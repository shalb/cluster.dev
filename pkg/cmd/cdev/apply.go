package cdev

import (
	"github.com/apex/log"
	"github.com/shalb/cluster.dev/pkg/config"
	"github.com/shalb/cluster.dev/pkg/project"
	"github.com/spf13/cobra"
)

// planCmd represents the plan command
var applyCmd = &cobra.Command{
	Use:   "apply",
	Short: "Deploys or updates infrastructure according to project configuration",
	Run: func(cmd *cobra.Command, args []string) {
		project, err := project.LoadProjectFull()
		if err != nil {
			log.Fatalf("Fatal error: apply: %v", err.Error())
		}
		err = project.LockState()
		if err != nil {
			log.Fatalf("Fatal error: apply: lock state: %v", err.Error())
		}
		err = project.Apply()
		if err != nil {
			project.UnLockState()
			log.Fatalf("Fatal error: apply: %v", err.Error())
		}
		err = project.PrintOutputs()
		if err != nil {
			log.Fatalf("Fatal error: apply: print outputs %v", err.Error())
		}
		project.UnLockState()
	},
}

func init() {
	rootCmd.AddCommand(applyCmd)
	applyCmd.Flags().BoolVar(&config.Global.IgnoreState, "ignore-state", false, "Apply even if the state has not changed.")
	applyCmd.Flags().BoolVar(&config.Global.Force, "force", false, "Skip interactive approval.")
}
