package project

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"

	"github.com/apex/log"
	"github.com/shalb/cluster.dev/pkg/config"
	"github.com/shalb/cluster.dev/pkg/utils"
)

func (p *Project) SaveState() error {
	st := stateData{
		Markers: p.Markers,
		Modules: map[string]interface{}{},
	}
	for key, mod := range p.Modules {
		st.Modules[key] = mod.GetState()
	}
	buffer := &bytes.Buffer{}
	encoder := json.NewEncoder(buffer)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent(" ", " ")
	err := encoder.Encode(st)
	if err != nil {
		return fmt.Errorf("saving project state: %v", err.Error())
	}
	return ioutil.WriteFile(config.Global.StateFileName, buffer.Bytes(), fs.ModePerm)
}

type stateData struct {
	Markers map[string]interface{} `json:"markers"`
	Modules map[string]interface{} `json:"modules"`
}

type StateProject struct {
	Project
}

func (p *Project) LoadState() (*StateProject, error) {
	if _, err := os.Stat(config.Global.StateCacheDir); os.IsNotExist(err) {
		err := os.Mkdir(config.Global.StateCacheDir, 0755)
		if err != nil {
			return nil, err
		}
	}
	err := removeDirContent(config.Global.StateCacheDir)
	if err != nil {
		return nil, err
	}

	stateD := stateData{}

	loadedStateFile, err := ioutil.ReadFile(config.Global.StateFileName)
	if err == nil {
		err = utils.JSONDecode(loadedStateFile, &stateD)
		if err != nil {
			return nil, err
		}
	}
	statePrj := StateProject{
		Project: Project{
			name:            p.Name(),
			secrets:         p.secrets,
			configData:      p.configData,
			configDataFile:  p.configDataFile,
			objects:         p.objects,
			Modules:         make(map[string]Module),
			Markers:         stateD.Markers,
			Infrastructures: make(map[string]*Infrastructure),
			Backends:        p.Backends,
			codeCacheDir:    config.Global.StateCacheDir,
		},
	}

	for mName, mState := range stateD.Modules {
		log.Debugf("Loading module from state: %v", mName)

		key, exists := mState.(map[string]interface{})["type"]
		if !exists {
			return nil, fmt.Errorf("loading state: internal error: bad module type in state")
		}
		mod, err := ModuleFactoriesMap[key.(string)].NewFromState(mState.(map[string]interface{}), mName, p)
		if err != nil {
			return nil, fmt.Errorf("loading state: error loading module from state: %v", err.Error())
		}
		statePrj.Modules[mName] = mod
	}
	err = statePrj.prepareModules()
	if err != nil {
		return nil, err
	}
	return &statePrj, nil
}
