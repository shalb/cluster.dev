package common

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/apex/log"
	"github.com/shalb/cluster.dev/pkg/config"
	"github.com/shalb/cluster.dev/pkg/project"
	"gopkg.in/yaml.v3"
)

func (m *Module) readDeps(depsData interface{}) ([]*project.Dependency, error) {
	rawDepsList := []string{}
	switch depsData.(type) {
	case string:
		rawDepsList = append(rawDepsList, depsData.(string))
	case []string:
		rawDepsList = append(rawDepsList, depsData.([]string)...)
	}
	var res []*project.Dependency
	for _, dep := range rawDepsList {
		splDep := strings.Split(dep, ".")
		if len(splDep) != 2 {
			return nil, fmt.Errorf("Incorrect module dependency '%v'", dep)
		}
		infNm := splDep[0]
		if infNm == "this" {
			infNm = m.InfraName()
		}
		res = append(res, &project.Dependency{
			InfraName:  infNm,
			ModuleName: splDep[1],
		})
		log.Debugf("Dependency added: %v --> %v.%v", m.Key(), infNm, splDep[1])
	}
	return res, nil
}

func readHook(hookData interface{}, hookType string) (*hookSpec, error) {
	hook, ok := hookData.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("%s configuration error", hookType)
	}
	cmd, cmdExists := hook["command"].(string)
	script, scrExists := hook["script"].(string)
	if cmdExists && scrExists {
		return nil, fmt.Errorf("Error in %s config, use 'script' or 'command' option, not both", hookType)
	}
	if !cmdExists && !scrExists {
		return nil, fmt.Errorf("Error in %s config, use one of 'script' or 'command' option", hookType)
	}
	ScriptData := hookSpec{
		Command:   "",
		OnDestroy: false,
		OnApply:   true,
		OnPlan:    false,
	}
	ymlTmp, err := yaml.Marshal(hookData)
	if err != nil {
		log.Debug(err.Error())
		return nil, err
	}
	err = yaml.Unmarshal(ymlTmp, &ScriptData)
	if err != nil {
		log.Debug(err.Error())
		return nil, err
	}
	if cmdExists {
		ScriptData.Command = fmt.Sprintf("#!/usr/bin/env sh\nset -e\n\n%s", cmd)
	} else {
		cmdTmp, err := ioutil.ReadFile(filepath.Join(config.Global.WorkingDir, script))
		ScriptData.Command = string(cmdTmp)
		if err != nil {
			return nil, fmt.Errorf("can't load %s script: %v", hookType, err.Error())
		}
	}
	return &ScriptData, nil

}
