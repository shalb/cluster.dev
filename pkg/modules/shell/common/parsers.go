package common

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/apex/log"
	"github.com/shalb/cluster.dev/pkg/project"
	"github.com/shalb/cluster.dev/pkg/utils"
)

// JSONOutputParser parse in (expected JSON string)
// and stores it in the value pointed to by out.
func (u *Unit) JSONOutputParser(in string, out *project.UnitLinksT) error {
	if out == nil || out.IsEmpty() {
		log.Debugf("RegexOutputParser: unit has no expected outputs, ignore")
		return nil
	}
	outTmp := make(map[string]interface{})

	err := utils.JSONDecode([]byte(in), outTmp)
	if err != nil {
		return err
	}
	for _, expOutput := range out.List {
		data, exists := outTmp[expOutput.OutputName]
		if !exists {
			return fmt.Errorf("unit has no output named '%v', expected by another unit", expOutput.OutputName)
		}
		expOutput.OutputData = data
	}
	return nil
}

// RegexOutputParser parse each line od in string with key/value regexp
// and stores result as a map in the value pointed to by out.
func (u *Unit) RegexOutputParser(in string, out *project.UnitLinksT) error {
	if out == nil || out.IsEmpty() {
		log.Debugf("RegexOutputParser: unit has no expected outputs, ignore")
		return nil
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
		re, err := regexp.Compile(u.GetOutputsConf.Regexp)
		if err != nil {
			return err
		}
		parsed := re.FindStringSubmatch(ln)
		if len(parsed) < 2 {
			// ignore "not found" and show warn
			log.Warnf("can't parse the output string '%v' with regexp '%v'", ln, u.GetOutputsConf.Regexp)
			continue
		}
		// Use first occurrence as key and value.
		outTmp[parsed[1]] = parsed[2]
	}
	for _, expOutput := range out.List {
		data, exists := outTmp[expOutput.OutputName]
		if !exists {
			return fmt.Errorf("unit has no output named '%v', expected by another unit", expOutput.OutputName)
		}
		expOutput.OutputData = data
	}
	return nil
}

// SeparatorOutputParser split each line of in string with using separator
// and stores result as a map in the value pointed to by out.
func (u *Unit) SeparatorOutputParser(in string, out *project.UnitLinksT) error {
	if out == nil || out.IsEmpty() {
		log.Debugf("RegexOutputParser: unit has no expected outputs, ignore")
		return nil
	}
	lines := strings.Split(in, "\n")
	if len(lines) == 0 {
		return nil
	}
	outTmp := make(map[string]string)
	for _, ln := range lines {
		// log.Warn(ln)
		if len(ln) == 0 {
			// ignore empty string
			continue
		}
		kv := strings.SplitN(ln, u.GetOutputsConf.Separator, 2)
		if len(kv) != 2 || len(ln) < len(u.GetOutputsConf.Separator) {
			// ignore line if separator does not found
			log.Warnf("can't parse the output string '%v' , separator '%v' does not found", ln, u.GetOutputsConf.Separator)
			continue
		}
		key := strings.Trim(kv[0], " ")
		outTmp[key] = kv[1]
	}
	for _, expOutput := range out.List {
		data, exists := outTmp[expOutput.OutputName]
		if !exists {
			return fmt.Errorf("unit has no output named '%v', expected by another unit", expOutput.OutputName)
		}
		expOutput.OutputData = data
	}
	return nil
}
