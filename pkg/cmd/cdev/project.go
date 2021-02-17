package cdev

import (
	"github.com/apex/log"
	"github.com/shalb/cluster.dev/pkg/project"
	"github.com/spf13/cobra"
)

// projectsCmd represents the plan command
var projectCmd = &cobra.Command{
	Use:   "project",
	Short: "Manage projects",
}

func init() {
	rootCmd.AddCommand(projectCmd)
	projectCmd.AddCommand(projectLs)
}

// projectsCmd represents the plan command
var projectLs = &cobra.Command{
	Use:   "info",
	Short: "List projects",
	Run: func(cmd *cobra.Command, args []string) {
		p, err := project.LoadProjectFull()
		if err != nil {
			log.Info("No project found in the current directory.")
			return
		}
		log.Info("Project info:")
		p.PrintInfo()
	},
}
