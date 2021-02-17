package cdev

import (
	"github.com/apex/log"
	"github.com/shalb/cluster.dev/pkg/config"
	"github.com/shalb/cluster.dev/pkg/project"
	"github.com/shalb/cluster.dev/pkg/project/ui"
	"github.com/spf13/cobra"
)

// newCmd represents cdev new command
var newCmd = &cobra.Command{
	Use:   "new",
	Short: "Code generator. Creates new object like 'project' or 'secret' from template in interactive mode",
}

func init() {
	rootCmd.AddCommand(newCmd)
	newCmd.AddCommand(newProject)
	newCmd.AddCommand(newSecret)
}

// projectsCmd represents the plan command
var newProject = &cobra.Command{
	Use:   "project",
	Short: "Generate new project from template in curent dir. Directory must be empty",
	Run: func(cmd *cobra.Command, args []string) {
		if project.ProjectsFilesExists() {
			log.Fatalf("project creating: some project's data (yaml files) found in current directory, use command in empty dir")
		}
		err := ui.CreteProject(config.Global.WorkingDir)
		if err != nil {
			log.Fatalf("Create project: %v", err.Error())
		}
	},
}

var newSecret = &cobra.Command{
	Use:   "secret",
	Short: "Generate new secret from template in curent project. Directory must contain the project",
	Run: func(cmd *cobra.Command, args []string) {
		_, err := project.LoadProjectBase()
		if err != nil {
			log.Fatalf("secret creating: check project: %v", err)
		}
		err = ui.CreateSecret()
		if err != nil {
			log.Fatalf("Create project: %v", err.Error())
		}
	},
}
