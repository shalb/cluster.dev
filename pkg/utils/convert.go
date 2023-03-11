package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"

	"gopkg.in/yaml.v3"
)

// JSONCopy - convert interface data in to type of out with JSON tags.
func JSONCopy(in, out interface{}) error {
	// t := reflect.TypeOf(out)
	if reflect.ValueOf(out).IsNil() {
		return fmt.Errorf("can't write to nil target")
	}

	buffer := &bytes.Buffer{}
	encoder := json.NewEncoder(buffer)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent(" ", " ")
	err := encoder.Encode(in)
	if err != nil {
		return err
	}
	//log.Warnf("JSON: %v", buffer.String())
	decoder := json.NewDecoder(buffer)
	err = decoder.Decode(&out)
	if err != nil {
		return err
	}
	return nil
}

// YAMLInterfaceToType - convert interface data in to type of out with YAML tags.
func YAMLInterfaceToType(in, out interface{}) error {
	res, err := yaml.Marshal(in)
	if err != nil {
		return err
	}
	return ResolveYamlError(res, yaml.Unmarshal(res, out))
}

// JSONEncode encode in to JSON.
func JSONEncode(in interface{}) ([]byte, error) {
	buffer := &bytes.Buffer{}
	encoder := json.NewEncoder(buffer)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent(" ", " ")
	err := encoder.Encode(in)
	return buffer.Bytes(), err
}

// JSONEncode encode in to JSON.
func JSONEncodeString(in interface{}) (string, error) {
	buffer := &bytes.Buffer{}
	encoder := json.NewEncoder(buffer)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent(" ", " ")
	err := encoder.Encode(in)
	return buffer.String(), err
}

// JSONDecode decode in to .
func JSONDecode(in []byte, out interface{}) error {
	buffer := bytes.NewBuffer(in)
	decoder := json.NewDecoder(buffer)
	return decoder.Decode(&out)
}

func MergeMaps(mOne, mTwo map[string]interface{}) (res map[string]interface{}) {
	for k, v := range mOne {
		res[k] = v
	}
	for k, v := range mTwo {
		res[k] = v
	}
	return
}
// TerraformJSONOutputParse parse data from terraform output --json command to map and line-to-line string
func TerraformJSONOutputParse(in string) (out map[string]string, stringOut string, err error) {
	type tfOutputDataSpec struct {
		Sensitive bool        `json:"sensitive"`
		Type      interface{} `json:"type"`
		Value     interface{} `json:"value"`
	}
	tfOutputData := make(map[string]tfOutputDataSpec)
	err = JSONDecode([]byte(in), &tfOutputData)
	if err != nil {
		return
	}
  out = make(map[string]string)
	for key, val := range tfOutputData {
    tp := reflect.ValueOf(val.Type)
    var strValue string
    if tp.Kind() != reflect.String || val.Type.(string) != "string" {
			strValue, err = JSONEncodeString(val.Value)
      if err != nil {
        return nil, "", fmt.Errorf("JSON encode output value: %w", err)
      }
		} else {
      strValue = fmt.Sprintf("%v", val.Value)
    }
		out[key] = strValue
    stringOut += fmt.Sprintf("%s = %s\n", key, strValue)
	}
	return
}
