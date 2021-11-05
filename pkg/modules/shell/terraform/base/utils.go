package base

import (
	"fmt"

	"github.com/shalb/cluster.dev/pkg/project"
)

func DependencyToRemoteStateRef(dep *project.DependencyOutput) (remoteStateRef string) {
	remoteStateRef = fmt.Sprintf("data.terraform_remote_state.%s-%s.outputs.%s", dep.StackName, dep.UnitName, dep.Output)
	return
}
func DependencyToBashRemoteState(dep *project.DependencyOutput) (remoteStateRef string) {
	remoteStateRef = fmt.Sprintf("\"$(terraform -chdir=../%v.%v/ output -raw %v)\"", dep.StackName, dep.UnitName, dep.Output)
	return
}
