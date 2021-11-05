package tfmodule

import (
	"encoding/base64"
	"fmt"

	"github.com/shalb/cluster.dev/pkg/modules/shell/terraform/base"
	"github.com/shalb/cluster.dev/pkg/project"
	"github.com/shalb/cluster.dev/pkg/utils"
)

type State struct {
	base.StateSpec
	Source      string            `json:"source"`
	Version     string            `json:"version,omitempty"`
	ModType     string            `json:"type"`
	Inputs      interface{}       `json:"inputs,omitempty"`
	LocalModule map[string]string `json:"local_module"`
}

type StateDiff struct {
	base.StateSpecDiff
	Source      string            `json:"source"`
	Version     string            `json:"version,omitempty"`
	Inputs      interface{}       `json:"inputs,omitempty"`
	LocalModule map[string]string `json:"local_module"`
}

func (m *Unit) GetState() interface{} {
	st := m.GetState()
	stTf := State{
		StateSpec: st.(base.StateSpec),
		Inputs:    m.inputs,
		ModType:   m.KindKey(),
		Source:    m.source,
		Version:   m.version,
	}
	if m.localUnit != nil {
		stTf.LocalModule = make(map[string]string)
		for dir, file := range m.localUnit {
			stTf.LocalModule[dir] = base64.StdEncoding.EncodeToString(file)
		}
	}
	return stTf
}

func (m *Unit) GetDiffData() interface{} {
	st := m.GetStateDiff()
	stTf := StateDiff{
		StateSpecDiff: st,
		Inputs:        m.inputs,
		Source:        m.source,
		Version:       m.version,
	}
	if m.localUnit != nil && utils.IsLocalPath(m.source) {
		stTf.LocalModule = make(map[string]string)
		for dir, file := range m.localUnit {
			stTf.LocalModule[dir] = base64.StdEncoding.EncodeToString(file)
		}
	}
	diffData := map[string]interface{}{}
	res := map[string]interface{}{}
	utils.JSONCopy(stTf, &diffData)
	m.ReplaceRemoteStatesForDiff(diffData, &res)
	return res
}

func (s *State) GetType() string {
	return s.ModType
}

func (m *Unit) LoadState(stateData interface{}, modKey string, p *project.StateProject) error {
	s := State{}
	err := utils.JSONCopy(stateData, &s)
	if err != nil {
		return fmt.Errorf("load state: %v", err.Error())
	}
	inputs, ok := s.Inputs.(map[string]interface{})
	if !ok {
		m.inputs = nil
	} else {
		m.inputs = inputs
	}
	m.source = s.Source
	m.version = s.Version
	err = m.Unit.LoadState(s.StateSpec, modKey, p)
	if err != nil {
		return fmt.Errorf("load state: %v", err.Error())
	}
	if utils.IsLocalPath(m.source) {
		m.localUnit = make(map[string][]byte)
		for dir, file := range s.LocalModule {
			decodedFile, err := base64.StdEncoding.DecodeString(file)
			if err != nil {
				return fmt.Errorf("load state: %v", err.Error())
			}
			m.localUnit[dir] = decodedFile
		}
	}
	return nil
}
