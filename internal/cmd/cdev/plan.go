package cdev

import (
	"fmt"

	"github.com/apex/log"
	"github.com/shalb/cluster.dev/internal/config"
	"github.com/shalb/cluster.dev/internal/project"
	"github.com/spf13/cobra"
)

// planCmd represents the plan command
var planCmd = &cobra.Command{
	Use:           "plan",
	Short:         "Show changes than will be applied in current project",
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		project, err := project.LoadProjectFull()
		if err != nil {
			return NewCmdErr(nil, "plan", fmt.Errorf("load project configuration: %w", err))
		}
		log.Info("Planning...")
		_, err = project.Plan()
		if err != nil {
			return NewCmdErr(project, "plan", fmt.Errorf("build plan: %w", err))
		}
		return NewCmdErr(project, "plan", nil)
	},
}

func init() {
	rootCmd.AddCommand(planCmd)
	// planCmd.Flags().BoolVar(&config.Global.ShowTerraformPlan, "tf-plan", false, "Also show units terraform plan if possible.")
	planCmd.Flags().BoolVar(&config.Global.IgnoreState, "force", false, "Show plan (if set tf-plan) even if the state has not changed.")
}
