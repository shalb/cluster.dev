package cdev

import (
	"github.com/apex/log"
	"github.com/shalb/cluster.dev/internal/project"
	"github.com/shalb/cluster.dev/internal/project/ui"
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
			log.Fatalf("Fatal error: secret ls: %v", err.Error())
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
			log.Fatalf("Fatal error: secret edit: %v", err.Error())
		}
		if len(args) != 1 {
			log.Fatal("Fatal error: secret edit: secret name is required")
		}
		err = p.Edit(args[0])
		if err != nil {
			log.Fatalf("Fatal error: secret edit: %v", err.Error())
		}
	},
}

var secretCreate = &cobra.Command{
	Use:   "create",
	Short: "Generate new secret in current directory. Directory must contain the project",
	Run: func(cmd *cobra.Command, args []string) {
		_, err := project.LoadProjectBase()
		if err != nil {
			log.Fatalf("Fatal error: secret create: ", err.Error())
		}
		err = ui.CreateSecret()
		if err != nil {
			log.Fatalf("Fatal error: secret create: ", err.Error())
		}
	},
}
