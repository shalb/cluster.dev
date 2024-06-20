package cdev

import (
	"github.com/apex/log"
	"github.com/shalb/cluster.dev/internal/config"
	"github.com/shalb/cluster.dev/internal/project"
	"github.com/shalb/cluster.dev/pkg/utils"
	"github.com/spf13/cobra"
)

// planCmd represents the plan command
var applyCmd = &cobra.Command{
	Use:           "apply",
	SilenceUsage:  true,
	SilenceErrors: true,
	Short:         "Deploys or updates infrastructure according to project configuration",
	RunE: func(cmd *cobra.Command, args []string) error {
		project, err := project.LoadProjectFull()
		if utils.GetEnv("CDEV_COLLECT_USAGE_STATS", "true") != "false" {
			log.Infof("Sending usage statistic. To disable statistics collection, export the CDEV_COLLECT_USAGE_STATS=false environment variable")
		}
		if err != nil {
			return NewCmdErr(project, "apply", err)
		}
		err = project.LockState()
		if err != nil {
			return NewCmdErr(project, "apply", err)
		}
		defer project.UnLockState()
		err = project.Apply()
		if err != nil {
			return NewCmdErr(project, "apply", err)
		}
		log.Info("The project was successfully applied")
		err = project.PrintOutputs()
		if err != nil {
			return NewCmdErr(project, "apply", err)
		}
		return NewCmdErr(project, "apply", nil)
	},
}

func init() {
	rootCmd.AddCommand(applyCmd)
	applyCmd.Flags().BoolVar(&config.Global.IgnoreState, "ignore-state", false, "Apply even if the state has not changed.")
	applyCmd.Flags().BoolVar(&config.Global.Force, "force", false, "Skip interactive approval.")
	applyCmd.Flags().StringArrayVarP(&config.Global.Targets, "target", "t", []string{}, "Units and stack that will be applied. All others will not apply.")
}
