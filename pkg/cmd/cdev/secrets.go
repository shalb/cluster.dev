package cdev

import (
	"fmt"

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
	secretCmd.AddCommand(secretCreate)
	for secTp, _ := range project.SecretDriversMap {
		secretCreate.AddCommand(getCreateSubcommand(secTp))
	}
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
		err = p.Edit(args[0])
		if err != nil {
			log.Fatal(err.Error())
		}
	},
}

var secretCreate = &cobra.Command{
	Use:   "create",
	Short: "Create new secret",
}

func getCreateSubcommand(secretType string) (res *cobra.Command) {
	res = &cobra.Command{
		Use:   fmt.Sprintf("%v [secret_name]", secretType),
		Short: fmt.Sprintf("Create new secret type of %v", secretType),
		Run: func(cmd *cobra.Command, args []string) {
			p, err := project.NewEmptyProject(config.Global.ProjectConf, config.Global.Manifests)
			if err != nil {
				log.Fatal(err.Error())
			}
			if len(args) != 1 {
				log.Fatalf("Secret name is required")
			}
			err = p.Create(secretType, args[0])
			if err != nil {
				log.Fatal(err.Error())
			}
		},
	}
	return
}
