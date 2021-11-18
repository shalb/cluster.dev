package utils

import (
	"bytes"
	"fmt"

	"github.com/apex/log"
	"gopkg.in/yaml.v3"
)

func ReadYAMLObjects(objData []byte) ([]map[string]interface{}, error) {
	objects := []map[string]interface{}{}
	dec := yaml.NewDecoder(bytes.NewReader(objData))
	for {
		var parsedConf = make(map[string]interface{})
		err := dec.Decode(&parsedConf)
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			log.Debugf("can't decode config to yaml: %s", err.Error())
			return nil, fmt.Errorf("can't decode config to yaml: %s", ResolveYamlError(objData, err))
		}
		objects = append(objects, parsedConf)
	}
	return objects, nil
}

// ReadYAML same as ReadYAMLObjects but parse only data with 1 yaml object.
func ReadYAML(objData []byte) (res map[string]interface{}, err error) {
	err = yaml.Unmarshal(objData, &res)
	err = ResolveYamlError(objData, err)
	return
}
