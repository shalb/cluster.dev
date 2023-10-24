package hcltools

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/apex/log"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/shalb/cluster.dev/pkg/hcltools/tfkschema"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/mitchellh/reflectwalk"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/gocty"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/runtime"
)

// WriteObject converts a Kubernetes runtime.Object to HCL
func WriteObject(obj runtime.Object, dst *hclwrite.Body) (int, error) {
	w, err := NewObjectWalker(obj, dst)
	if err != nil {
		return 0, err
	}
	reflectwalk.Walk(obj, w)

	return w.warnCount, nil
}

// ObjectWalker implements reflectwalk.Walker interfaces
// It's used to "walk" the Kubernetes API Objects structure and generate
// an HCL document based on the values defined.
type ObjectWalker struct {
	// The Kubernetes API Object to be walked
	RuntimeObject runtime.Object

	// The HCL body where HCL blocks will be appended
	dst *hclwrite.Body

	// Terraform resource type (e.g. kubernetes_pod)
	resourceType string
	// Terraform resource name (adapted from ObjectMeta name attribute)
	resourceName string

	// top level HCL
	isTopLevel bool

	// sub block tracking
	currentBlock *hclBlock

	// stack of Struct fields
	fields []*reflect.StructField

	// slices of structs
	slices []*reflect.StructField
	// sliceField tracks the reflect.StructField for the current slice
	sliceField *reflect.StructField
	// the stack of the Slice element types that are popped and pushed as we walk through object graph
	sliceElemTypes []reflect.Type
	// Flag to indicate if our reflectwalk functions can skip further processing of slice elements.
	// Slices of primitive values get rendered all at once when we enter the Slice so they don't need
	// further processing for each element.
	ignoreSliceElems bool
	warnCount        int
}

// NewObjectWalker returns a new ObjectWalker object
// dst is the hclwrite.Body where HCL blocks will be appended.
func NewObjectWalker(obj runtime.Object, dst *hclwrite.Body) (*ObjectWalker, error) {
	if obj == nil {
		return nil, fmt.Errorf("obj cannot be nil")
	}

	w := &ObjectWalker{
		RuntimeObject: obj,
		isTopLevel:    true,
		dst:           dst,
	}

	return w, nil
}

func (w *ObjectWalker) setCurrentSlice(f *reflect.StructField) {
	if f != nil {
		// log.Debugf("setting setCurrentSlice to %s", f.Name)
		w.sliceField = f
	}
}

func (w *ObjectWalker) currentSlice() *reflect.StructField {
	if len(w.slices) > 0 {
		return w.slices[len(w.slices)-1]
	}

	return nil
}

func (w *ObjectWalker) fieldPop() *reflect.StructField {
	result := w.fields[len(w.fields)-1]
	w.fields = w.fields[:len(w.fields)-1]

	// log.Debugf("fieldPop %s", result.Name)
	return result
}

func (w *ObjectWalker) fieldPush(v *reflect.StructField) {
	w.fields = append(w.fields, v)
	// log.Debugf("fieldPush %s", v.Name)
}

func (w *ObjectWalker) field() *reflect.StructField {
	if len(w.fields) > 0 {
		f := w.fields[len(w.fields)-1]
		return f
	}
	return nil
}

func (w *ObjectWalker) slicePop() *reflect.StructField {
	result := w.slices[len(w.slices)-1]
	w.slices = w.slices[:len(w.slices)-1]

	//log.Debugf("slicePop %s", result.Name)
	w.setCurrentSlice(result)
	return result
}

func (w *ObjectWalker) slicePush(v *reflect.StructField) {
	w.slices = append(w.slices, v)
	//log.Debugf("slicePush %s", v.Name)
	w.setCurrentSlice(v)
}

func (w *ObjectWalker) sliceType() reflect.Type {
	var result reflect.Type
	currSlice := w.currentSlice()
	if currSlice != nil {
		result = currSlice.Type
		// log.Debugf("sliceType %s", result.Name())
	} else {
		result = reflect.TypeOf(nil)
		// log.Debugf("sliceType nil")
	}

	return result
}

func (w *ObjectWalker) sliceElemTypePush(ty reflect.Type) {
	w.sliceElemTypes = append(w.sliceElemTypes, ty)
	// log.Debugf("sliceElemTypePush %s", ty.Name())
}

func (w *ObjectWalker) sliceElemTypePop() reflect.Type {
	result := w.sliceElemTypes[len(w.sliceElemTypes)-1]
	w.sliceElemTypes = w.sliceElemTypes[:len(w.sliceElemTypes)-1]

	// log.Debugf("sliceElemTypePop %s", result.Name())
	return result
}

func (w *ObjectWalker) sliceElemType() reflect.Type {
	var result reflect.Type
	if len(w.sliceElemTypes) > 0 {
		result = w.sliceElemTypes[len(w.sliceElemTypes)-1]
	} else {
		result = reflect.TypeOf(struct{}{})
	}

	// log.Debugf("sliceElemType %s", result.Name())
	return result
}

// openBlock opens a new HCL resource block or sub-block
// It creates a hclBlock object so we can track hierarchy of blocks
// within the resource tree
func (w *ObjectWalker) openBlock(name, fieldName string, hcl *hclwrite.Block) *hclBlock {
	// log.Debugf("opening hclBlock for field: %s", name)
	b := &hclBlock{
		name:      name,
		fieldName: fieldName,
		parent:    w.currentBlock,
		hcl:       hcl,
	}

	w.currentBlock = b
	return b
}

// closeBlock writes the generated HCL to the hclwriter
func (w *ObjectWalker) closeBlock() *hclBlock {
	// log.Debugf("closing hclBlock: %s", w.currentBlock.name)

	parent := w.currentBlock.parent
	current := w.currentBlock

	// TODO: move append logic to hcl_block to be consistent
	if parent == nil {
		// we are closing the top level block, write directly to HCL File
		w.dst.AppendBlock(current.hcl)

	} else {
		if current.hasValue || tfkschema.IncludedOnZero(w.currentBlock.fieldName) || current.isRequired() {
			if current.unsupported {
				// don't append this block or child blocks
				log.Warnf("excluding attribute [%s] not found in Terraform schema", current.FullSchemaName())
			} else {
				// communicate back up the tree that we found a non-zero value
				parent.hasValue = true

				if current.isMap {
					if len(current.hclMap) > 0 {
						parent.SetAttributeValue(current.name, cty.MapVal(current.hclMap))
					}

				} else if !current.inlined {
					parent.AppendBlock(current.hcl)
				}
			}
		}

		w.currentBlock = parent
	}
	return w.currentBlock
}

// Enter is called by reflectwalk.Walk each time we enter a level
func (w *ObjectWalker) Enter(l reflectwalk.Location) error {
	// log.Debugf(fmt.Sprint("entering ", l))

	return nil
}

// Exit is called by reflectwalk each time it exits from a reflectwalk.Location
func (w *ObjectWalker) Exit(l reflectwalk.Location) error {
	switch l {
	case reflectwalk.Slice:
		if !w.ignoreSliceElems {
			w.sliceElemTypePop()
		}
		w.ignoreSliceElems = false
		w.slicePop()

	case reflectwalk.Struct:
		fallthrough

	case reflectwalk.Map:
		w.closeBlock()

	case reflectwalk.StructField:
		w.fieldPop()
	}

	// log.Debugf("exiting %s", l)
	return nil
}

// Struct is called every time reflectwalk enters a Struct
func (w *ObjectWalker) Struct(v reflect.Value) error {
	if !v.CanInterface() {
		// log.Debugf("skipping Struct [field: %s, type: %s] - CanInterface() = false", w.field().Name, v.Type())
		return nil
	}

	ty := reflect.TypeOf(v.Interface())

	if w.isTopLevel {
		// Create the top level HCL block
		// e.g.
		//   resource "kubernetes_pod" "name" { }
		topLevelBlock := hclwrite.NewBlock("resource", []string{w.ResourceType(), w.ResourceName()})
		w.openBlock(w.ResourceType(), tfkschema.TypeMeta(w.RuntimeObject).Kind, topLevelBlock)
		w.isTopLevel = false

	} else {
		// this struct will be a sub-block
		// create a new HCL block and add to parent
		field := w.field()

		if w.sliceElemType() == ty || w.sliceType() == ty {
			// When iterating over a slice of complex types, each HCL block name is based on the
			// StructField metadata of the containing Slice instead of the StructField of each Slice element.
			// Update field, so when we create the HCL block below it uses the Slice StructField
			field = w.currentSlice()
		}

		// generate a block name
		blockName := tfkschema.ToTerraformSubBlockName(field, w.currentBlock.FullSchemaName())
		// log.Debugf("creating block [%s] for field [%s]", blockName, field.Name)
		b := w.openBlock(blockName, field.Name, hclwrite.NewBlock(blockName, nil))

		// Skip some Kubernetes complex types that should be treated as primitives.
		// Do this after opening the block above because reflectwalk will
		// still call Exit for this struct and we need the calls to closeBlock() to marry up
		// TODO: figure out a uniform way to handle these cases
		switch v.Interface().(type) {
		case resource.Quantity:
			return reflectwalk.SkipEntry
		case intstr.IntOrString:
			ios := v.Interface().(intstr.IntOrString)
			if ios.IntVal > 0 || ios.StrVal != "" {
				b.hasValue = false
				b.parent.SetAttributeValue(blockName, w.convertCtyValue(v.Interface()))
				b.parent.hasValue = true
			}
			return reflectwalk.SkipEntry
		}

		// flag inlined
		b.inlined = IsInlineStruct(field)

		// check if block is supported by Terraform
		b.unsupported = !tfkschema.IsAttributeSupported(b.FullSchemaName())
	}

	return nil
}

// StructField is called by reflectwalk whenever it enters a field of a struct.
// We ignore Invalid fields, or some API fields we don't need to convert to HCL.
// The rest are added to the StuctField stack so we have access to the
// StructField data in other funcs.
func (w *ObjectWalker) StructField(field reflect.StructField, v reflect.Value) error {
	if !v.IsValid() {
		// log.Debugf(fmt.Sprint("skipping invalid ", field.Name))
		return reflectwalk.SkipEntry

	} else if ignoredField(field.Name) {
		// log.Debugf(fmt.Sprint("ignoring ", field.Name))
		return reflectwalk.SkipEntry

	} else {
		w.fieldPush(&field)

	}
	return nil
}

// Primitive is called whenever reflectwalk visits a Primitive value.
// If it's not a zero value, add an Attribute to the current HCL Block.
func (w *ObjectWalker) Primitive(v reflect.Value) error {
	if !w.ignoreSliceElems && v.CanAddr() && v.CanInterface() {
		// log.Debugf(fmt.Sprintf("Primitive: %s = %v (%T)", w.field().Name, v.Interface(), v.Interface()))

		if !IsZero(v) || tfkschema.IncludedOnZero(w.field().Name) {
			w.currentBlock.hasValue = true
			w.currentBlock.SetAttributeValue(
				tfkschema.ToTerraformAttributeName(w.field(), w.currentBlock.FullSchemaName()),
				w.convertCtyValue(v.Interface()),
			)
		}
	}
	return nil
}

// Map is called everytime reflectwalk enters a Map
// Golang maps are usually output as HCL maps, but sometimes as sub-blocks.
func (w *ObjectWalker) Map(m reflect.Value) error {
	blockName := tfkschema.ToTerraformSubBlockName(w.field(), w.currentBlock.FullSchemaName())
	hcl := hclwrite.NewBlock(blockName, nil)
	b := w.openBlock(blockName, w.field().Name, hcl)

	// If this field is also typed as Map in the Terraform schema, flag the block appropriately.
	// This will impact whether the block is rendered as a map or HCL sub-block.
	schemaElem := tfkschema.ResourceField(w.currentBlock.FullSchemaName())
	if schemaElem != nil && schemaElem.Type == schema.TypeMap {
		b.isMap = true
		b.hclMap = map[string]cty.Value{}
	}

	return nil
}

// MapElem is called every time reflectwalk enters a Map element
//
//	normalize the element key, and write element value to the HCL block as an attribute value
func (w *ObjectWalker) MapElem(m, k, v reflect.Value) error {
	// log.Debugf(fmt.Sprintf("MapElem: %s = %v (%T)", k, v.Interface(), v.Interface()))

	if !IsZero(v) {
		w.currentBlock.hasValue = true
		w.currentBlock.SetAttributeValue(
			k.String(),
			w.convertCtyValue(v.Interface()),
		)
	}

	return nil
}

/*
Slice implements reflectwalk.SliceWalker interface, and is called each time reflectwalk enters a Slice
Golang slices need to be converted to HCL in one of two ways:

*1 - a simple list of primitive values:

	list_name = ["foo", "bar", "baz"]

*2 - a list of complex objects that will be rendered as repeating HCL blocks

	container {
		name  = "blah"
		image = "nginx"
	}

	container {
		name  = "foo"
		image = "sidecar"
	}

For the second case, each time we process a SliceElem we need to use the StructField data of the Slice itself, and not the slice elem.
*/
func (w *ObjectWalker) Slice(v reflect.Value) error {
	w.slicePush(w.field())
	if !v.IsValid() {
		// log.Debugf("skipping invalid slice ")
		w.ignoreSliceElems = true

	} else if IsZero(v) {
		// log.Debugf("skipping empty slice ")
		w.ignoreSliceElems = true

	} else {
		// determine type of slice elements
		numEntries := v.Len()
		var vt reflect.Type
		if numEntries > 0 {
			w.currentBlock.hasValue = true
			vt = v.Index(0).Type()
		}

		switch {
		case vt.Kind() == reflect.Struct:
			fallthrough
		case vt.Kind() == reflect.Ptr:
			// Slice of complex types
			w.sliceElemTypePush(vt)

		default:
			// Slice of primitives
			valTy, err := gocty.ImpliedType(v.Interface())
			if err != nil {
				return fmt.Errorf("cannot encode %T as HCL expression: %w", valTy, err)
			}

			val, err := gocty.ToCtyValue(v.Interface(), valTy)
			if err != nil {
				// This should never happen, since we should always be able
				// to decode into the implied type.
				return fmt.Errorf("failed cty encode: %w", err)
			}

			// primitive type
			w.currentBlock.hasValue = true
			w.currentBlock.hcl.Body().SetAttributeValue(
				tfkschema.ToTerraformAttributeName(w.field(), w.currentBlock.FullSchemaName()),
				val,
			)

			// hint to other funcs that we don't need to walk through all Slice Elements because the
			// primitive values have already been rendered
			w.ignoreSliceElems = true
		}

	}

	return nil
}

// SliceElem implements reflectwalk.SliceWalker interface
func (w *ObjectWalker) SliceElem(i int, v reflect.Value) error {
	// log.Debugf("Elem %d: %T", i, v.Interface())
	return nil
}

// convertCtyValue takes an interface and converts to HCL types
func (w *ObjectWalker) convertCtyValue(val interface{}) cty.Value {
	// log.Debugf("processing %s (%T)", w.field().Name, val)
	switch val.(type) {
	case string:
		return cty.StringVal(val.(string))
	case bool:
		return cty.BoolVal(val.(bool))
	case int:
		return cty.NumberIntVal(int64(val.(int)))
	case int32:
		// On volume source blocks, the mode and default_mode attributes are now mandatorily a string representation of an octal value with a leading zero
		if w.currentSlice() != nil && w.currentSlice().Name == "Volumes" && (w.field().Name == "DefaultMode" || w.field().Name == "Mode") {
			str := "0" + strconv.FormatInt(int64(val.(int32)), 8)
			log.Debugf("converting %s from decimal int '%d' to octal string '%s'", w.field().Name, val.(int32), str)
			return cty.StringVal(str)
		}

		return cty.NumberIntVal(int64(val.(int32)))
	case *int32:
		val = *val.(*int32)
		return cty.NumberIntVal(int64(val.(int32)))
	case int64:
		return cty.NumberIntVal(int64(val.(int64)))
	case float64:
		return cty.NumberFloatVal(float64(val.(float64)))
	case map[string]interface{}:
		ctyMap := map[string]cty.Value{}
		for k, v := range val.(map[string]interface{}) {
			ctyMap[k] = w.convertCtyValue(v)
		}

		return cty.ObjectVal(ctyMap)
	case resource.Quantity:
		qty := val.(resource.Quantity)
		qtyPtr := &qty
		return cty.StringVal(qtyPtr.String())

	case intstr.IntOrString:
		ios := val.(intstr.IntOrString)
		iosPtr := &ios
		return cty.StringVal(iosPtr.String())

	case v1.Protocol:
		return cty.StringVal(string(val.(v1.Protocol)))

	default:
		if s, ok := val.(fmt.Stringer); ok {
			return cty.StringVal(s.String())
		}
		log.Debugf("unhandled variable type: %T", val)

		// last resort
		return cty.StringVal(fmt.Sprintf("%s", val))
	}
}

var ignoredFields = []string{
	"CreationTimestamp",
	"DeletionTimestamp",
	"Generation",
	"OwnerReferences",
	"ResourceVersion",
	"SelfLink",
	"TypeMeta",
	"Status",
	"UID",
}
var ignoredFieldMap map[string]bool

func init() {
	ignoredFieldMap = make(map[string]bool, len(ignoredFields))
	for _, v := range ignoredFields {
		ignoredFieldMap[v] = true
	}
}

func ignoredField(name string) bool {
	_, ok := ignoredFieldMap[name]
	return ok
}

// ResourceName returns the Terraform Resource name for the Kubernetes Object
func (w *ObjectWalker) ResourceName() string {
	if w.resourceName == "" {
		w.resourceName = tfkschema.ToTerraformResourceName(w.RuntimeObject)
	}

	return w.resourceName
}

// ResourceType returns the Terraform Resource type for the Kubernetes Object
func (w *ObjectWalker) ResourceType() string {
	if w.resourceType == "" {
		w.resourceType = tfkschema.ToTerraformResourceType(w.RuntimeObject)
	}

	return w.resourceType
}

// IsZero reports whether is considered the zero / empty / unset value fo the type
func IsZero(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Array, reflect.Map, reflect.Slice:
		return v.IsNil() || v.Len() == 0
	case reflect.String:
		return v.Len() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Interface, reflect.Ptr:
		return v.IsNil()
	}
	return false
}

// IsInlineStruct looks at the json tag of the given StructField, to determine
// if it has been marked as "inline"
// e.g. someField string `json:",inline"`
func IsInlineStruct(field *reflect.StructField) bool {
	jsonTag := field.Tag.Get("json")
	if jsonTag == "" {
		fmt.Printf("WARNING - field [%s] has no json tag value", field.Name)
		return false
	}

	comma := strings.Index(jsonTag, ",")
	if comma == -1 {
		return false
	}

	tagParts := strings.Split(jsonTag, ",")
	for _, part := range tagParts {
		if part == "inline" {
			return true
		}
	}

	return false
}
