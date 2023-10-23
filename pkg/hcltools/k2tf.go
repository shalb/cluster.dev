package hcltools

import (
	"reflect"
	"strings"

	"github.com/apex/log"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/iancoleman/strcase"
	"github.com/jinzhu/inflection"
	"github.com/mitchellh/reflectwalk"
	"gopkg.in/yaml.v2"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"

	// "k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes/scheme"
)

func Kubernetes2HCLCustom(manifest interface{}, key string, rootBody *hclwrite.Body) error {
	unitBlock := rootBody.AppendNewBlock("resource", []string{"kubernetes_manifest", key})
	unitBody := unitBlock.Body()
	tokens := hclwrite.Tokens{&hclwrite.Token{Type: hclsyntax.TokenQuotedLit, Bytes: []byte(" kubernetes"), SpacesBefore: 1}}
	unitBody.SetAttributeRaw("provider", tokens)
	ctyVal, err := InterfaceToCty(manifest)
	if err != nil {
		return err
	}
	unitBody.SetAttributeValue("manifest", ctyVal)
	return nil
}

func Kubernetes2HCL(manifest interface{}, dst *hclwrite.Body) error {
	manifestRaw, err := yaml.Marshal(manifest)
	if err != nil {
		return err
	}
	d := scheme.Codecs.UniversalDeserializer()
	obj, _, err := d.Decode([]byte(manifestRaw), nil, nil)
	if err != nil {
		log.Debug("Could not decode YAML object, malformed manifest or unknown k8s API/Kind, generating common resource")
		return err
	}
	w, err := NewObjectWalker(obj, dst)
	if err != nil {
		return err
	}

	return reflectwalk.Walk(obj, w)
}

func ToTerraformSubBlockName(field *reflect.StructField, path string) string {
	name := extractProtobufName(field)

	return NormalizeTerraformName(name, true, path)
}

func ToTerraformAttributeName(field *reflect.StructField, path string) string {
	name := extractProtobufName(field)

	return NormalizeTerraformName(name, false, path)
}

// ToTerraformResourceType converts a Kubernetes API Object Type name to the
// equivalent `terraform-provider-kubernetes` schema name.
// Src: https://github.com/sl1pm4t/k2tf
func ToTerraformResourceType(obj runtime.Object, blockKind *schema.GroupVersionKind) string {
	if blockKind == nil {
		return ""
	}
	kind := blockKind.Kind
	switch kind {
	case "Ingress":
		if blockKind.Version == "networking.k8s.io/v1" {
			kind = "ingress_v1"
		} else {
			kind = "ingress"
		}
	default:
		kind = NormalizeTerraformName(blockKind.Kind, false, "")
	}
	return "kubernetes_" + kind
}

// normalizeTerraformName converts the given string to snake case
// and optionally to singular form of the given word
// s is the string to normalize
// set toSingular to true to singularize the given word
// path is the full schema path to the named element
// Src: https://github.com/sl1pm4t/k2tf
func NormalizeTerraformName(s string, toSingular bool, path string) string {
	switch s {
	case "DaemonSet":
		return "daemonset"

	case "nonResourceURLs":
		if strings.Contains(path, "role.rule") {
			return "non_resource_urls"
		}

	case "updateStrategy":
		if !strings.Contains(path, "stateful") {
			return "strategy"
		}

	case "limits":
		if strings.Contains(path, "limit_range.spec") {
			return "limit"
		}

	case "ports":
		if strings.Contains(path, "kubernetes_network_policy.spec") {
			return "ports"
		}

	case "externalIPs":
		if strings.Contains(path, "kubernetes_service.spec") {
			return "external_ips"
		}
	}

	if toSingular {
		s = inflection.Singular(s)
	}
	s = strcase.ToSnake(s)

	// colons and dots are not allowed by Terraform
	s = strings.ReplaceAll(s, ":", "_")
	s = strings.ReplaceAll(s, ".", "_")

	return s
}

func extractProtobufName(field *reflect.StructField) string {
	protoTag := field.Tag.Get("protobuf")
	if protoTag == "" {
		log.Warnf("field [%s] has no protobuf tag", field.Name)
		return field.Name
	}

	tagParts := strings.Split(protoTag, ",")
	for _, part := range tagParts {
		if strings.Contains(part, "name=") {
			return part[5:]
		}
	}

	log.Warnf("field [%s] protobuf tag has no 'name'", field.Name)
	return field.Name
}
