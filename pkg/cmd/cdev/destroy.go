package cdev

import (
	"github.com/shalb/cluster.dev/pkg/config"
	"github.com/shalb/cluster.dev/pkg/project"
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
		err = project.Destroy()
		if err != nil {
			project.UnLockState()
			return NewCmdErr(project, "destroy", err)
		}
		project.UnLockState()
		return NewCmdErr(project, "destroy", nil)
	},
}

func init() {
	rootCmd.AddCommand(destroyCmd)
	destroyCmd.Flags().BoolVar(&config.Global.IgnoreState, "ignore-state", false, "Destroy current configuration and ignore state.")
	destroyCmd.Flags().BoolVar(&config.Global.Force, "force", false, "Skip interactive approval.")
}
