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
		log.Info("Destroying...")
		err = project.Destroy()
		if err != nil {
			log.Fatalf("Fatal error: destroy: %v", err.Error())
		}
	},
}

func init() {
	rootCmd.AddCommand(destroyCmd)
	destroyCmd.Flags().BoolVar(&config.Global.Force, "force", false, "Destroy curent configuration and ignore state.")
}