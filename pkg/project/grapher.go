package project

import (
	"container/list"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/apex/log"
	"github.com/shalb/cluster.dev/pkg/config"
)

type modResult struct {
	Mod   Unit
	Error error
}

type grapher struct {
	finished       map[string]modResult
	inProgress     map[string]Unit
	queue          list.List
	units          map[string]Unit
	unFinished     map[string]Unit
	mux            sync.Mutex
	waitForModDone chan modResult
	maxParallel    int
	hasError       bool
	reverse        bool
	unitsErrors    map[string]error
	sigTrap        chan os.Signal
	stopChan       chan struct{}
}

func (g *grapher) Init(project *Project, maxParallel int, reverse bool) error {
	if err := checkUnitDependencies(project); err != nil {
		return err
	}
	if maxParallel < 1 {
		return fmt.Errorf("maxParallel should be greater then 0")
	}
	g.units = make(map[string]Unit)
	g.unitsErrors = make(map[string]error)
	g.unFinished = make(map[string]Unit)
	for key, mod := range project.Units {
		g.units[key] = mod
		g.unFinished[key] = mod
	}
	g.maxParallel = maxParallel
	g.queue.Init()
	g.inProgress = make(map[string]Unit)
	g.finished = make(map[string]modResult)
	g.waitForModDone = make(chan modResult)
	g.hasError = false
	g.reverse = reverse
	g.updateQueue()
	g.stopChan = make(chan struct{})
	g.listenHupSig()
	return nil
}

func (g *grapher) HasError() bool {
	return g.hasError
}

func (g *grapher) updateQueue() int {
	if g.reverse {
		return g.updateReverseQueue()
	}
	return g.updateDirectQueue()
}

func (g *grapher) updateDirectQueue() int {
	count := 0
	for key, mod := range g.units {
		isReady := true
		if !mod.Dependencies().IsEmpty() {
			for _, dep := range mod.Dependencies().Slice() {
				if er, ok := g.finished[dep.Unit.Key()]; !ok || er.Error != nil {
					isReady = false
					break
				}
			}
		}
		if isReady {
			g.queue.PushBack(mod)
			delete(g.units, key)
			count++
		}
	}
	return count
}

func (g *grapher) updateReverseQueue() int {
	count := 0
	for key, mod := range g.units {
		isReady := true
		dependedMods := findDependedUnits(g.unFinished, mod)
		if len(dependedMods) > 0 {
			isReady = false
		}
		if isReady {
			g.queue.PushBack(mod)
			delete(g.units, key)
			count++
		}
	}
	return count
}

func (g *grapher) GetNextAsync() (Unit, func(error), error) {
	g.mux.Lock()
	defer g.mux.Unlock()
	for {
		if config.Interupted {
			g.queue.Init()
			g.units = make(map[string]Unit)
			g.updateQueue()
			return nil, nil, fmt.Errorf("interupted")
		}
		if g.queue.Len() > 0 && len(g.inProgress) < g.maxParallel {
			modElem := g.queue.Front()
			mod := modElem.Value.(Unit)
			finFunc := func(err error) {
				g.waitForModDone <- modResult{mod, err}
			}
			g.queue.Remove(modElem)
			g.inProgress[mod.Key()] = mod
			return mod, finFunc, nil
		}
		if g.Len() == 0 {
			return nil, nil, nil
		}
		doneMod := <-g.waitForModDone
		g.setUnitDone(doneMod)
		if doneMod.Error != nil {
			return doneMod.Mod, nil, fmt.Errorf("error while unit running")
		}
	}
}

func (g *grapher) GetNextSync() Unit {
	if g.Len() == 0 {
		return nil
	}
	modElem := g.queue.Front()
	mod := modElem.Value.(Unit)
	g.queue.Remove(modElem)
	g.setUnitDone(modResult{mod, nil})
	return mod
}

func (g *grapher) GetSequenceSet() []Unit {
	res := make([]Unit, g.Len())
	mCount := g.Len()
	for i := 0; i < mCount; i++ {
		md := g.GetNextSync()
		if md == nil {
			log.Fatal("Building apply units set: geting nil unit, undefined behavior")
		}
		res[i] = md
	}
	return res
}

func (g *grapher) setUnitDone(doneMod modResult) {
	g.finished[doneMod.Mod.Key()] = doneMod
	delete(g.inProgress, doneMod.Mod.Key())
	delete(g.unFinished, doneMod.Mod.Key())
	if doneMod.Error != nil {
		g.unitsErrors[doneMod.Mod.Key()] = doneMod.Error
		g.hasError = true
	}
	g.updateQueue()
}

func (g *grapher) Errors() map[string]error {
	return g.unitsErrors
}

func (g *grapher) Wait() {
	for {
		if len(g.inProgress) == 0 {
			return
		}
		doneMod := <-g.waitForModDone
		g.setUnitDone(doneMod)
	}
}

func (g *grapher) Len() int {
	return len(g.units) + g.queue.Len() + len(g.inProgress)
}

func checkUnitDependencies(p *Project) error {
	errDepth := 15
	for _, mod := range p.Units {
		if ok := checkDependenciesRecursive(mod, errDepth); !ok {
			return fmt.Errorf("Unresolved dependency in unit %v.%v", mod.Stack().Name, mod.Name())
		}
	}
	return nil
}

func checkDependenciesRecursive(mod Unit, maxDepth int) bool {
	if maxDepth == 0 {
		return false
	}
	//log.Errorf("checkDependenciesRecursive %v", mod.Dependencies())
	for _, dep := range mod.Dependencies().Slice() {
		// log.Errorf("checkDependenciesRecursive FOR %v\n %+v", dep.Unit.Name(), mod.Name())
		if ok := checkDependenciesRecursive(dep.Unit, maxDepth-1); !ok {
			return false
		}
	}
	return true
}

func findDependedUnits(modList map[string]Unit, targetMod Unit) map[string]Unit {
	res := map[string]Unit{}
	for key, mod := range modList {
		// log.Infof("findDependedUnits '%v':", mod.Name())
		if mod.Key() == targetMod.Key() {
			continue
		}
		for _, dep := range mod.Dependencies().Slice() {
			if dep.Unit.Key() == targetMod.Key() {
				// log.Infof("      '%v':", dep.TargetUnitName)
				// log.Warnf("Tm: %v, M: %v Dependency: %v", targetMod.Name(), mod.Name(), dep.TargetUnitName)
				res[key] = mod
			}
		}
	}
	//log.Debugf("Searching depended from unit: %v\n Result: %v", targetMod.Name(), res)
	return res
}

func (g *grapher) listenHupSig() {
	signals := []os.Signal{syscall.SIGTERM, syscall.SIGINT}
	g.sigTrap = make(chan os.Signal, 1)
	signal.Notify(g.sigTrap, signals...)
	// log.Warn("Listening signals...")
	go func() {
		for {
			select {
			case <-g.sigTrap:
				config.Interupted = true
			case <-g.stopChan:
				// log.Warn("Stop listening")
				signal.Stop(g.sigTrap)
				g.sigTrap <- nil
				close(g.sigTrap)
				return
			}
		}
	}()
	return
}

func (g *grapher) Close() error {
	g.stopChan <- struct{}{}
	return nil
}
