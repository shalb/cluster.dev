package base

import (
	"fmt"

	"github.com/shalb/cluster.dev/pkg/project"
)

func DependencyToRemoteStateRef(dep *project.ULinkT) (remoteStateRef string) {
	remoteStateRef = fmt.Sprintf("data.terraform_remote_state.%s-%s.outputs.%s", dep.TargetStackName, dep.TargetUnitName, dep.OutputName)
	return
}
func DependencyToBashRemoteState(dep *project.ULinkT) (remoteStateRef string) {
	remoteStateRef = fmt.Sprintf("\"$(terraform -chdir=../%v.%v/ output -raw %v)\"", dep.TargetStackName, dep.TargetUnitName, dep.OutputName)
	return
}
