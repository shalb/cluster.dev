package project

import (
	"fmt"
	"strings"
)

// import (
// 	"container/list"
// 	"fmt"
// 	"os"
// 	"os/signal"
// 	"strings"
// 	"sync"
// 	"syscall"

// 	"github.com/apex/log"
// 	"github.com/shalb/cluster.dev/pkg/config"
// )

// type unitExecuteResult struct {
// 	Mod   Unit
// 	Error error
// }

// type grapher struct {
// 	finished       map[string]unitExecuteResult
// 	inProgress     map[string]Unit
// 	queue          list.List
// 	units          map[string]Unit
// 	unFinished     map[string]Unit
// 	mux            sync.Mutex
// 	waitForModDone chan unitExecuteResult
// 	maxParallel    int
// 	hasError       bool
// 	reverse        bool
// 	unitsErrors    map[string]error
// 	sigTrap        chan os.Signal
// 	stopChan       chan struct{}
// }

// func (g *grapher) InitP(planningStatus *ProjectPlanningStatus, maxParallel int, reverse bool) {
// 	if maxParallel < 1 {
// 		log.Fatal("Internal error, parallelism < 1.")
// 	}
// 	g.units = make(map[string]Unit)
// 	g.unFinished = make(map[string]Unit)
// 	g.inProgress = make(map[string]Unit)

// 	g.unitsErrors = make(map[string]error)

// 	for _, uStatus := range planningStatus.Slice() {
// 		g.units[uStatus.UnitPtr.Key()] = uStatus.UnitPtr
// 		g.unFinished[uStatus.UnitPtr.Key()] = uStatus.UnitPtr
// 	}
// 	g.maxParallel = maxParallel
// 	g.queue.Init()

// 	g.finished = make(map[string]unitExecuteResult)
// 	g.waitForModDone = make(chan unitExecuteResult)
// 	g.hasError = false
// 	g.reverse = reverse
// 	g.updateQueue()
// 	g.stopChan = make(chan struct{})
// 	g.listenHupSig()
// }

// func (g *grapher) HasError() bool {
// 	return g.hasError
// }

// func (g *grapher) updateQueue() int {
// 	if g.reverse {
// 		return g.updateReverseQueue()
// 	}
// 	return g.updateDirectQueue()
// }

// func (g *grapher) UnitFinished(u Unit) bool {
// 	return g.unFinished[u.Key()] != nil
// }

// func (g *grapher) updateDirectQueue() int {
// 	count := 0
// 	for key, mod := range g.units {
// 		isReady := true
// 		for _, dep := range mod.Dependencies().Slice() {
// 			if g.UnitFinished(dep.Unit) {
// 				isReady = false
// 				break
// 			}
// 		}
// 		if isReady {
// 			g.queue.PushBack(mod)
// 			delete(g.units, key)
// 			count++
// 		}
// 	}
// 	return count
// }

// func (g *grapher) updateReverseQueue() int {
// 	count := 0
// 	for key, mod := range g.units {
// 		dependedMods := findDependedUnits(g.unFinished, mod)
// 		if len(dependedMods) > 0 {
// 			continue
// 		}
// 		g.queue.PushBack(mod)
// 		delete(g.units, key)
// 		count++
// 	}
// 	return count
// }

// func (g *grapher) GetNextAsync() (Unit, func(error), error) {
// 	g.mux.Lock()
// 	defer g.mux.Unlock()
// 	for {
// 		if config.Interrupted {
// 			g.queue.Init()
// 			g.units = make(map[string]Unit)
// 			g.updateQueue()
// 			return nil, nil, fmt.Errorf("interupted")
// 		}
// 		if g.queue.Len() > 0 && len(g.inProgress) < g.maxParallel {
// 			modElem := g.queue.Front()
// 			mod := modElem.Value.(Unit)
// 			finFunc := func(err error) {
// 				g.waitForModDone <- unitExecuteResult{mod, err}
// 			}
// 			g.queue.Remove(modElem)
// 			g.inProgress[mod.Key()] = mod
// 			return mod, finFunc, nil
// 		}
// 		if g.Len() == 0 {
// 			return nil, nil, nil
// 		}
// 		doneMod := <-g.waitForModDone
// 		g.setUnitDone(doneMod)
// 		if doneMod.Error != nil {
// 			return doneMod.Mod, nil, fmt.Errorf("error while unit running")
// 		}
// 	}
// }

// func (g *grapher) GetNextSync() Unit {
// 	if g.Len() == 0 {
// 		return nil
// 	}
// 	modElem := g.queue.Front()
// 	mod := modElem.Value.(Unit)
// 	g.queue.Remove(modElem)
// 	g.setUnitDone(unitExecuteResult{mod, nil})
// 	return mod
// }

// func (g *grapher) GetSequenceSet() []Unit {
// 	res := make([]Unit, g.Len())
// 	mCount := g.Len()
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

// func (g *grapher) setUnitDone(doneMod unitExecuteResult) {
// 	g.finished[doneMod.Mod.Key()] = doneMod
// 	delete(g.inProgress, doneMod.Mod.Key())
// 	delete(g.unFinished, doneMod.Mod.Key())
// 	if doneMod.Error != nil {
// 		g.unitsErrors[doneMod.Mod.Key()] = doneMod.Error
// 		g.hasError = true
// 	}
// 	g.updateQueue()
// }

// func (g *grapher) Errors() map[string]error {
// 	return g.unitsErrors
// }

// func (g *grapher) Wait() {
// 	for {
// 		if len(g.inProgress) == 0 {
// 			return
// 		}
// 		doneMod := <-g.waitForModDone
// 		g.setUnitDone(doneMod)
// 	}
// }

// func (g *grapher) Len() int {
// 	return len(g.units) + g.queue.Len() + len(g.inProgress)
// }

func checkUnitDependencies(p *Project) error {
	for _, uniit := range p.Units {
		if err := checkDependenciesRecursive(uniit); err != nil {
			return fmt.Errorf("unresolved dependency in unit %v.%v: %w", uniit.Stack().Name, uniit.Name(), err)
		}
	}
	return nil
}

func checkDependenciesRecursive(unit Unit, chain ...string) error {
	if err := checkUnitDependenciesCircle(chain); err != nil {
		return err
	}
	chain = append(chain, unit.Key())
	for _, dep := range unit.Dependencies().Slice() {
		if err := checkDependenciesRecursive(dep.Unit, chain...); err != nil {
			return err
		}
	}
	return nil
}

func checkUnitDependenciesCircle(chain []string) error {
	if len(chain) < 2 {
		return nil
	}
	circleCheck := []string{}
	for _, str := range chain {
		for _, comareStr := range circleCheck {
			// log.Warnf("Compare: %v == %v", str, )
			if str == comareStr {
				circleCheck = append(circleCheck, str)
				return fmt.Errorf("loop: %s", strings.Join(circleCheck, " -> "))
			}
		}
		circleCheck = append(circleCheck, str)
	}
	return nil
}

func findDependedUnits(modList map[string]Unit, targetMod Unit) map[string]Unit {
	res := map[string]Unit{}
	for key, mod := range modList {
		if mod.Key() == targetMod.Key() {
			continue
		}
		for _, dep := range mod.Dependencies().Slice() {
			if dep.Unit.Key() == targetMod.Key() {
				res[key] = mod
			}
		}
	}
	return res
}

// func (g *grapher) listenHupSig() {
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

// func (g *grapher) Close() error {
// 	g.stopChan <- struct{}{}
// 	return nil
// }