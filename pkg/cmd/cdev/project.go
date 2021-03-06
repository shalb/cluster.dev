package cdev

import (
	"fmt"

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
var listAllTemplates bool

func init() {
	rootCmd.AddCommand(projectCmd)
	projectCmd.AddCommand(projectLs)
	projectCmd.AddCommand(projectCreate)
	projectCreate.Flags().BoolVar(&config.Global.Interactive, "interactive", false, "Use interactive mode for project generation")
	projectCreate.Flags().BoolVar(&listAllTemplates, "list-templates", false, "Show all available templates for project generation")
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
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return fmt.Errorf("requires a template git URL argument")
		}
		if len(args) > 2 {
			return fmt.Errorf("too many arguments")
		}
		return nil
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
		if len(args) < 1 {
			log.Fatal("project creating: ")
		}
		err := ui.CreteProject(config.Global.WorkingDir, args[0], args[1:]...)
		if err != nil {
			log.Fatalf("Create project: %v", err.Error())
		}
	},
}
