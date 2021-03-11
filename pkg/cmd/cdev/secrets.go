package cdev

import (
	"github.com/apex/log"
	"github.com/shalb/cluster.dev/pkg/project"
	"github.com/shalb/cluster.dev/pkg/project/ui"
	"github.com/spf13/cobra"
)

// secretsCmd represents the plan command
var secretCmd = &cobra.Command{
	Use:   "secret",
	Short: "Manage secrets",
}

func init() {
	rootCmd.AddCommand(secretCmd)
	secretCmd.AddCommand(secretLs)
	secretCmd.AddCommand(secretEdit)
	secretCmd.AddCommand(secretCreate)
}

// secretsCmd represents the plan command
var secretLs = &cobra.Command{
	Use:   "ls",
	Short: "List secrets",
	Run: func(cmd *cobra.Command, args []string) {
		p, err := project.LoadProjectBase()
		if err != nil {
			log.Fatal(err.Error())
		}
		p.PrintSecretsList()
	},
}

var secretEdit = &cobra.Command{
	Use:   "edit [secret_name]",
	Short: "Create new secret or edit if exists",
	Run: func(cmd *cobra.Command, args []string) {
		p, err := project.LoadProjectBase()
		if err != nil {
			log.Fatal(err.Error())
		}
		if len(args) != 1 {
			log.Fatalf("Secret name is required")
		}
		err = p.Edit(args[0])
		if err != nil {
			log.Fatal(err.Error())
		}
	},
}

var secretCreate = &cobra.Command{
	Use:   "create",
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
