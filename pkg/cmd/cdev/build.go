package cdev

import (
	"github.com/apex/log"
	"github.com/shalb/cluster.dev/pkg/project"
	"github.com/spf13/cobra"
)

var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Build all units",
	Run: func(cmd *cobra.Command, args []string) {
		project, err := project.LoadProjectFull()
		if err != nil {
			log.Fatalf("Fatal error: build: %v", err.Error())
		}
		err = project.Build()
		log.Info("Building...")
		if err != nil {
			log.Fatalf("Fatal error: build: %v", err.Error())
		}
		log.Infof("The project was built successfully. Build directory: %v", project.CodeCacheDir)
	},
}

func init() {
	rootCmd.AddCommand(buildCmd)
}
