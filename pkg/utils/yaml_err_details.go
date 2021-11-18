package utils

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

func FindLineNumInYamlError(err error) []int {
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
	if num-3 > 0 {
		startPosition = num - 3
	} else {
		startFilecorrection = num - 4
	}
	for _, k := range strings.Split(string(data), "\n")[startPosition:] {
		lineNum := startPosition + 1 + iterator
		lineText := fmt.Sprintf("%v: %s", lineNum, k)
		if iterator == num-startPosition-1 {
			lineText = lineText + "   <<<<<<<<<"
		}
		errString = append(errString, lineText)
		iterator++
		if iterator == 7+startFilecorrection {
			break
		}
	}
	return strings.Join(errString, "\n")
}

func ResolveYamlError(data []byte, err error) error {
	if err != nil {
		errNum := FindLineNumInYamlError(err)
		yamlErr := NewYamlError(data, errNum)
		return fmt.Errorf("%v %v", err.Error(), yamlErr)
	}
	return nil
}
