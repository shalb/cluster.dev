package project

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"

	"github.com/apex/log"
	"github.com/olekukonko/tablewriter"
	"github.com/shalb/cluster.dev/pkg/colors"
	"github.com/shalb/cluster.dev/pkg/config"
	"github.com/shalb/cluster.dev/pkg/utils"
)

// CreateMarker generate hash string for template markers.
func CreateMarker(link ULinkT) (string, error) {
	if link.LinkType == "" {
		return "", fmt.Errorf("internal error: create unit link: LinkType field is empty")
	}
	if link.TargenStackName == "" {
		return "", fmt.Errorf("internal error: create unit link: StackName field is empty")
	}
	if link.TargetUnitName == "" {
		return "", fmt.Errorf("internal error: create unit link: UnitName field is empty")
	}

	var markerPath string
	if link.OutputName == "" {
		markerPath = fmt.Sprintf("%v.%v.%v", link.LinkType, link.TargenStackName, link.TargetUnitName)
	} else {
		markerPath = fmt.Sprintf("%v.%v.%v.%v", link.LinkType, link.TargenStackName, link.TargetUnitName, link.OutputName)
	}
	hash := utils.Md5(markerPath)
	return fmt.Sprintf("%s.%s.%s", hash, markerPath, hash), nil
}

func printVersion() string {
	return config.Global.Version
}

func removeDirContent(dir string) error {
	d, err := os.Open(dir)
	if err != nil {
		return err
	}
	defer d.Close()
	names, err := d.Readdirnames(-1)
	if err != nil {
		return err
	}
	for _, name := range names {
		err = os.RemoveAll(filepath.Join(dir, name))
		if err != nil {
			return err
		}
	}
	return nil
}

func findUnit(unit Unit, modsList map[string]Unit) *Unit {
	mod, exists := modsList[fmt.Sprintf("%s.%s", unit.Stack().Name, unit.Name())]
	// log.Printf("Check Mod: %s, exists: %v, list %v", name, exists, modsList)
	if !exists {
		return nil
	}
	return &mod
}

// ScanMarkers use marker scanner function to replace templated markers.
func ScanMarkers(data interface{}, procFunc MarkerScanner, unit Unit) error {
	if data == nil {
		return nil
	}
	out := reflect.ValueOf(data)
	if out.Kind() == reflect.Ptr && !out.IsNil() {
		out = out.Elem()

		//log.Fatalf("%v \n%v ", out.Kind(), out)
	}
	switch out.Kind() {
	case reflect.Slice:
		// log.Warnf("slice %v", out)
		for i := 0; i < out.Len(); i++ {
			// log.Errorf("%v", out)
			sliceElem := out.Index(i)
			sliceElemKind := sliceElem.Kind()
			if sliceElem.Kind() == reflect.Interface || sliceElem.Kind() == reflect.Ptr {
				sliceElemKind = sliceElem.Elem().Kind()
			}
			if sliceElemKind == reflect.String {
				val, err := procFunc(sliceElem, unit)
				if err != nil {
					return err
				}
				// log.Warnf("slice set")
				out.Index(i).Set(val)
			} else {
				// log.Warnf("slice scan")
				err := ScanMarkers(out.Index(i).Interface(), procFunc, unit)
				if err != nil {
					return err
				}
			}
		}
	case reflect.Map:
		// log.Warn("map")
		for _, key := range out.MapKeys() {
			elem := out.MapIndex(key)
			if elem.Kind() != reflect.Interface && elem.Kind() != reflect.Ptr {
				val, err := procFunc(elem, unit)
				if err != nil {
					return err
				}
				if val.Kind() != elem.Kind() {
					log.Fatal("ScanMarkers: type conversion error")
				}
				out.SetMapIndex(key, val)
				continue
			}
			if elem.Elem().Kind() == reflect.String {
				val, err := procFunc(out.MapIndex(key), unit)
				if err != nil {
					return err
				}
				out.SetMapIndex(key, val)
			} else {
				err := ScanMarkers(elem.Interface(), procFunc, unit)
				if err != nil {
					return err
				}
			}
		}
	case reflect.Struct:
		// log.Warn("struct")
		for i := 0; i < out.NumField(); i++ {
			if out.Field(i).Kind() == reflect.String {
				val, err := procFunc(reflect.ValueOf(out.Field(i).Interface()), unit)
				if err != nil {
					return err
				}
				out.Field(i).Set(val)
			} else {
				err := ScanMarkers(out.Field(i).Interface(), procFunc, unit)
				if err != nil {
					return err
				}
			}
		}
	case reflect.Interface:
		// log.Warn("interface")
		if reflect.TypeOf(out.Interface()).Kind() == reflect.String {
			if !out.CanSet() {
				log.Fatal("Internal error: can't set interface field.")
			}
			val, err := procFunc(out, unit)
			if err != nil {
				return err
			}
			out.Set(val)
		} else {
			err := ScanMarkers(out.Interface(), procFunc, unit)
			if err != nil {
				return err
			}
		}
	case reflect.String:
		// log.Warn("string")
		val, err := procFunc(out, unit)
		if err != nil {
			return err
		}
		if !out.CanSet() {
			log.Fatalf("Internal error: can't set string field. %v", out)
		}
		out.Set(val)
	default:
		// log.Debugf("Unknown type: %v", out.Type())
	}
	return nil
}

func ConvertToTfVarName(name string) string {
	reg, err := regexp.Compile("[^a-zA-Z0-9]+")
	if err != nil {
		log.Fatal(err.Error())
	}
	processedString := reg.ReplaceAllString(name, "_")
	return strings.ToLower(processedString)
}

func ConvertToShellVarName(name string) string {
	return strings.ToUpper(ConvertToTfVarName(name))
}

func ConvertToShellVar(name string) string {
	return fmt.Sprintf("${%s}", ConvertToShellVarName(name))
}

func ProjectsFilesExists() bool {
	project := NewEmptyProject()
	_ = project.readManifests()
	if project.configDataFile != nil || len(project.objectsFiles) > 0 {
		return true
	}
	return false
}

func showPlanResults(deployList, updateList, destroyList, unchangedList []string) {
	fmt.Println(colors.Fmt(colors.WhiteBold).Sprint("Plan results:"))

	if len(deployList)+len(updateList)+len(destroyList) == 0 {
		fmt.Println(colors.Fmt(colors.WhiteBold).Sprint("No changes, nothing to do."))
		return
	}
	table := tablewriter.NewWriter(os.Stdout)

	headers := []string{}
	unitsTable := []string{}

	var deployString, updateString, destroyString, unchangedString string
	for i, modName := range deployList {
		if i != 0 {
			deployString += "\n"
		}
		deployString += colors.Fmt(colors.Green).Sprint(modName)
	}
	for i, modName := range updateList {
		if i != 0 {
			updateString += "\n"
		}
		updateString += colors.Fmt(colors.Yellow).Sprint(modName)
	}
	for i, modName := range destroyList {
		if i != 0 {
			destroyString += "\n"
		}
		destroyString += colors.Fmt(colors.Red).Sprint(modName)
	}
	for i, modName := range unchangedList {
		if i != 0 {
			unchangedString += "\n"
		}
		unchangedString += colors.Fmt(colors.White).Sprint(modName)
	}
	if len(deployList) > 0 {
		headers = append(headers, "Will be deployed")
		unitsTable = append(unitsTable, deployString)
	}
	if len(updateList) > 0 {
		headers = append(headers, "Will be updated")
		unitsTable = append(unitsTable, updateString)
	}
	if len(destroyList) > 0 {
		headers = append(headers, "Will be destroyed")
		unitsTable = append(unitsTable, destroyString)
	}
	if len(unchangedList) > 0 {
		headers = append(headers, "Unchanged")
		unitsTable = append(unitsTable, unchangedString)
	}
	table.SetHeader(headers)
	table.Append(unitsTable)
	table.Render()
}
