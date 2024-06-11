package project

import (
	"fmt"
	"sync"

	"github.com/apex/log"
	"github.com/shalb/cluster.dev/internal/config"
)

type ExecutionStatus uint16

const (
	Backlog ExecutionStatus = iota + 1
	ReadyForExec
	InProgress
	Finished
)

type ExecSet struct {
	execUnits []*UnitPlanningStatus
}

func (e *ExecSet) Find(u Unit) *UnitPlanningStatus {
	for _, eu := range e.execUnits {
		if eu.UnitPtr == u {
			return eu
		}
	}
	return nil
}

func (e *ExecSet) Index(u Unit) int {
	for i, eu := range e.execUnits {
		if eu.UnitPtr == u {
			return i
		}
	}
	return -1
}

func (e *ExecSet) AddUnit(u *UnitPlanningStatus) {
	if u == nil {
		log.Debug("Internal problem: create exec set, added unit is nil, ignore")
		return
	}
	if e.Find(u.UnitPtr) != nil {
		return
	}
	e.execUnits = append(e.execUnits, u)
}

func (e *ExecSet) Delete(u Unit) {
	index := e.Index(u)
	if index < 0 {
		return
	}
	e.execUnits = append(e.execUnits[:index], e.execUnits[index+1:]...)
}

func (e *ExecSet) Slice() []*UnitPlanningStatus {
	return e.execUnits
}

func (e *ExecSet) Len() int {
	return len(e.execUnits)
}

func (e *ExecSet) IsEmpty() bool {
	return len(e.execUnits) == 0
}

func (e *ExecSet) Front() *UnitPlanningStatus {
	if len(e.execUnits) > 0 {
		return e.execUnits[0]
	}
	return nil
}

func (e *ExecSet) StatusFilter(statusList ...ExecutionStatus) *ExecSet {
	res := ExecSet{
		execUnits: make([]*UnitPlanningStatus, 0),
	}
	for _, ue := range e.execUnits {
		for _, status := range statusList {
			if ue.UnitPtr.GetExecStatus() == status {
				res.AddUnit(ue)
				break
			}
		}
	}
	return &res
}

func NewExecSet(planningStatus *ProjectPlanningStatus) *ExecSet {
	res := ExecSet{
		execUnits: make([]*UnitPlanningStatus, 0),
	}
	for _, unit := range planningStatus.Slice() {
		res.AddUnit(unit)
		unit.UnitPtr.SetExecStatus(Backlog)
	}
	return &res
}

type graph struct {
	units         *ExecSet
	mux           sync.Mutex
	waitUnitDone  chan *UnitPlanningStatus
	maxParallel   int
	indexedSlice  []*UnitPlanningStatus
	planningUnits *ProjectPlanningStatus
	// sigTrap      chan os.Signal
	// stopChan     chan struct{}
}

func (g *graph) BuildNew(planningStatus *ProjectPlanningStatus, maxParallel int) error {
	g.units = NewExecSet(planningStatus)
	g.maxParallel = maxParallel
	g.waitUnitDone = make(chan *UnitPlanningStatus)
	g.planningUnits = planningStatus
	return g.checkAndBuildIndexes()
	// g.listenHupSig()
}

func (g *graph) GetNextAsync() (*UnitPlanningStatus, func(error), error) {
	g.mux.Lock()
	defer g.mux.Unlock()
	for {
		g.updateQueueNew()
		if config.Interrupted {
			return nil, nil, fmt.Errorf("interrupted")
		}
		readyFroExecList := g.units.StatusFilter(ReadyForExec)
		if readyFroExecList.Len() > 0 && g.units.StatusFilter(InProgress).Len() < g.maxParallel {
			unitForExec := readyFroExecList.Front()
			finFunc := func(err error) {
				g.waitUnitDone <- unitForExec
			}
			unitForExec.UnitPtr.SetExecStatus(InProgress)
			g.updateQueueNew()
			return unitForExec, finFunc, nil
		}
		if g.units.StatusFilter(Backlog, InProgress, ReadyForExec).IsEmpty() {
			return nil, nil, nil
		}
		unitFinished := <-g.waitUnitDone
		unitFinished.UnitPtr.SetExecStatus(Finished)
		g.updateQueueNew()
		if unitFinished.UnitPtr.ExecError() != nil {
			return unitFinished, nil, fmt.Errorf("error while unit running")
		}
	}
}

func (g *graph) checkAndBuildIndexes() error {
	i := 0
	g.indexedSlice = []*UnitPlanningStatus{}
	apply := []*UnitPlanningStatus{}
	notChanged := []*UnitPlanningStatus{}
	for {
		readyCount := g.updateQueueNew()
		if readyCount == 0 {
			if g.units.StatusFilter(Backlog).Len() > 0 {
				return fmt.Errorf("the graph is broken, can't resolve sequence")
			}
			break
		}
		for _, u := range g.units.StatusFilter(ReadyForExec).Slice() {
			switch u.Operation {
			case Destroy:
				// Place 'destroy' units first in queue
				g.indexedSlice = append(g.indexedSlice, u)
				u.Index = i
			case Apply, Update:
				// Then place 'apply/update' units
				apply = append(apply, u)
				u.Index = i
			default:
				// Place notChanged units to the end of queue and mark them as 'Finished' in graph
				notChanged = append(notChanged, u)
				u.Index = -1
			}
			// Mark unit as finished for next updateQueue
			u.UnitPtr.SetExecStatus(Finished)
		}
		i++
	}
	g.indexedSlice = append(g.indexedSlice, apply...)
	g.indexedSlice = append(g.indexedSlice, notChanged...)
	g.resetUnitsStatus()
	return nil
}

// IndexedSlice return the slice of units in sorted in order ready for exec.
func (g *graph) IndexedSlice() []*UnitPlanningStatus {
	return g.indexedSlice
}

func (g *graph) resetUnitsStatus() {
	for _, u := range g.units.Slice() {
		if u.Operation != NotChanged {
			u.UnitPtr.SetExecStatus(Backlog)
		}
	}
}

func (g *graph) Errors() []Unit {
	res := []Unit{}
	for _, u := range g.units.Slice() {
		if u.UnitPtr.ExecError() != nil {
			res = append(res, u.UnitPtr)
		}
	}
	return res
}

func (g *graph) Len() int {
	return g.units.StatusFilter(Backlog, InProgress, ReadyForExec).Len()
}

func (g *graph) updateQueueNew() int {
	count := 0
	for _, unit := range g.units.StatusFilter(Backlog).Slice() {
		blockedByDep := false
		switch unit.Operation {
		case Apply, Update:
			for _, dep := range unit.UnitPtr.Dependencies().Slice() {
				if g.units.StatusFilter(Backlog, InProgress, ReadyForExec).Find(dep.Unit) != nil {
					blockedByDep = true
					break
				}
			}
		case Destroy:
			for _, unitForCheck := range g.units.StatusFilter(Backlog, InProgress, ReadyForExec).Slice() {
				for _, link := range unitForCheck.UnitPtr.Dependencies().Slice() {
					if link.UnitKey() == unit.UnitPtr.Key() {
						blockedByDep = true
					}
				}
			}
		case NotChanged:
			unit.UnitPtr.SetExecStatus(Finished)
		}
		if !blockedByDep {
			unit.UnitPtr.SetExecStatus(ReadyForExec)
			count++
			continue
		}
	}
	return count
}

func (g *graph) Wait() {
	for {
		if g.units.StatusFilter(InProgress).Len() == 0 {
			return
		}
		doneUnit := <-g.waitUnitDone
		doneUnit.UnitPtr.SetExecStatus(Finished)
	}
}
