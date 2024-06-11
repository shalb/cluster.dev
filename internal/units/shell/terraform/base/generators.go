package base

import (
	"fmt"
	"io/fs"
	"strings"

	"github.com/apex/log"
	"github.com/shalb/cluster.dev/internal/project"
	"github.com/shalb/cluster.dev/pkg/hcltools"
)

// genBackendCodeBlock generate backend code block for this unit.
func (u *Unit) genBackendCodeBlock() ([]byte, error) {

	f, err := (*u.BackendPtr).GetBackendHCL(u.StackName(), u.Name())
	if err != nil {
		log.Debug(err.Error())
		return nil, err
	}
	if len(u.RequiredProviders) < 1 {
		return f.Bytes(), nil
	}
	tb := f.Body().Blocks()[0]
	tfBlock := tb.Body().AppendNewBlock("required_providers", []string{})
	for name, prov := range u.RequiredProviders {

		reqProvs, err := hcltools.InterfaceToCty(prov)
		if err != nil {
			return nil, err
		}
		tfBlock.Body().SetAttributeValue(name, reqProvs)
	}
	return f.Bytes(), nil
}

// genDepsRemoteStates generate terraform remote states for all dependencies of this unit.
func (u *Unit) genDepsRemoteStates() ([]byte, error) {
	var res []byte
	DeDuplication := map[project.Unit]bool{}
	for _, dep := range u.Dependencies().ByLinkTypes(RemoteStateLinkType).Slice() {
		// Ignore duplicated dependencies.
		if _, ok := DeDuplication[dep.Unit]; ok {
			continue
		}
		// Ignore dependencies without output (user defined as 'depends_on' option.)
		if dep.OutputName == "" {
			continue
		}
		// De-duplication.
		DeDuplication[dep.Unit] = true
		modBackend := dep.Unit.Stack().Backend
		if dep.Unit == nil || dep.Unit.Stack() == nil {
			continue
		}
		// log.Warnf("%v", modBackend)
		rs, err := modBackend.GetRemoteStateHCL(dep.Unit.Stack().Name, dep.Unit.Name())
		if err != nil {
			log.Debug(err.Error())
			return nil, err
		}
		res = append(res, rs...)
	}
	return res, nil
}

func (u *Unit) Build() error {
	var err error
	init, err := u.genBackendCodeBlock()
	if err != nil {
		log.Debug(err.Error())
		return err
	}
	if u.Providers != nil {
		providers, err := hcltools.ProvidersToHCL(u.Providers)
		if err != nil {
			log.Debug(err.Error())
			return err
		}
		for hash, marker := range u.ProjectPtr.UnitLinks.ByLinkTypes(RemoteStateLinkType).Map() {
			if marker.TargetStackName == "this" {
				marker.TargetStackName = u.Stack().Name
			}
			refStr := DependencyToRemoteStateRef(marker)
			hcltools.ReplaceStingMarkerInBody(providers.Body(), hash, refStr)
		}
		init = append(init, providers.Bytes()...)
	}
	err = u.CreateFiles.AddOverride("init.tf", string(init), fs.ModePerm)
	if err != nil {
		return fmt.Errorf("build unit %v: %w\n%v", u.Key(), err, u.CreateFiles.SPrintLs())
	}
	// Create remote_state.tf
	remoteStates, err := u.genDepsRemoteStates()
	if err != nil {
		log.Debug(err.Error())
		return err
	}
	// log.Errorf("Remote states: %v\nUnit name: %v", len(remoteStates), m.Key())
	if len(remoteStates) > 0 {
		err = u.CreateFiles.AddOverride("remote_states.tf", string(remoteStates), fs.ModePerm)
		if err != nil {
			return fmt.Errorf("build unit %v: %w", u.Key(), err)
		}
	}
	if u.PreHook != nil {
		err := u.replaceRemoteStatesForBash(&u.PreHook.Command)
		if err != nil {
			return err
		}
	}
	if u.PostHook != nil {
		err := u.replaceRemoteStatesForBash(&u.PostHook.Command)
		if err != nil {
			return err
		}
	}

	return u.Unit.Build()
}
func (u *Unit) replaceRemoteStatesForBash(cmd *string) error {
	if cmd == nil {
		return nil
	}
	markersList := u.ProjectPtr.UnitLinks.ByLinkTypes(RemoteStateLinkType).Map()
	for hash, marker := range markersList {
		if marker.TargetStackName == "this" {
			return fmt.Errorf("internal error, debug: %+v", marker)
			//marker.TargenStackName = m.Stack().Name
		}
		refStr := DependencyToBashRemoteState(marker)
		c := strings.ReplaceAll(*cmd, hash, refStr)
		cmd = &c
	}
	return nil
}
