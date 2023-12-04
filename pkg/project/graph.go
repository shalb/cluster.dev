package project

import (
	"fmt"
	"sync"

	"github.com/apex/log"
	"github.com/shalb/cluster.dev/pkg/config"
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
	units        *ExecSet
	mux          sync.Mutex
	waitUnitDone chan Unit
	maxParallel  int
	reverse      bool
	// sigTrap      chan os.Signal
	// stopChan     chan struct{}
}

func (g *graph) BuildDirect(planningStatus *ProjectPlanningStatus, maxParallel int) error {
	return g.build(planningStatus, maxParallel, false)
}

func (g *graph) BuildReverse(planningStatus *ProjectPlanningStatus, maxParallel int) error {
	return g.build(planningStatus, maxParallel, true)
}

func (g *graph) build(planningStatus *ProjectPlanningStatus, maxParallel int, reverse bool) error {
	g.units = NewExecSet(planningStatus)
	g.reverse = reverse
	g.maxParallel = maxParallel
	g.waitUnitDone = make(chan Unit)
	return g.checkAndBuildIndexes()
	// g.listenHupSig()
}

func (g *graph) GetNextAsync() (Unit, func(error), error) {
	g.mux.Lock()
	defer g.mux.Unlock()
	for {
		g.updateQueue()
		if config.Interrupted {
			return nil, nil, fmt.Errorf("interrupted")
		}
		readyFroExecList := g.units.StatusFilter(ReadyForExec)
		if readyFroExecList.Len() > 0 && g.units.StatusFilter(InProgress).Len() < g.maxParallel {
			unitForExec := readyFroExecList.Front()
			finFunc := func(err error) {
				g.waitUnitDone <- unitForExec.UnitPtr
			}
			unitForExec.UnitPtr.SetExecStatus(InProgress)
			g.updateQueue()
			return unitForExec.UnitPtr, finFunc, nil
		}
		if g.units.StatusFilter(Backlog, InProgress, ReadyForExec).IsEmpty() {
			return nil, nil, nil
		}
		unitFinished := <-g.waitUnitDone
		unitFinished.SetExecStatus(Finished)
		g.updateQueue()
		if unitFinished.ExecError() != nil {
			return unitFinished, nil, fmt.Errorf("error while unit running")
		}
	}
}

func (g *graph) checkAndBuildIndexes() error {
	i := 0
	for {
		readyCount := g.updateQueue()
		if readyCount == 0 {
			if g.units.StatusFilter(Backlog).Len() > 0 {
				return fmt.Errorf("the graph is broken, can't resolve sequence")
			}
			break
		}
		for _, u := range g.units.StatusFilter(ReadyForExec).Slice() {
			u.Index = i
			u.UnitPtr.SetExecStatus(Finished)
		}
		i++
	}
	g.resetUnitsStatus() // back all units to backlog
	return nil
}

func (g *graph) Slice() []*UnitPlanningStatus {
	i := 0
	res := []*UnitPlanningStatus{}
	for {
		readyCount := g.updateQueue()
		if readyCount == 0 {
			if g.units.StatusFilter(Backlog).Len() > 0 {
				return nil
			}
			break
		}
		for _, u := range g.units.StatusFilter(ReadyForExec).Slice() {
			res = append(res, u)
		}
		i++
	}
	g.resetUnitsStatus() // back all units to backlog
	return res
}

func (g *graph) resetUnitsStatus() {
	for _, u := range g.units.Slice() {
		u.UnitPtr.SetExecStatus(Backlog)
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

func (g *graph) updateQueue() int {
	if g.reverse {
		return g.updateReverseQueue()
	}
	return g.updateDirectQueue()
}

func (g *graph) updateDirectQueue() int {
	count := 0
	for _, unit := range g.units.StatusFilter(Backlog).Slice() {
		blockedByDep := false
		for _, dep := range unit.UnitPtr.Dependencies().Slice() {
			if g.units.StatusFilter(Backlog, InProgress, ReadyForExec).Find(dep.Unit) != nil {
				blockedByDep = true
				break
			}
		}
		if !blockedByDep {
			unit.UnitPtr.SetExecStatus(ReadyForExec)
			count++
		}
	}
	return count
}

func (g *graph) updateReverseQueue() int {
	count := 0
	for _, unit := range g.units.StatusFilter(Backlog).Slice() {
		// for _, dep := range unit.Dependencies().Slice() {
		graphDepFind := g.units.StatusFilter(Backlog, InProgress, ReadyForExec).Find(unit.UnitPtr)
		if graphDepFind == nil || graphDepFind.UnitPtr.GetExecStatus() == Finished {
			continue
		}
		unit.UnitPtr.SetExecStatus(ReadyForExec)
		count++
		// }
	}
	return count
}

func (g *graph) Wait() {
	for {
		if g.units.StatusFilter(InProgress).Len() == 0 {
			return
		}
		doneUnit := <-g.waitUnitDone
		doneUnit.SetExecStatus(Finished)
	}
}

// func (g *graph) GetSequenceSet() []Unit {
// 	mCount := len(g.units.Slice())
// 	res := make([]Unit, mCount)
// 	for i := 0; i < mCount; i++ {
// 		md := g.GetNextSync()
// 		if md == nil {
// 			log.Fatal("Building apply units set: getting nil unit, undefined behavior")
// 		}
// 		res[i] = md
// 		log.Infof("GetSequenceSet %v %v", i, md.Key())
// 	}
// 	return res
// }

// func (g *grapherNew) listenHupSig() {
// 	signals := []os.Signal{syscall.SIGTERM, syscall.SIGINT}
// 	g.sigTrap = make(chan os.Signal, 1)
// 	signal.Notify(g.sigTrap, signals...)
// 	// log.Warn("Listening signals...")
// 	go func() {
// 		for {
// 			select {
// 			case <-g.sigTrap:
// 				config.Interrupted = true
// 			case <-g.stopChan:
// 				// log.Warn("Stop listening")
// 				signal.Stop(g.sigTrap)
// 				g.sigTrap <- nil
// 				close(g.sigTrap)
// 				return
// 			}
// 		}
// 	}()
// }

// func (g *grapherNew) Close() error {
// 	g.stopChan <- struct{}{}
// 	return nil
// }
