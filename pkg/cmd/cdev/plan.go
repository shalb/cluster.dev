package cdev

import (
	"github.com/apex/log"
	"github.com/shalb/cluster.dev/pkg/config"
	"github.com/shalb/cluster.dev/pkg/project"
	"github.com/spf13/cobra"
)

// planCmd represents the plan command
var planCmd = &cobra.Command{
	Use:   "plan",
	Short: "Plan all modules",
	Run: func(cmd *cobra.Command, args []string) {
		project, err := project.LoadProjectFull()
		if err != nil {
			log.Fatalf("Fatal error: plan: %v", err.Error())
		}
		log.Info("Planning...")
		err = project.Plan()
		if err != nil {
			log.Fatalf("Fatal error: plan: %v", err.Error())
		}
	},
}

func init() {
	rootCmd.AddCommand(planCmd)
	planCmd.Flags().BoolVar(&config.Global.ShowTerraformPlan, "tf-plan", false, "Also show modules terraform plan if possible.")
	planCmd.Flags().BoolVar(&config.Global.IgnoreState, "force", false, "Show plan (if set tf-plan) even if the state has not changed.")
}
