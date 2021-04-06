package hcltools

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/apex/log"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/zclconf/go-cty/cty"
	ctyjson "github.com/zclconf/go-cty/cty/json"
	"gopkg.in/yaml.v3"
)

// Describe witch provider key is block, no value.
const providersSpec = `
kubernetes:
  exec: true
helm:
  kubernetes:
    exec: true
aws:
  assume_role: true
  ignore_tags: true
google:
  batching: true
google-beta:
  batching: true
`

// InterfaceToCty convert go type tu cty.Value(for hlc lib), using hack with JSON marshal/unmarshal.
func InterfaceToCty(in interface{}) (cty.Value, error) {
	js, err := json.Marshal(in)
	if err != nil {
		return cty.NilVal, err
	}
	var ctyVal ctyjson.SimpleJSONValue
	err = ctyVal.UnmarshalJSON(js)
	if err != nil {
		return cty.NilVal, err
	}

	return ctyVal.Value, nil
}

// ReplaceStingMarkerInBody replace all substrings in TokenQuotedLit tokens to value as terraform expression.
func ReplaceStingMarkerInBody(body *hclwrite.Body, marker, value string) {

	for _, bl := range body.Blocks() {
		ReplaceStingMarkerInBody(bl.Body(), marker, value)
	}
	attrs := body.Attributes()
	for name, attr := range attrs {
		var cleanedExprTokens hclwrite.Tokens
		tokens := attr.Expr().BuildTokens(nil)
		cleanedExprTokens = replaceStingMarker(&tokens, marker, value)
		body.SetAttributeRaw(name, cleanedExprTokens)
	}
}

func replaceStingMarker(tokens *hclwrite.Tokens, marker, value string) hclwrite.Tokens {
	res := hclwrite.Tokens{}
	ignoreNext := false
	for _, tok := range *tokens {
		if ignoreNext {
			ignoreNext = false
			continue
		}
		if tok.Type == hclsyntax.TokenQuotedLit && strings.Contains(string(tok.Bytes), marker) {
			changeTo := value
			if len(tok.Bytes) > len(marker) {
				changeTo = fmt.Sprintf("${%v}", value)
				replacer := strings.NewReplacer(marker, changeTo)
				changeTo = replacer.Replace(string(tok.Bytes))
				res = append(res, &hclwrite.Token{Type: hclsyntax.TokenQuotedLit, Bytes: []byte(changeTo), SpacesBefore: 0})
			} else {
				res[len(res)-1].Type = hclsyntax.TokenIdent
				res[len(res)-1].Bytes = []byte(changeTo)
				ignoreNext = true
			}
			continue
		}
		newToken := *tok
		res = append(res, &newToken)
	}
	return res
}

// CreateTokensForOutput create slice of tokens, from splitted by dot substrings of in.
func CreateTokensForOutput(in string) hclwrite.Tokens {
	var res hclwrite.Tokens = hclwrite.Tokens{}
	splittedStr := strings.Split(in, ".")
	for elem, substr := range splittedStr {
		res = append(res, &hclwrite.Token{
			Type:         hclsyntax.TokenIdent,
			Bytes:        []byte(substr),
			SpacesBefore: 0,
		})
		if elem != len(splittedStr)-1 {
			res = append(res, &hclwrite.Token{
				Type:         hclsyntax.TokenDot,
				Bytes:        []byte{'.'},
				SpacesBefore: 0,
			})
		}
	}
	return res
}

func ProvidersToHCL(in interface{}) (*hclwrite.File, error) {
	provSpecConf := map[string]interface{}{}
	err := yaml.Unmarshal([]byte(providersSpec), provSpecConf)
	if err != nil {
		log.Debugf("Internal error ProvidersToHCL: %v", providersSpec)
		return nil, fmt.Errorf("ProvidersToHCL: %v", err.Error())
	}
	data, ok := in.([]interface{})
	if !ok {
		log.Debug("Malformed provider configuration")
		return nil, fmt.Errorf("providerToHCL: malformed provider configuration")
	}

	f := hclwrite.NewEmptyFile()
	rootBody := f.Body()
	for _, block := range data {
		providerMap, ok := block.(map[string]interface{})
		if !ok || len(providerMap) != 1 {
			log.Debug("Malformed provider configuration")
			return nil, fmt.Errorf("providerToHCL: malformed provider configuration")
		}
		for provName, provSpec := range providerMap {
			provBlock := rootBody.AppendNewBlock("provider", []string{provName})
			provSubSpec, _ := provSpecConf[provName]
			if _, ok := provSpec.(map[string]interface{}); !ok {
				log.Debug("Malformed provider configuration")
				return nil, fmt.Errorf("providerToHCL: malformed provider configuration")
			}
			err := fillBlock(provSpec.(map[string]interface{}), provBlock, provSubSpec)
			if err != nil {
				return nil, err
			}
		}
	}
	return f, nil
}
func fillBlock(data map[string]interface{}, result *hclwrite.Block, configMap interface{}) error {
	for key, val := range data {
		if _, ok := val.(map[string]interface{}); ok {
			if isBlock, subCat := checkHCLConfig(key, configMap); isBlock {
				bl := result.Body().AppendNewBlock(key, []string{})
				err := fillBlock(val.(map[string]interface{}), bl, subCat)
				if err != nil {
					log.Debug(err.Error())
					return err
				}
				continue
			}
		}
		dt, err := InterfaceToCty(val)
		if err != nil {
			log.Debug(err.Error())
			return err
		}
		result.Body().SetAttributeValue(key, dt)
	}
	return nil
}

func checkHCLConfig(key string, configMap interface{}) (bool, interface{}) {
	if configMap == nil {
		return false, nil
	}
	checkMap, ok := configMap.(map[string]interface{})
	if !ok {
		return false, nil
	}
	res, exists := checkMap[key]
	if !exists {
		return false, nil
	}
	return true, res
}
