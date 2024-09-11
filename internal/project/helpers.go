package project

import (
	"context"
	"fmt"
	"os"
	"reflect"
	"regexp"
	"strings"

	"github.com/Masterminds/semver"
	"github.com/apex/log"
	"github.com/google/go-github/github"

	"github.com/olekukonko/tablewriter"
	"github.com/shalb/cluster.dev/internal/config"
	"github.com/shalb/cluster.dev/pkg/colors"
	"github.com/shalb/cluster.dev/pkg/utils"
)

type UnitOperation uint16

const (
	Apply UnitOperation = iota + 1
	Destroy
	Update
	NotChanged
)

func (u UnitOperation) String() string {
	mapperStatus := map[uint16]string{
		1: colors.Fmt(colors.Green).Sprint("Apply"),
		2: colors.Fmt(colors.Red).Sprint("Destroy"),
		3: colors.Fmt(colors.Yellow).Sprint("Update"),
		4: colors.Fmt(colors.White).Sprint("NotChanged"),
	}
	return mapperStatus[uint16(u)]
}

func (u UnitOperation) HasChanges() bool {
	return u != NotChanged
}

type UnitPlanningStatus struct {
	UnitPtr   Unit
	Diff      string
	Operation UnitOperation
	IsTainted bool
	Index     int
}

type ProjectPlanningStatus struct {
	units []*UnitPlanningStatus
}

func (s *ProjectPlanningStatus) BuildGraph() (*graph, error) {
	graphRet := graph{}
	err := graphRet.BuildNew(s, config.Global.MaxParallel)
	return &graphRet, err
}

// FindUnit searching unit by pointer, return *UnitPlanningStatus only for same unit
func (s *ProjectPlanningStatus) FindUnit(unit Unit) *UnitPlanningStatus {
	if unit == nil {
		return nil
	}
	for _, us := range s.units {
		if us.UnitPtr == unit {
			return us
		}
	}
	return nil
}

// FindUnit searching unit by pointer, return *UnitPlanningStatus if unit with same key
// exists (possible to have 2 different units ptr with same key - project/projectState)
func (s *ProjectPlanningStatus) FindUnitByKey(unit Unit) *UnitPlanningStatus {
	if unit == nil {
		return nil
	}
	for _, us := range s.units {
		if us.UnitPtr.Key() == unit.Key() {
			return us
		}
	}
	return nil
}

func (s *ProjectPlanningStatus) OperationFilter(ops ...UnitOperation) *ProjectPlanningStatus {
	res := ProjectPlanningStatus{
		units: make([]*UnitPlanningStatus, 0),
	}
	if len(ops) == 0 {
		return &res
	}
	for _, uo := range s.units {
		for _, op := range ops {
			if uo.Operation == op {
				res.units = append(res.units, uo)
			}
		}
	}
	return &res
}

func (s *ProjectPlanningStatus) Add(u Unit, op UnitOperation, diff string, tainted bool) {
	uo := UnitPlanningStatus{
		UnitPtr:   u,
		Operation: op,
		Diff:      diff,
		Index:     -1,
		IsTainted: tainted,
	}
	s.units = append(s.units, &uo)
}

func (s *ProjectPlanningStatus) AddOrUpdate(u Unit, op UnitOperation, diff string, tainted bool) {
	uo := UnitPlanningStatus{
		UnitPtr:   u,
		Operation: op,
		Diff:      diff,
		Index:     -1,
		IsTainted: tainted,
	}
	existingUnit := s.FindUnit(u)
	if existingUnit == nil {
		s.units = append(s.units, &uo)
	} else {
		existingUnit.Diff = diff
		existingUnit.Operation = op
	}
}

func DiscoverCdevLastRelease() error {
	var (
		client         = github.NewClient(nil)
		ctx            = context.Background()
		org     string = "shalb"
		project string = "cluster.dev"
	)

	latestRelease, _, err := client.Repositories.GetLatestRelease(ctx, org, project)
	if err != nil {
		return err
	}
	curVersion, err := semver.NewVersion(config.Version)
	if err != nil {
		return fmt.Errorf("check failed: %v, current version: %v", err, config.Global.Version)
	}
	reqVerConstraints, err := semver.NewConstraint(*latestRelease.TagName)
	if err != nil {
		return fmt.Errorf("check failed: %v, latest stable release: %v", err, *latestRelease.TagName)
	}
	ok, _ := reqVerConstraints.Validate(curVersion)
	if !ok {
		return fmt.Errorf("the new cdev version is available. Current version: '%v', latest stable release: '%v'. Visit https://docs.cluster.dev/installation-upgrade/ to upgrade", curVersion, *latestRelease.TagName)
	}
	return nil
}

func (s *ProjectPlanningStatus) AddIfNotExists(u Unit, op UnitOperation, diff string, tainted bool) {
	existingUnit := s.FindUnitByKey(u)
	if existingUnit == nil {
		uo := UnitPlanningStatus{
			UnitPtr:   u,
			Operation: op,
			Diff:      diff,
			IsTainted: tainted,
			Index:     -1,
		}
		log.Errorf("ProjectPlanningStatus AddOrUpdate: %v", u.Key())
		s.units = append(s.units, &uo)
	}
}

func (s *ProjectPlanningStatus) HasChanges() bool {
	for _, un := range s.units {
		if un.Operation != NotChanged {
			return true
		}
	}
	return false
}

func (s *ProjectPlanningStatus) Len() int {
	return len(s.units)
}

func (s *ProjectPlanningStatus) Print() {
	for _, unitStatus := range s.units {
		fmt.Printf("UnitName: %v, Unit status: %v\n", unitStatus.UnitPtr.Key(), unitStatus.Operation.String())
	}
}

func (s *ProjectPlanningStatus) Slice() []*UnitPlanningStatus {
	return s.units
}

// CreateMarker generate hash string for template markers.
func CreateMarker(link ULinkT) (string, error) {
	if link.LinkType == "" {
		return "", fmt.Errorf("internal error: create unit link: LinkType field is empty")
	}
	if link.TargetStackName == "" {
		return "", fmt.Errorf("internal error: create unit link: StackName field is empty")
	}
	if link.TargetUnitName == "" {
		return "", fmt.Errorf("internal error: create unit link: UnitName field is empty")
	}

	var markerPath string
	if link.OutputName == "" {
		markerPath = fmt.Sprintf("%v.%v.%v", link.LinkType, link.TargetStackName, link.TargetUnitName)
	} else {
		markerPath = fmt.Sprintf("%v.%v.%v.%v", link.LinkType, link.TargetStackName, link.TargetUnitName, link.OutputName)
	}
	hash := utils.Md5(markerPath)

	return EscapeForMarkerStr(fmt.Sprintf("%s.%s.%s", hash, markerPath, hash))
}

// EscapeForMarkerStr convert URL to string which can be used as marker.
func EscapeForMarkerStr(in string) (string, error) {
	reg, err := regexp.Compile(`[^A-Za-z0-9_\-\.]+`)
	if err != nil {
		return "", err
	}
	newStr := reg.ReplaceAllString(in, "_")
	return newStr, nil
}

func printVersion() string {
	return config.Global.Version
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
					return fmt.Errorf("ScanMarkers: type conversion error")
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
				return fmt.Errorf("internal error: can't set interface field")
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
			return fmt.Errorf("internal error: can't set string field. %v", out)
		}
		out.Set(val)
	default:
		// log.Debugf("Unknown type: %v", out.Type())
	}
	return nil
}

func ConvertToTfVarName(name string) string {
	reg, _ := regexp.Compile("[^a-zA-Z0-9]+")
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

func showPlanResults(opStatus *graph) error {
	fmt.Println(colors.Fmt(colors.WhiteBold).Sprint("Plan results:"))

	if opStatus.Len() == 0 {
		fmt.Println(colors.Fmt(colors.WhiteBold).Sprint("No changes, nothing to do."))
		return nil
	}
	table := tablewriter.NewWriter(os.Stdout)

	headers := []string{}
	unitsTable := []string{}

	var deployString, updateString, destroyString, unchangedString string
	for _, unit := range opStatus.IndexedSlice() {
		log.Infof(colors.Fmt(colors.LightWhiteBold).Sprintf("Planning unit '%v':", unit.UnitPtr.Key()))
		switch unit.Operation {
		case Apply:
			fmt.Printf("%v\n", unit.Diff)
			if len(deployString) != 0 {
				deployString += "\n"
			}
			deployString += RenderUnitPlanningString(unit)
		case Update:
			fmt.Printf("%v\n", unit.Diff)
			if len(updateString) != 0 {
				updateString += "\n"
			}
			updateString += RenderUnitPlanningString(unit)
		case Destroy:
			fmt.Printf("%v\n", unit.Diff)
			if len(destroyString) != 0 {
				destroyString += "\n"
			}
			destroyString += RenderUnitPlanningString(unit)
		case NotChanged:
			log.Infof(colors.Fmt(colors.GreenBold).Sprint("Not changed."))
			if len(unchangedString) != 0 {
				unchangedString += "\n"
			}
			unchangedString += RenderUnitPlanningString(unit)
		}
	}

	if opStatus.planningUnits.OperationFilter(Apply).Len() > 0 {
		headers = append(headers, "Will be deployed")
		unitsTable = append(unitsTable, deployString)
	}
	if opStatus.planningUnits.OperationFilter(Update).Len() > 0 {
		headers = append(headers, "Will be updated")
		unitsTable = append(unitsTable, updateString)
	}
	if opStatus.planningUnits.OperationFilter(Destroy).Len() > 0 {
		headers = append(headers, "Will be destroyed")
		unitsTable = append(unitsTable, destroyString)
	}
	if opStatus.planningUnits.OperationFilter(NotChanged).Len() > 0 {
		headers = append(headers, "Unchanged")
		unitsTable = append(unitsTable, unchangedString)
	}
	table.SetHeader(headers)
	table.Append(unitsTable)
	table.Render()
	return nil
}

func RenderUnitPlanningString(uStatus *UnitPlanningStatus) string {
	keyForRender := uStatus.UnitPtr.Key()
	if config.Global.LogLevel == "debug" {
		keyForRender += fmt.Sprintf("(%v)", uStatus.Index)
	}
	switch uStatus.Operation {
	case Update:
		if uStatus.IsTainted {
			return colors.Fmt(colors.Orange).Sprintf("%s(tainted)", keyForRender)
		} else {
			return colors.Fmt(colors.Yellow).Sprint(keyForRender)
		}
	case Apply:
		if uStatus.IsTainted {
			return colors.Fmt(colors.Green).Sprintf("%s(tainted)", keyForRender)
		} else {
			return colors.Fmt(colors.Green).Sprint(keyForRender)
		}
	case Destroy:
		if uStatus.IsTainted {
			return colors.Fmt(colors.Red).Sprintf("%s(tainted)", keyForRender)
		} else {
			return colors.Fmt(colors.Red).Sprint(keyForRender)
		}
	case NotChanged:
		return colors.Fmt(colors.White).Sprint(keyForRender)
	}
	// Impossible, crush
	log.Fatalf("Unexpected internal error. Unknown unit status '%v'", uStatus.Operation.String())
	return ""
}

func DependenciesRecursiveIterate(u Unit, f func(Unit) error) error {
	return dependenciesRecursiveIterateDepth(u, f, 0)
}

func dependenciesRecursiveIterateDepth(u Unit, f func(Unit) error, depth int) error {
	if depth > 20 {
		log.Fatalf("Internal error: may be unexpected dependencies loop")
	}
	for _, dep := range u.Dependencies().Slice() {
		err := f(dep.Unit)
		if err != nil {
			return err
		}
		dependenciesRecursiveIterateDepth(dep.Unit, f, depth+1)
	}
	return nil
}

func checkUnitPath(u Unit, path string) bool {
	pathSpl := strings.Split(path, ".")
	if len(pathSpl) == 1 { // path is only stack name, so it include all units of stack
		return u.Stack().Name == pathSpl[0]
	}
	if len(pathSpl) == 2 {
		return (u.Stack().Name == pathSpl[0] && u.Name() == pathSpl[1])
	}

	return false
}

func IsUnitExcludedByTarget(u Unit) bool {
	if len(config.Global.Targets) == 0 && len(config.Global.TargetsExclude) == 0 {
		return false
	}
	// Target is set
	if len(config.Global.Targets) > 0 {
		for _, uPath := range config.Global.Targets {
			if checkUnitPath(u, uPath) {
				return false
			}
		}
		return true
	}
	// Exclude-target is set
	for _, uPath := range config.Global.TargetsExclude {
		if checkUnitPath(u, uPath) {
			return true
		}
	}
	return false
}
