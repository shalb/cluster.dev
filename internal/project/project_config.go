package project

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/apex/log"
	"github.com/shalb/cluster.dev/internal/config"
	"github.com/shalb/cluster.dev/pkg/utils"
	"gopkg.in/yaml.v3"
)

const defaultProjectName = "default-project"
const ignoreFileName = ".cdevignore"

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
		return fmt.Errorf("error in project config: backend is not defined. To use default local backend, set 'backend: default' option")
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

	ignoreFileFullPath := filepath.Join(config.Global.WorkingDir, ignoreFileName)
	ignoreData, _ := os.ReadFile(ignoreFileFullPath) // Ignore error, its ok
	ignoreList := strings.Split(string(ignoreData), "\n")

	ignoreFileCheck := func(filename string) bool {
		for _, ignoreFile := range ignoreList {
			if ignoreFile == filename {
				return true
			}
		}
		return false
	}

	for _, file := range files {
		// log.Warnf("Read Files: %v, list: %v", file, ignoreList)
		fileName, _ := filepath.Rel(config.Global.WorkingDir, file)
		if ignoreFileCheck(fileName) {
			log.Debugf("File ignored: %v", fileName)
			continue
		}
		isProjectConfig := regexp.MustCompile(ConfigFilePattern).MatchString(fileName)
		if isProjectConfig {
			p.configDataFile, err = os.ReadFile(file)
		} else {
			objFiles[file], err = os.ReadFile(file)
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
