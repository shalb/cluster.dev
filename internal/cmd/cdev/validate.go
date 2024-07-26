package cdev

import (
	"github.com/apex/log"
	"github.com/gookit/color"
	"github.com/shalb/cluster.dev/internal/config"
	"github.com/shalb/cluster.dev/internal/project"
	"github.com/spf13/cobra"
)

// projectsCmd represents the plan command
var projectValidate = &cobra.Command{
	Use:   "validate",
	Short: "Validates the configuration files in a project directory, referring only to the configuration and not accessing remote state bucket",
	Run: func(cmd *cobra.Command, args []string) {
		config.Global.IgnoreState = true
		_, err := project.LoadProjectFull()
		if err != nil {
			log.Fatalf("Project configuration check: %v\n%v", color.Style{color.FgGreen, color.OpBold}.Sprintf("fail"), err.Error())
		}
		log.Infof("Project configuration check: %v", color.Style{color.FgGreen, color.OpBold}.Sprintf("valid"))
	},
}

func init() {
	rootCmd.AddCommand(projectValidate)
}
