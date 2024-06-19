package cdev

import (
	"github.com/apex/log"
	"github.com/shalb/cluster.dev/internal/config"
	"github.com/shalb/cluster.dev/internal/project"
	"github.com/spf13/cobra"
)

var destroyCmd = &cobra.Command{
	Use:           "destroy",
	SilenceUsage:  true,
	SilenceErrors: true,
	Short:         "Destroys infrastructure deployed by current project",
	RunE: func(cmd *cobra.Command, args []string) error {
		project, err := project.LoadProjectFull()
		if err != nil {
			return NewCmdErr(project, "destroy", err)
		}
		err = project.LockState()
		if err != nil {
			return NewCmdErr(project, "destroy", err)
		}
		defer project.UnLockState()
		err = project.Destroy()
		if err != nil {
			return NewCmdErr(project, "destroy", err)
		}
		log.Info("The project was successfully destroyed")
		return NewCmdErr(project, "destroy", nil)
	},
}

func init() {
	rootCmd.AddCommand(destroyCmd)
	destroyCmd.Flags().BoolVar(&config.Global.IgnoreState, "ignore-state", false, "Destroy current configuration and ignore state.")
	destroyCmd.Flags().BoolVar(&config.Global.Force, "force", false, "Skip interactive approval.")
	destroyCmd.Flags().StringArrayVarP(&config.Global.Targets, "target", "t", []string{}, "Units and stack that will be destroyed. All others will not destroy.")
}
