package project

import (
  "fmt"
  "github.com/Masterminds/semver"
  "github.com/apex/log"
  "github.com/shalb/cluster.dev/pkg/config"
  "gopkg.in/yaml.v3"
  "regexp"
  "strconv"
  "strings"
)

const infraTemplateObjKindKey = "InfraTemplate"

type InfraTemplate struct {
	Name             string                   `yaml:"name"`
	Kind             string                   `yaml:"kind"`
	Modules          []map[string]interface{} `yaml:"modules"`
	ReqClientVersion string                   `yaml:"cliVersion"`
}

func FindLineNumInYamlError (err error) []int {
  var list []int
  re := regexp.MustCompile(`line (\d+):`)
  submatch := re.FindStringSubmatch(err.Error())
  if len(submatch) > 0 {
    for _, lineStr := range submatch[1:] {
      lineNum, err := strconv.Atoi(lineStr)
      if err == nil {
        list = append(list, lineNum)
      }
    }
  }
  return list
}

func NewYamlError(data []byte, lines []int) string {
  if len(lines) == 0 {
    return ""
  }
  errString := []string{", Details:"}
  num := lines[0] // TODO: Need to check if multiple lines can be returned in the error
  iterator := 0
  startPosition := 0
  startFilecorrection := 0
  if num - 3 > 0 {
    startPosition = num - 3
  } else {
    startFilecorrection = num - 4
  }
  for _, k := range strings.Split(string(data), "\n")[startPosition:] {
    lineNum := startPosition + 1 + iterator
    lineText := fmt.Sprintf("%v: %s", lineNum, k)
    if iterator == num - startPosition - 1 {
      lineText = lineText + "   <<<<<<<<<"
    }
    errString = append(errString, lineText)
    iterator++
    if iterator == 7 + startFilecorrection {
      break
    }
  }
  return strings.Join(errString, "\n")
}

func NewInfraTemplate(data []byte) (*InfraTemplate, error) {
	iTmpl := InfraTemplate{}
	err := yaml.Unmarshal(data, &iTmpl)
	if err != nil {
	  errNum := FindLineNumInYamlError(err)
	  yamlErr := NewYamlError(data, errNum)
		if config.Global.TraceLog {
			log.Debug(string(data))
		}
		return nil, fmt.Errorf("unmarshal template data: %v %v", err.Error(), yamlErr)
	}
	if len(iTmpl.Modules) < 1 {
		return nil, fmt.Errorf("parsing template: template must contain at least one module")
	}
	if iTmpl.Name == "" {
		return nil, fmt.Errorf("parsing template: template must contain 'name' field")
	}
	if iTmpl.Kind != infraTemplateObjKindKey {
		return nil, fmt.Errorf("parsing template: unknown template object kind or kind is not set: '%v'", iTmpl.Kind)
	}
	log.Debug("check client version")
	if iTmpl.ReqClientVersion != "" {
		reqVerConstraints, err := semver.NewConstraint(iTmpl.ReqClientVersion)
		if err != nil {
			return nil, fmt.Errorf("parsing template: can't parse required client version: %v", iTmpl.ReqClientVersion)
		}
		ver, err := semver.NewVersion(config.Global.Version)
		if err != nil {
			// Invalid curent cli version. May be test revision.
			// TODO need check!!
			return nil, fmt.Errorf("parsing template: internalcan't parse client version: %v", iTmpl.ReqClientVersion)
		}
		ok, messages := reqVerConstraints.Validate(ver)
		log.Debugf("Validating version: cli: %v, req: %v", ver, iTmpl.ReqClientVersion)
		if !ok {
			errReasons := ""
			for _, msg := range messages {
				errReasons += fmt.Sprintf("%v\n", msg)
			}
			return nil, fmt.Errorf("cdev template version validation error: \n%v", errReasons)
		}
	}
	// i.TemplateSrc = src
	return &iTmpl, nil
}
