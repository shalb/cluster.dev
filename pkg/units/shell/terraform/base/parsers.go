package base

import (
	"fmt"
	"reflect"

	"github.com/apex/log"
	"github.com/shalb/cluster.dev/pkg/project"
	"github.com/shalb/cluster.dev/pkg/utils"
)

// TerraformJSONParser parse in (expected JSON string)
// and stores it in the value pointed to by out.
func TerraformJSONParser(in string, out *project.UnitLinksT) error {
	if out == nil || out.IsEmpty() {
		log.Debugf("RegexOutputParser: unit has no expected outputs, ignore")
		return nil
	}
	type tfOutputDataSpec struct {
		Sensitive bool        `json:"sensitive"`
		Type      interface{} `json:"type"`
		Value     interface{} `json:"value"`
	}
	tfOutputData := make(map[string]tfOutputDataSpec)
	err := utils.JSONDecode([]byte(in), &tfOutputData)
	if err != nil {
		return err
	}

	outTmp := make(map[string]string)
	for key, val := range tfOutputData {
		tp := reflect.ValueOf(val.Type)
		if tp.Kind() != reflect.String || val.Type.(string) != "string" {
			log.Warnf("parse terraform outputs: the value is not in string format! we will convert it to string, but it is recommended to use remote states instead of outputs")
		}
		strValue := fmt.Sprintf("%v", val.Value)
		outTmp[key] = strValue
	}
	for _, expOutput := range out.List {
		data, exists := outTmp[expOutput.OutputName]
		if !exists {
			return fmt.Errorf("parse outputs: unit has no output named '%v', expected by another unit", expOutput.OutputName)
		}
		expOutput.OutputData = data
	}
	return nil
}
