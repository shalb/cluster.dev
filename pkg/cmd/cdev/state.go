package cdev

import (
	"github.com/apex/log"
	"github.com/shalb/cluster.dev/pkg/project"
	"github.com/spf13/cobra"
)

// planCmd represents the plan command
var stateCmd = &cobra.Command{
	Use:   "state",
	Short: "State operations",
}

// planCmd represents the plan command
var stateUnlockCmd = &cobra.Command{
	Use:   "unlock",
	Short: "Unlock state force",
	Run: func(cmd *cobra.Command, args []string) {
		project, err := project.LoadProjectFull()
		if err != nil {
			log.Fatalf("Fatal error: plan: %v", err.Error())
		}
		log.Info("Unlocking state...")
		err = project.UnLockState()
		if err != nil {
			log.Fatalf("Fatal error: plan: %v", err.Error())
		}
	},
}

func init() {
	rootCmd.AddCommand(stateCmd)
	stateCmd.AddCommand(stateUnlockCmd)
}
