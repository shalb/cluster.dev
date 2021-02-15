package cdev

import (
	"github.com/apex/log"
	"github.com/shalb/cluster.dev/pkg/project"
	"github.com/shalb/cluster.dev/pkg/project/ui"
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
	projectCmd.AddCommand(projectCreate)
}

// projectsCmd represents the plan command
var projectLs = &cobra.Command{
	Use:   "info",
	Short: "List projects",
	Run: func(cmd *cobra.Command, args []string) {
		//p, err := project.LoadProjectBase()
		// if err != nil {
		// 	log.Info("No project found in the current directory.")
		// 	return
		// }
		ui.Create()
	},
}

var projectCreate = &cobra.Command{
	Use:   "create",
	Short: "Create new project",
	Run: func(cmd *cobra.Command, args []string) {
		p, err := project.LoadProjectBase()
		if err != nil {
			log.Fatal(err.Error())
		}
		if len(args) != 1 {
			log.Fatalf("project name is required")
		}
		err = p.Edit(args[0])
		if err != nil {
			log.Fatal(err.Error())
		}
	},
}
