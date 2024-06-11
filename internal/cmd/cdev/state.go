package cdev

import (
	"github.com/apex/log"
	"github.com/shalb/cluster.dev/internal/config"
	"github.com/shalb/cluster.dev/internal/project"
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
		config.Global.IgnoreState = true
		project, err := project.LoadProjectFull()
		if err != nil {
			log.Fatalf("Fatal error: state unlock: %v", err.Error())
		}
		log.Info("Unlocking state...")
		err = project.UnLockState()
		if err != nil {
			log.Fatalf("Fatal error: state unlock: %v", err.Error())
		}
	},
}

// planCmd represents the plan command
var stateUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Updates the state of the current project to version %v. Make sure that the state of the project is consistent (run cdev apply with the old version before)",
	Run: func(cmd *cobra.Command, args []string) {
		config.Global.IgnoreState = true
		project, err := project.LoadProjectFull()
		if err != nil {
			log.Fatalf("Fatal error: state update: %v", err.Error())
		}
		project.LockState()
		defer project.UnLockState()

		err = project.BackupState()
		if err != nil {

			log.Fatalf("Fatal error: state update: %v", err.Error())
		}
		log.Info("Updating state...")
		err = project.SaveState()
		if err != nil {
			log.Fatalf("Fatal error: state update: %v", err.Error())
		}
	},
}

// planCmd represents the plan command
var statePullCmd = &cobra.Command{
	Use:   "pull",
	Short: "Downloads the remote state",
	Run: func(cmd *cobra.Command, args []string) {
		config.Global.IgnoreState = true
		project, err := project.LoadProjectFull()
		if err != nil {
			log.Fatalf("Fatal error: state pull: %v", err.Error())
		}
		project.LockState()
		defer project.UnLockState()

		log.Info("Updating state...")
		err = project.PullState()
		if err != nil {
			log.Fatalf("Fatal error: state pull: %v", err.Error())
		}
	},
}

func init() {
	rootCmd.AddCommand(stateCmd)
	stateCmd.AddCommand(stateUnlockCmd)
	stateCmd.AddCommand(stateUpdateCmd)
	stateCmd.AddCommand(statePullCmd)
}
