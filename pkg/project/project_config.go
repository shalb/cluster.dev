package project

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"

	"github.com/apex/log"
	"github.com/shalb/cluster.dev/pkg/config"
	"github.com/shalb/cluster.dev/pkg/utils"
	"gopkg.in/yaml.v3"
)

const defaultProjectName = "default-project"

func (p *Project) parseProjectConfig() error {

	if p.configDataFile == nil {
		p.StateBackendName = "default"
		p.name = defaultProjectName
		return nil
	}
	var prjConfParsed map[string]interface{}
	err := yaml.Unmarshal(p.configDataFile, &prjConfParsed)
	if err != nil {
		return fmt.Errorf("parsing project config: %v", utils.ResolveYamlError(p.configDataFile, err))
	}
	if name, ok := prjConfParsed["name"].(string); !ok {
		return fmt.Errorf("error in project config: name is required")
	} else {
		p.name = name
	}

	if kn, ok := prjConfParsed["kind"].(string); !ok || kn != projectObjKindKey {
		return fmt.Errorf("error in project config: kind is required")
	}

	if exports, ok := prjConfParsed["exports"]; ok {
		err = p.ExportEnvs(exports)
		if err != nil {
			return fmt.Errorf(err.Error())
		}
	}

	if stateBackend, exists := prjConfParsed["backend"].(string); exists {
		p.StateBackendName = stateBackend
	} else {
		log.Warn("Set default project backend")
		p.StateBackendName = "default"
	}

	p.configData["project"] = prjConfParsed
	return nil
}

// Return project conf and slice of others config files.
func (p *Project) readManifests() (err error) {
	var files []string
	files, _ = filepath.Glob(config.Global.WorkingDir + "/*.yaml")
	filesYML, _ := filepath.Glob(config.Global.WorkingDir + "/*.yml")
	files = append(files, filesYML...)
	objFiles := make(map[string][]byte)

	for _, file := range files {
		// log.Warnf("Read Files: %v", file)
		fileName, _ := filepath.Rel(config.Global.WorkingDir, file)
		isProjectConfig := regexp.MustCompile(ConfigFilePattern).MatchString(fileName)
		if isProjectConfig {
			p.configDataFile, err = ioutil.ReadFile(file)
		} else {
			objFiles[file], err = ioutil.ReadFile(file)
		}
		if err != nil {
			return fmt.Errorf("reading configs %v: %v", file, err)
		}
	}
	p.objectsFiles = objFiles
	return nil
}

func (p *Project) ExportEnvs(ex interface{}) error {
	exports, correct := ex.(map[string]interface{})
	if !correct {
		return fmt.Errorf("exports: malformed exports configuration")
	}
	for key, val := range exports {
		log.Debugf("Exports: %v", key)
		valStr := fmt.Sprintf("%v", val)
		os.Setenv(key, valStr)
	}
	return nil
}
