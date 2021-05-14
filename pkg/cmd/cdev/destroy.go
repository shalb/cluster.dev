package cdev

import (
	"github.com/apex/log"
	"github.com/shalb/cluster.dev/pkg/config"
	"github.com/shalb/cluster.dev/pkg/project"
	"github.com/spf13/cobra"
)

var destroyCmd = &cobra.Command{
	Use:   "destroy",
	Short: "Destroy all modules",
	Run: func(cmd *cobra.Command, args []string) {
		project, err := project.LoadProjectFull()
		if err != nil {
			log.Fatalf("Fatal error: destroy: %v", err.Error())
		}
		err = project.LockState()
		if err != nil {
			log.Fatalf("Fatal error: destroy: lock state: %v", err.Error())
		}
		err = project.Destroy()
		if err != nil {
			project.UnLockState()
			log.Fatalf("Fatal error: destroy: %v", err.Error())
		}
		project.UnLockState()
	},
}

func init() {
	rootCmd.AddCommand(destroyCmd)
	destroyCmd.Flags().BoolVar(&config.Global.IgnoreState, "ignore-state", false, "Destroy current configuration and ignore state.")
	destroyCmd.Flags().BoolVar(&config.Global.Force, "force", false, "Skip interactive approval.")
}
