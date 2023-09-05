package project

import (
	"fmt"
	"sync"

	"github.com/apex/log"
)

// ULinkT describe unit link betwen one target unit and multiple cli units, which uses this unit (output or remote state, or custom unit dependency).
type ULinkT struct {
	Unit            Unit        `json:"-"`
	LinkType        string      `json:"link_type"`
	TargetUnitName  string      `json:"target_unit_name"`
	TargetStackName string      `json:"target_stack_name"`
	OutputName      string      `json:"output_name"`
	OutputData      interface{} `json:"output_data"`
}

// UnitLinksT describe a set of links (dependencies) betwen units inside project.
type UnitLinksT struct {
	LinksList map[string]*ULinkT `json:"unit_links_list,omitempty"`
	MapMutex  sync.RWMutex
}

func (u *ULinkT) UnitKey() (res string) {
	if u.TargetStackName == "" || u.TargetUnitName == "" {
		return
	}
	res = fmt.Sprintf("%v.%v", u.TargetStackName, u.TargetUnitName)
	return
}

func (u *ULinkT) InitUnitPtr(p *Project) (err error) {
	if u.TargetStackName == "" || u.TargetUnitName == "" {
		return fmt.Errorf("stack name or unit name is empty")
	}
	modKey := fmt.Sprintf("%s.%s", u.TargetStackName, u.TargetUnitName)
	depUnit, exists := p.Units[modKey]
	if !exists {
		return fmt.Errorf("link unit does not exists '%s'", modKey)
	}
	u.Unit = depUnit
	return
}

func (u *ULinkT) LinkPath() (res string) {
	if u.TargetStackName == "" || u.TargetUnitName == "" || u.LinkType == "" {
		return
	}
	if u.OutputName == "" {
		res = fmt.Sprintf("%v.%v.%v", u.LinkType, u.TargetStackName, u.TargetUnitName)
	} else {
		res = fmt.Sprintf("%v.%v.%v.%v", u.LinkType, u.TargetStackName, u.TargetUnitName, u.OutputName)
	}
	return
}

func (o *UnitLinksT) Set(l *ULinkT) (string, error) {
	o.MapMutex.Lock()
	defer o.MapMutex.Unlock()
	if o.LinksList == nil {
		o.LinksList = make(map[string]*ULinkT)
	}
	key, err := CreateMarker(*l)
	if err != nil {
		return "", err
	}
	o.LinksList[key] = l
	return key, err
}

// Insert insert element to map, override if exists
func (o *UnitLinksT) Insert(key string, l *ULinkT) {
	o.MapMutex.Lock()
	defer o.MapMutex.Unlock()
	if o.LinksList == nil {
		o.LinksList = make(map[string]*ULinkT)
	}
	o.LinksList[key] = l
}

// Insert insert element to map, return error if exists
func (o *UnitLinksT) InsertTry(key string, l *ULinkT) error {
	o.MapMutex.Lock()
	defer o.MapMutex.Unlock()
	if o.LinksList == nil {
		o.LinksList = make(map[string]*ULinkT)
	}
	_, exists := o.LinksList[key]
	if exists {
		return fmt.Errorf("add unit link to map: key '%s' already exists", key)
	}
	o.LinksList[key] = l
	return nil
}

func (o *UnitLinksT) Join(l *UnitLinksT) error {
	if l.IsEmpty() {
		return nil
	}
	for _, link := range l.Map() {
		_, err := o.Set(link)
		if err != nil {
			return err
		}
	}
	return nil
}

// JoinWithDataReplace join source links into o. If link exists - only copy output data.
func (o *UnitLinksT) JoinWithDataReplace(source *UnitLinksT) error {
	if source.IsEmpty() {
		return nil
	}
	for key, link := range source.Map() {
		targetLink := o.Get(key)
		if targetLink != nil {
			if targetLink.OutputData == nil {
				targetLink.OutputData = link.OutputData
			}
		} else {
			_, err := o.Set(link)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
func (o *UnitLinksT) Get(key string) (res *ULinkT) {
	o.MapMutex.RLock()
	defer o.MapMutex.RUnlock()
	if o.LinksList == nil {
		return nil
	}
	res = o.LinksList[key]
	return
}

func (o *UnitLinksT) Delete(marker string) (err error) {
	o.MapMutex.Lock()
	defer o.MapMutex.Unlock()
	if o.LinksList == nil {
		return nil
	}
	_, exists := o.LinksList[marker]
	if exists {
		delete(o.LinksList, marker)
	}
	return
}

func (o *UnitLinksT) Slice() (res []*ULinkT) {
	o.MapMutex.RLock()
	defer o.MapMutex.RUnlock()
	if o.LinksList == nil {
		return nil
	}
	res = make([]*ULinkT, len(o.LinksList))
	i := 0
	for _, el := range o.LinksList {
		res[i] = el
		i++
	}
	return
}

func (o *UnitLinksT) Map() (res map[string]*ULinkT) {
	o.MapMutex.RLock()
	defer o.MapMutex.RUnlock()
	res = make(map[string]*ULinkT)
	for k, v := range o.LinksList {
		res[k] = v
	}
	return res
}

// ByLinkTypes returns sublist with unit link types == any of outputType slice. Returns full list if outputType is empty.
func (o *UnitLinksT) ByLinkTypes(outputType ...string) (res *UnitLinksT) {
	res = &UnitLinksT{
		LinksList: make(map[string]*ULinkT),
	}
	if o.IsEmpty() {
		return
	}
	for key, link := range o.Map() {
		if len(outputType) == 0 {
			res.Insert(key, link)
			continue
		}
		for _, tp := range outputType {
			if tp == link.LinkType {
				res.Insert(key, link)
				break
			}
		}
	}
	return
}

// UniqUnits return list of uniq links units.
func (o *UnitLinksT) UniqUnits() map[string]Unit {
	res := make(map[string]Unit)
	if o.IsEmpty() {
		return nil
	}
	for _, el := range o.Map() {
		unit, exists := res[el.UnitKey()]
		if !exists {
			res[el.UnitKey()] = el.Unit
			continue
		}
		if unit != nil {
			continue
		}
		if el.Unit == nil {
			log.Warnf("Dev debug. Nil unit pointer %v. Pls check.", el.UnitKey())
		}
		res[el.UnitKey()] = el.Unit
	}
	return res
}

func (o *UnitLinksT) ByTargetUnit(unit Unit) (res *UnitLinksT) {
	res = &UnitLinksT{
		LinksList: make(map[string]*ULinkT),
	}
	if o.LinksList == nil {
		return
	}
	for key, el := range o.Map() {
		if unit.Name() == el.TargetUnitName && unit.Stack().Name == el.TargetStackName {
			res.LinksList[key] = el
			res.Insert(key, el)
		}
	}
	return
}

func (o *UnitLinksT) IsEmpty() bool {
	o.MapMutex.RLock()
	defer o.MapMutex.RUnlock()
	return o.LinksList == nil || len(o.LinksList) == 0
}

func (o *UnitLinksT) Size() int {
	o.MapMutex.RLock()
	defer o.MapMutex.RUnlock()
	return len(o.LinksList)
}

func NewUnitLinksT() *UnitLinksT {
	res := &UnitLinksT{
		LinksList: make(map[string]*ULinkT),
	}
	return res
}
