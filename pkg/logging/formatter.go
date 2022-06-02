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
	fmtVerbLongFile
	fmtVerbShortFile
	fmtVerbLongPkg
	fmtVerbShortPkg
	fmtVerbLongFunc
	fmtVerbShortFunc
	fmtVerbCallPath
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
	case fmtVerbLongPkg:
		return pkg
	case fmtVerbShortPkg:
		return path.Base(pkg)
	case fmtVerbLongFunc:
		return fun
	case fmtVerbShortFunc:
		i = strings.LastIndex(fun, ".")
		return fun[i+1:]
	}
	panic("unexpected func formatter")
}

func FormatCallPath(callDepth int, depth int) string {
	v := ""
	callers := make([]uintptr, 64)
	n := runtime.Callers(callDepth+2, callers)
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
			v += formatFuncName(fmtVerbLongFunc, f.Name())
		}
	}
	return v
}
