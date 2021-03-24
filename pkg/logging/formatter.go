package logging

import (
	"path"
	"runtime"
	"strings"
)

type fmtVerb int

const (
	fmtVerbTime fmtVerb = iota
	fmtVerbLevel
	fmtVerbID
	fmtVerbPid
	fmtVerbProgram
	fmtVerbModule
	fmtVerbMessage
	fmtVerbLongfile
	fmtVerbShortfile
	fmtVerbLongpkg
	fmtVerbShortpkg
	fmtVerbLongfunc
	fmtVerbShortfunc
	fmtVerbCallpath
	fmtVerbLevelColor
	fmtVerbUnknown
	fmtVerbStatic
)

func formatFuncName(v fmtVerb, f string) string {
	i := strings.LastIndex(f, "/")
	j := strings.Index(f[i+1:], ".")
	if j < 1 {
		return "???"
	}
	pkg, fun := f[:i+j+1], f[i+j+2:]
	switch v {
	case fmtVerbLongpkg:
		return pkg
	case fmtVerbShortpkg:
		return path.Base(pkg)
	case fmtVerbLongfunc:
		return fun
	case fmtVerbShortfunc:
		i = strings.LastIndex(fun, ".")
		return fun[i+1:]
	}
	panic("unexpected func formatter")
}

func formatCallpath(calldepth int, depth int) string {
	v := ""
	callers := make([]uintptr, 64)
	n := runtime.Callers(calldepth+2, callers)
	oldPc := callers[n-1]

	start := n - 3
	if depth > 0 && start >= depth {
		start = depth - 1
		v += "~."
	}
	recursiveCall := false
	for i := start; i >= 0; i-- {
		pc := callers[i]
		if oldPc == pc {
			recursiveCall = true
			continue
		}
		oldPc = pc
		if recursiveCall {
			recursiveCall = false
			v += ".."
		}
		if i < start {
			v += "."
		}
		if f := runtime.FuncForPC(pc); f != nil {
			v += formatFuncName(fmtVerbLongfunc, f.Name())
		}
	}
	return v
}
