package hcltools

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/zclconf/go-cty/cty"
	ctyjson "github.com/zclconf/go-cty/cty/json"
)

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
