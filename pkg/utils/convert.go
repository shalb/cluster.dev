package utils

import (
	"bytes"
	"encoding/json"
)

// JSONInterfaceToType - convert interface data in to type of out with JSON tags.
func JSONInterfaceToType(in, out interface{}) error {
	buffer := &bytes.Buffer{}
	encoder := json.NewEncoder(buffer)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent(" ", " ")
	err := encoder.Encode(in)
	if err != nil {
		return err
	}

	decoder := json.NewDecoder(buffer)
	err = decoder.Decode(&out)
	if err != nil {
		return err
	}
	return nil
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

// JSONDecode decode in to .
func JSONDecode(in []byte, out interface{}) error {
	return json.Unmarshal(in, out)
}
