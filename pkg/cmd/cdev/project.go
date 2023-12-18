package cdev

import (
	"fmt"
	"strings"

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
	Short: "Shows detailed information about the current project, such as the number of units and their types. Number of stacks, etc",
	Run: func(cmd *cobra.Command, args []string) {
		p, err := project.LoadProjectFull()
		if err != nil {
			log.Errorf("Project configuration error: %v", err.Error())
			return
		}
		log.Info("Project info:")
		p.PrintInfo()
	},
}

// projectsCmd represents the plan command
var projectCreate = &cobra.Command{
	Use:   "create",
	Short: "Generate new project from generator-template in current directory. Directory should not contain yaml or yml files",
	Run: func(cmd *cobra.Command, args []string) {
		if listAllTemplates {
			list, err := ui.GetProjectTemplates(args[0])
			if err != nil {
				log.Fatalf("List project templates: %v", err.Error())
			}
			res := strings.Join(list, "\n")
			fmt.Println(res)
			return
		}
		if project.ProjectsFilesExists() {
			log.Fatalf("project creating: some project's data (yaml files) found in current directory, use command in empty dir")
		}
		if len(args) < 1 {
			log.Fatal("project creating: ")
		}
		err := ui.CreateProject(config.Global.WorkingDir, args[0], args[1:]...)
		if err != nil {
			log.Fatalf("Create project: %v", err.Error())
		}
	},
}
