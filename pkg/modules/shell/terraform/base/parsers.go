package base

import (
	"fmt"
	"reflect"

	"github.com/apex/log"
	"github.com/shalb/cluster.dev/pkg/utils"
)

// TerraformJSONParser parse in (expected JSON string)
// and stores it in the value pointed to by out.
func TerraformJSONParser(in string, out interface{}) error {
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

	rv := reflect.ValueOf(out)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return fmt.Errorf("can't set unaddressable value")
	}

	outTmp := make(map[string]string)
	for key, val := range tfOutputData {
		tp := reflect.ValueOf(val.Type)
		if tp.Kind() != reflect.String || val.Type.(string) != "string" {
			log.Warnf("parse terraform outputs: value not in string format! we will convert it to string, but it is recommended to use remote states instead of outputs")
		}
		strValue := fmt.Sprintf("%v", val.Value)
		outTmp[key] = strValue
	}
	rv.Elem().Set(reflect.ValueOf(outTmp))
	return nil
}
