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
		project, err := project.NewProject(config.Global.ProjectConf, config.Global.Manifests)
		if err != nil {
			log.Fatal(err.Error())
		}
		log.Info("Planing...")
		err = project.Build()
		if err != nil {
			log.Fatal(err.Error())
		}
		err = project.Plan()
		if err != nil {
			log.Fatal(err.Error())
		}
	},
}

func init() {
	rootCmd.AddCommand(planCmd)
}
