package cdev

import (
	"github.com/apex/log"
	"github.com/shalb/cluster.dev/pkg/config"
	"github.com/shalb/cluster.dev/pkg/project"
	"github.com/spf13/cobra"
)

var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Build all modules",
	Run: func(cmd *cobra.Command, args []string) {
		project, err := project.NewProject(config.Global.ProjectConf, config.Global.Manifests)
		if err != nil {
			log.Fatal(err.Error())
		}
		err = project.Build()
		log.Info("Building...")
		if err != nil {
			log.Fatal(err.Error())
		}
	},
}

func init() {
	rootCmd.AddCommand(buildCmd)
}
