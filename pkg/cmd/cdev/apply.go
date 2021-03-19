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
	Short: "Apply all modules",
	Run: func(cmd *cobra.Command, args []string) {
		project, err := project.LoadProjectFull()
		if err != nil {
			log.Fatal(err.Error())
		}
		log.Info("Applying...")
		err = project.Apply()
		if err != nil {
			log.Fatal(err.Error())
		}
	},
}

func init() {
	rootCmd.AddCommand(applyCmd)
	applyCmd.Flags().BoolVar(&config.Global.IgnoreState, "ignore-state", false, "Apply even if the state has not changed.")
}
