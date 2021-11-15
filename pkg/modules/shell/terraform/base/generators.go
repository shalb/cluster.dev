package base

import (
	"fmt"
	"io/fs"
	"strings"

	"github.com/apex/log"
	"github.com/shalb/cluster.dev/pkg/hcltools"
	"github.com/shalb/cluster.dev/pkg/project"
)

// genBackendCodeBlock generate backend code block for this unit.
func (m *Unit) genBackendCodeBlock() ([]byte, error) {

	f, err := (*m.BackendPtr).GetBackendHCL(m.StackName(), m.Name())
	if err != nil {
		log.Debug(err.Error())
		return nil, err
	}
	if len(m.RequiredProviders) < 1 {
		return f.Bytes(), nil
	}
	tb := f.Body().Blocks()[0]
	tfBlock := tb.Body().AppendNewBlock("required_providers", []string{})
	for name, prov := range m.RequiredProviders {

		reqProvs, err := hcltools.InterfaceToCty(prov)
		if err != nil {
			return nil, err
		}
		tfBlock.Body().SetAttributeValue(name, reqProvs)
	}
	return f.Bytes(), nil
}

// genDepsRemoteStates generate terraform remote states for all dependencies of this unit.
func (m *Unit) genDepsRemoteStates() ([]byte, error) {
	var res []byte
	depsUniq := map[project.Unit]bool{}
	for _, dep := range *m.Dependencies() {
		//log.Warnf("dep: %+v", dep)
		// Ignore duplicated dependencies.
		if _, ok := depsUniq[dep.Unit]; ok {
			continue
		}
		// Ignore dependencies without output (user defined as 'depends_on' option.)
		if dep.Output == "" {
			continue
		}
		// De-duplication.
		depsUniq[dep.Unit] = true
		modBackend := dep.Unit.Stack().Backend
		rs, err := modBackend.GetRemoteStateHCL(dep.Unit.Stack().Name, dep.Unit.Name())
		if err != nil {
			log.Debug(err.Error())
			return nil, err
		}
		res = append(res, rs...)
	}
	return res, nil
}

func (m *Unit) Build() error {
	var err error
	init, err := m.genBackendCodeBlock()
	if err != nil {
		log.Debug(err.Error())
		return err
	}
	if m.Providers != nil {
		providers, err := hcltools.ProvidersToHCL(m.Providers)
		if err != nil {
			log.Debug(err.Error())
			return err
		}
		init = append(init, providers.Bytes()...)
	}

	err = m.CreateFiles.Add("init.tf", string(init), fs.ModePerm)
	if err != nil {
		return fmt.Errorf("build unit %v: %w\n%v", m.Key(), err, m.CreateFiles.SPrint())
	}
	// Create remote_state.tf
	remoteStates, err := m.genDepsRemoteStates()
	if err != nil {
		log.Debug(err.Error())
		return err
	}
	// log.Errorf("Remote states: %v\nUnit name: %v", len(remoteStates), m.Key())
	if len(remoteStates) > 0 {
		err = m.CreateFiles.Add("remote_states.tf", string(remoteStates), fs.ModePerm)
		if err != nil {
			return fmt.Errorf("build unit %v: %w", m.Key(), err)
		}
	}
	if m.PreHook != nil {
		err := m.replaceRemoteStatesForBash(&m.PreHook.Command)
		if err != nil {
			return err
		}
	}
	if m.PostHook != nil {
		err := m.replaceRemoteStatesForBash(&m.PostHook.Command)
		if err != nil {
			return err
		}
	}

	return m.Unit.Build()
}
func (m *Unit) replaceRemoteStatesForBash(cmd *string) error {
	if cmd == nil {
		return nil
	}
	markersList := map[string]*project.DependencyOutput{}
	err := m.Project().GetMarkers(RemoteStateMarkerCatName, &markersList)
	if err != nil {
		return err
	}
	for hash, marker := range markersList {
		if marker.StackName == "this" {
			marker.StackName = m.Stack().Name
		}
		refStr := DependencyToBashRemoteState(marker)
		c := strings.ReplaceAll(*cmd, hash, refStr)
		cmd = &c
	}
	return nil
}
