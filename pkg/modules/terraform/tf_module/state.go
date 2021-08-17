package tfmodule

import (
	"encoding/base64"
	"fmt"

	"github.com/apex/log"
	"github.com/shalb/cluster.dev/pkg/modules/terraform/common"
	"github.com/shalb/cluster.dev/pkg/project"
	"github.com/shalb/cluster.dev/pkg/utils"
)

type State struct {
	common.StateSpecCommon
	Source      string            `json:"source"`
	Version     string            `json:"version,omitempty"`
	ModType     string            `json:"type"`
	Inputs      interface{}       `json:"inputs,omitempty"`
	LocalModule map[string]string `json:"local_module"`
}

type StateDiff struct {
	common.StateSpecDiffCommon
	Source      string            `json:"source"`
	Version     string            `json:"version,omitempty"`
	Inputs      interface{}       `json:"inputs,omitempty"`
	LocalModule map[string]string `json:"local_module"`
}

func (m *Module) GetState() interface{} {
	st := m.GetStateCommon()
	stTf := State{
		StateSpecCommon: st,
		Inputs:          m.inputs,
		ModType:         m.KindKey(),
		Source:          m.source,
		Version:         m.version,
	}
	if m.localModule != nil {
		stTf.LocalModule = make(map[string]string)
		for dir, file := range m.localModule {
			stTf.LocalModule[dir] = base64.StdEncoding.EncodeToString(file)
		}
	}
	return stTf
}

func (m *Module) GetDiffData() interface{} {
	st := m.GetStateDiffCommon()
	stTf := StateDiff{
		StateSpecDiffCommon: st,
		Inputs:              m.inputs,
		Source:              m.source,
		Version:             m.version,
	}
	if m.localModule != nil && utils.IsLocalPath(m.source) {
		stTf.LocalModule = make(map[string]string)
		for dir, file := range m.localModule {
			stTf.LocalModule[dir] = base64.StdEncoding.EncodeToString(file)
		}
	}
	diffData := map[string]interface{}{}
	res := map[string]interface{}{}
	utils.JSONInterfaceToType(stTf, &diffData)
	m.ReplaceRemoteStatesForDiff(diffData, &res)
	return res
}

func (s *State) GetType() string {
	return s.ModType
}

func (m *Module) LoadState(stateData interface{}, modKey string, p *project.StateProject) error {
	s := State{}
	err := utils.JSONInterfaceToType(stateData, &s)
	if err != nil {
		return fmt.Errorf("load state: %v", err.Error())
	}
	inputs, ok := s.Inputs.(map[string]interface{})
	if !ok {
		m.inputs = make(map[string]interface{})
	} else {
		m.inputs = inputs
	}
	m.source = s.Source
	m.version = s.Version
	err = m.LoadStateCommon(s.StateSpecCommon, modKey, p)
	if err != nil {
		return fmt.Errorf("load state: %v", err.Error())
	}
	if utils.IsLocalPath(m.source) {
		m.localModule = make(map[string][]byte)
		for dir, file := range s.LocalModule {
			decodedFile, err := base64.StdEncoding.DecodeString(file)
			if err != nil {
				return fmt.Errorf("load state: %v", err.Error())
			}
			m.localModule[dir] = decodedFile
		}
	}
	log.Debugf("%+v", s.LocalModule)
	return nil
}
