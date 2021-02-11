package cdev

import (
	"github.com/apex/log"
	"github.com/shalb/cluster.dev/pkg/config"
	"github.com/shalb/cluster.dev/pkg/project"
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
}

// secretsCmd represents the plan command
var secretLs = &cobra.Command{
	Use:   "ls",
	Short: "List secrets",
	Run: func(cmd *cobra.Command, args []string) {
		p, err := project.NewEmptyProject(config.Global.ProjectConf, config.Global.Manifests)
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
		p, err := project.NewEmptyProject(config.Global.ProjectConf, config.Global.Manifests)
		if err != nil {
			log.Fatal(err.Error())
		}
		if len(args) != 1 {
			log.Fatalf("Secret name is required")
		}
		p.EditSecret(args[0])
	},
}
