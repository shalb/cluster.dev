package project

import (
	"fmt"
)

// ULinkT describe unit link betwen one target unit and multiple cli units, which uses this unit (output or remote state, or custom unit dependency).
type ULinkT struct {
	Unit            Unit        `json:"-"`
	LinkType        string      `json:"link_type"`
	TargetUnitName  string      `json:"target_unit_name"`
	TargenStackName string      `json:"target_stack_name"`
	OutputName      string      `json:"output_name"`
	OutputData      interface{} `json:"output_data"`
}

// UnitLinksT describe a set of links (dependencies) betwen units inside project.
type UnitLinksT struct {
	List map[string]*ULinkT `json:"unit_links_list,omitempty"`
}

func (u *ULinkT) UnitKey() (res string) {
	if u.TargenStackName == "" || u.TargetUnitName == "" {
		return
	}
	res = fmt.Sprintf("%v.%v", u.TargenStackName, u.TargetUnitName)
	return
}

func (u *ULinkT) InitUnitPtr(p *Project) (err error) {
	if u.TargenStackName == "" || u.TargetUnitName == "" {
		return fmt.Errorf("stack name or unit name is empty")
	}
	modKey := fmt.Sprintf("%s.%s", u.TargenStackName, u.TargetUnitName)
	depUnit, exists := p.Units[modKey]
	if !exists {
		return fmt.Errorf("link unit does not exists '%s'", modKey)
	}
	u.Unit = depUnit
	return
}

func (u *ULinkT) LinkPath() (res string) {
	if u.TargenStackName == "" || u.TargetUnitName == "" || u.LinkType == "" {
		return
	}
	res = fmt.Sprintf("%v.%v", u.TargenStackName, u.TargetUnitName)

	if u.OutputName == "" {
		res = fmt.Sprintf("%v.%v.%v", u.LinkType, u.TargenStackName, u.TargetUnitName)
	} else {
		res = fmt.Sprintf("%v.%v.%v.%v", u.LinkType, u.TargenStackName, u.TargetUnitName, u.OutputName)
	}
	return
}

func (o *UnitLinksT) Set(l *ULinkT) (string, error) {
	if o.List == nil {
		o.List = make(map[string]*ULinkT)
	}
	key, err := CreateMarker(*l)
	if err != nil {
		return "", err
	}
	o.List[key] = l
	return key, err
}

func (o *UnitLinksT) Join(l *UnitLinksT) error {
	if l.IsEmpty() {
		return nil
	}
	for _, linkl := range l.Map() {
		_, err := o.Set(linkl)
		if err != nil {
			return err
		}
	}
	return nil
}

func (o *UnitLinksT) Get(key string) (res *ULinkT) {
	if o.List == nil {
		return nil
	}
	res, _ = o.List[key]
	return
}

func (o *UnitLinksT) Delete(marker string) (err error) {
	if o.List == nil {
		return nil
	}
	_, exists := o.List[marker]
	if exists {
		delete(o.List, marker)
	}
	return
}

func (o *UnitLinksT) Slice() (res []*ULinkT) {
	if o.List == nil {
		return nil
	}
	res = make([]*ULinkT, len(o.List))
	i := 0
	for _, el := range o.List {
		res[i] = el
		i++
	}
	return
}

func (o *UnitLinksT) Map() (res map[string]*ULinkT) {
	return o.List
}

// ByLinkTypes returns sublist with unit link types == any of outputType slice. Returns full list if outputType is empty.
func (o *UnitLinksT) ByLinkTypes(outputType ...string) (res *UnitLinksT) {
	res = &UnitLinksT{
		List: make(map[string]*ULinkT),
	}
	if o.List == nil {
		return
	}
	for key, el := range o.List {
		if len(outputType) == 0 {
			res.List[key] = el
			continue
		}
		for _, tp := range outputType {
			if tp == el.LinkType {
				res.List[key] = el
				break
			}
		}
	}
	return
}

func (o *UnitLinksT) ByTargetUnit(unit Unit) (res *UnitLinksT) {
	res = &UnitLinksT{
		List: make(map[string]*ULinkT),
	}
	if o.List == nil {
		return
	}
	// res = make(map[string]*ULinkT)\
	for key, el := range o.List {
		// log.Warnf("ByTargetUnit: %v.%v == %v.%v", unit.Stack().Name, unit.Name(), el.TargenStackName, el.TargetUnitName)
		if unit.Name() == el.TargetUnitName && unit.Stack().Name == el.TargenStackName {
			res.List[key] = el
		}
	}
	return
}

func (o *UnitLinksT) IsEmpty() bool {
	return o.List == nil || len(o.List) == 0
}

func (o *UnitLinksT) Size() int {
	return len(o.List)
}

// func (o *UnitLinksT) Init(unitList map[string]Unit) error {
// 	for _, el := range o.List {
// 		log.Warnf("UnitLinksT init %v", el.TargetUnitName)
// 		uKey := fmt.Sprintf("%v.%v", el.TargenStackName, el.TargetUnitName)
// 		unit, exists := unitList[uKey]
// 		if !exists {
// 			return fmt.Errorf("init unit link: unit '%v' does not found", uKey)
// 		}
// 		el.Unit = unit
// 	}
// 	return nil
// }
