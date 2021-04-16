package cdev

import (
	"github.com/apex/log"
	"github.com/shalb/cluster.dev/pkg/config"
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
	projectCreate.Flags().BoolVar(&config.Global.Interactive, "interactive", false, "Use intteractive mode to for project generation")
}

// projectsCmd represents the plan command
var projectLs = &cobra.Command{
	Use:   "info",
	Short: "List projects",
	Run: func(cmd *cobra.Command, args []string) {
		p, err := project.LoadProjectFull()
		if err != nil {
			log.Errorf("No project found in the current directory. Configuration error: %v", err.Error())
			return
		}
		log.Info("Project info:")
		p.PrintInfo()
	},
}

// projectsCmd represents the plan command
var projectCreate = &cobra.Command{
	Use:   "create",
	Short: "Generate new project from template in curent dir. Directory must be empty",
	Run: func(cmd *cobra.Command, args []string) {
		if project.ProjectsFilesExists() {
			log.Fatalf("project creating: some project's data (yaml files) found in current directory, use command in empty dir")
		}
		err := ui.CreteProject(config.Global.WorkingDir, args...)
		if err != nil {
			log.Fatalf("Create project: %v", err.Error())
		}
	},
}
