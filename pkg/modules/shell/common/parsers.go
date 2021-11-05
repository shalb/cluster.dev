package common

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"github.com/apex/log"
	"github.com/shalb/cluster.dev/pkg/utils"
)

// JSONOutputParser parse in (expected JSON string)
// and stores it in the value pointed to by out.
func (m *Unit) JSONOutputParser(in string, out interface{}) error {
	return utils.JSONDecode([]byte(in), out)
}

// RegexOutputParser parse each line od in string with key/value regexp
// and stores result as a map in the value pointed to by out.
func (m *Unit) RegexOutputParser(in string, out interface{}) error {
	rv := reflect.ValueOf(out)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return fmt.Errorf("can't set unaddressable value")
	}
	lines := strings.Split(in, "\n")
	if len(lines) == 0 {
		return nil
	}

	outTmp := make(map[string]string)
	for _, ln := range lines {
		if len(ln) == 0 {
			// ignore empty string
			continue
		}
		re, err := regexp.Compile(m.GetOutputsConf.Regexp)
		if err != nil {
			return err
		}
		parsed := re.FindStringSubmatch(ln)
		// log.Warnf("Regexp: %v %q", m.GetOutputsConf.Regexp, re)
		if len(parsed) < 2 {
			// ignore "not found" and show warn
			log.Warnf("can't parse the output string '%v' with regexp '%v'", ln, m.GetOutputsConf.Regexp)
			continue
		}
		// Use first occurrence as key and value.
		outTmp[parsed[1]] = parsed[2]
	}
	rv.Elem().Set(reflect.ValueOf(outTmp))
	return nil
}

// SeparatorOutputParser split each line of in string with using separator
// and stores result as a map in the value pointed to by out.
func (m *Unit) SeparatorOutputParser(in string, out interface{}) error {
	rv := reflect.ValueOf(out)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return fmt.Errorf("can't set unadresseble value")
	}
	lines := strings.Split(in, "\n")
	if len(lines) == 0 {
		return nil
	}
	outTmp := make(map[string]string)
	for _, ln := range lines {
		if len(ln) == 0 {
			// ignore empty string
			continue
		}
		kv := strings.SplitN(ln, m.GetOutputsConf.Separator, 2)
		if len(kv) != 2 || len(ln) < len(m.GetOutputsConf.Separator) {
			// ignore line if separator does not found
			log.Warnf("can't parse the output string '%v' , separator '%v' does not found", ln, m.GetOutputsConf.Separator)
			continue
		}
		outTmp[kv[0]] = kv[1]
	}
	rv.Elem().Set(reflect.ValueOf(outTmp))
	return nil
}
