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
	Mod   Module
	Error error
}

type grapher struct {
	finished       map[string]modResult
	inProgress     map[string]Module
	queue          list.List
	modules        map[string]Module
	unFinished     map[string]Module
	mux            sync.Mutex
	waitForModDone chan modResult
	maxParallel    int
	hasError       bool
	reverse        bool
	modulesErrors  map[string]error
	sigTrap        chan os.Signal
	stopChan       chan struct{}
}

func (g *grapher) Init(project *Project, maxParallel int, reverse bool) error {
	if err := checkModulesDependencies(project); err != nil {
		return err
	}
	if maxParallel < 1 {
		return fmt.Errorf("maxParallel should be greater then 0")
	}
	g.modules = make(map[string]Module)
	g.modulesErrors = make(map[string]error)
	g.unFinished = make(map[string]Module)
	for key, mod := range project.Modules {
		g.modules[key] = mod
		g.unFinished[key] = mod
	}
	g.maxParallel = maxParallel
	g.queue.Init()
	g.inProgress = make(map[string]Module)
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
	for key, mod := range g.modules {
		isReady := true
		if len(*mod.Dependencies()) > 0 {
			for _, dep := range *mod.Dependencies() {
				if er, ok := g.finished[dep.Module.Key()]; !ok || er.Error != nil {
					isReady = false
					break
				}
			}
		}
		if isReady {
			g.queue.PushBack(mod)
			delete(g.modules, key)
			count++
		}
	}
	return count
}

func (g *grapher) updateReverseQueue() int {
	count := 0
	for key, mod := range g.modules {
		isReady := true
		dependedMods := findDependedModules(g.unFinished, mod)
		if len(dependedMods) > 0 {
			isReady = false
		}
		if isReady {
			g.queue.PushBack(mod)
			delete(g.modules, key)
			count++
		}
	}
	return count
}

func (g *grapher) GetNextAsync() (Module, func(error), error) {
	g.mux.Lock()
	defer g.mux.Unlock()
	for {
		if config.Interupted {
			g.queue.Init()
			g.modules = make(map[string]Module)
			g.updateQueue()
			return nil, nil, fmt.Errorf("interupted")
		}
		if g.queue.Len() > 0 && len(g.inProgress) < g.maxParallel {
			modElem := g.queue.Front()
			mod := modElem.Value.(Module)
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
		g.setModuleDone(doneMod)
		if doneMod.Error != nil {
			return doneMod.Mod, nil, fmt.Errorf("error while module running")
		}
	}
}

func (g *grapher) GetNextSync() Module {
	if g.Len() == 0 {
		return nil
	}
	modElem := g.queue.Front()
	mod := modElem.Value.(Module)
	g.queue.Remove(modElem)
	g.setModuleDone(modResult{mod, nil})
	return mod
}

func (g *grapher) GetSequenceSet() []Module {
	res := make([]Module, g.Len())
	mCount := g.Len()
	for i := 0; i < mCount; i++ {
		md := g.GetNextSync()
		if md == nil {
			log.Fatal("Building apply modules set: geting nil module, undefined behavior")
		}
		res[i] = md
	}
	return res
}

func (g *grapher) setModuleDone(doneMod modResult) {
	g.finished[doneMod.Mod.Key()] = doneMod
	delete(g.inProgress, doneMod.Mod.Key())
	delete(g.unFinished, doneMod.Mod.Key())
	if doneMod.Error != nil {
		g.modulesErrors[doneMod.Mod.Key()] = doneMod.Error
		g.hasError = true
	}
	g.updateQueue()
}

func (g *grapher) Errors() map[string]error {
	return g.modulesErrors
}

func (g *grapher) Wait() {
	for {
		if len(g.inProgress) == 0 {
			return
		}
		doneMod := <-g.waitForModDone
		g.setModuleDone(doneMod)
	}
}

func (g *grapher) Len() int {
	return len(g.modules) + g.queue.Len() + len(g.inProgress)
}

func checkModulesDependencies(p *Project) error {
	errDepth := 15
	for _, mod := range p.Modules {
		if ok := checkDependenciesRecursive(mod, errDepth); !ok {
			return fmt.Errorf("Unresolved dependency in module %v.%v", mod.StackName(), mod.Name())
		}
	}
	return nil
}

func checkDependenciesRecursive(mod Module, maxDepth int) bool {
	if maxDepth == 0 {
		return false
	}
	for _, dep := range *mod.Dependencies() {
		if ok := checkDependenciesRecursive(dep.Module, maxDepth-1); !ok {
			return false
		}
	}
	return true
}

func findDependedModules(modList map[string]Module, targetMod Module) map[string]Module {
	res := map[string]Module{}
	for key, mod := range modList {
		for _, dep := range *mod.Dependencies() {
			//log.Debugf("Tm: %v, M: %v Dependency: %v", targetMod.Name(), mod.Name(), dep.ModuleName)
			if dep.Module.Key() == targetMod.Key() {
				res[key] = mod
			}
		}
	}
	//log.Debugf("Searching depended from module: %v\n Result: %v", targetMod.Name(), res)
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
