package project

import (
  "fmt"
  "testing"
)

func TestFindLineNumInYamlError(t *testing.T) {
  err := fmt.Errorf("yaml: line 24: mapping values are not allowed in this context")
  list := FindLineNumInYamlError(err)
  expectedNum := 24
  if len(list) != 1 || list[0] != expectedNum {
    t.Errorf("Expected: [%v], actual value: %v", expectedNum, list)
    t.Fail()
  }

  err = fmt.Errorf("yaml: line 24: mapping values are not allowed in this context yaml: line 99: mapping values are not allowed in this context")
  list = FindLineNumInYamlError(err)
  expectedNum1 := 24
  if len(list) != 1 || list[0] != expectedNum1 {
    t.Errorf("Expected: [%v], actual value: %v", expectedNum1, list)
    t.Fail()
  }
}

func TestNewYamlError(t *testing.T) {
  testList := []byte(`line1
line2
line3
line4
line5
line6
line7
line8
line9
line10`)
  resultWith0 := NewYamlError(testList, []int{1})
  expectedResult := `, Details:
1: line1   <<<<<<<<<
2: line2
3: line3
4: line4`
  if resultWith0 != expectedResult {
    t.Errorf("Expected: [%v], actual value: [%v]", expectedResult, resultWith0)
    t.Fail()
  }
  resultWith5 := NewYamlError(testList, []int{5})
  expectedResult = `, Details:
3: line3
4: line4
5: line5   <<<<<<<<<
6: line6
7: line7
8: line8
9: line9`
  if resultWith5 != expectedResult {
    t.Errorf("Expected: [%v], actual value: [%v]", expectedResult, resultWith5)
    t.Fail()
  }
}
