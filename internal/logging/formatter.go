// Copyright 2013, Örjan Persson. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package logging

import (
	"os"
	"path"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
)

// TODO see Formatter interface in fmt/print.go
// TODO try text/template, maybe it have enough performance
// TODO other template systems?
// TODO make it possible to specify formats per backend?
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

	// Keep last, there are no match for these below.
	fmtVerbUnknown
	fmtVerbStatic
)

var fmtVerbs = []string{
	"time",
	"level",
	"id",
	"pid",
	"program",
	"module",
	"message",
	"longfile",
	"shortfile",
	"longpkg",
	"shortpkg",
	"longfunc",
	"shortfunc",
	"callpath",
	"color",
}

const rfc3339Milli = "2006-01-02T15:04:05.999Z07:00"

var defaultVerbsLayout = []string{
	rfc3339Milli,
	"s",
	"d",
	"d",
	"s",
	"s",
	"s",
	"s",
	"s",
	"s",
	"s",
	"s",
	"s",
	"0",
	"",
}

var (
	pid     = os.Getpid()
	program = filepath.Base(os.Args[0])
)

func getFmtVerbByName(name string) fmtVerb {
	for i, verb := range fmtVerbs {
		if name == verb {
			return fmtVerb(i)
		}
	}
	return fmtVerbUnknown
}

var formatRe = regexp.MustCompile(`%{([a-z]+)(?::(.*?[^\\]))?}`)

type part struct {
	verb   fmtVerb
	layout string
}

// stringFormatter contains a list of parts which explains how to build the
// formatted string passed on to the logging backend.
type stringFormatter struct {
	parts []part
}

// NewStringFormatter returns a new Formatter which outputs the log record as a
// string based on the 'verbs' specified in the format string.
//
// The verbs:
//
// General:
//     %{id}        Sequence number for log message (uint64).
//     %{pid}       Process id (int)
//     %{time}      Time when log occurred (time.Time)
//     %{level}     Log level (Level)
//     %{module}    Module (string)
//     %{program}   Basename of os.Args[0] (string)
//     %{message}   Message (string)
//     %{longfile}  Full file name and line number: /a/b/c/d.go:23
//     %{shortfile} Final file name element and line number: d.go:23
//     %{callpath}  Callpath like main.a.b.c...c  "..." meaning recursive call ~. meaning truncated path
//     %{color}     ANSI color based on log level
//
// For normal types, the output can be customized by using the 'verbs' defined
// in the fmt package, eg. '%{id:04d}' to make the id output be '%04d' as the
// format string.
//
// For time.Time, use the same layout as time.Format to change the time format
// when output, eg "2006-01-02T15:04:05.999Z-07:00".
//
// For the 'color' verb, the output can be adjusted to either use bold colors,
// i.e., '%{color:bold}' or to reset the ANSI attributes, i.e.,
// '%{color:reset}' Note that if you use the color verb explicitly, be sure to
// reset it or else the color state will persist past your log message.  e.g.,
// "%{color:bold}%{time:15:04:05} %{level:-8s}%{color:reset} %{message}" will
// just colorize the time and level, leaving the message uncolored.
//
// For the 'callpath' verb, the output can be adjusted to limit the printing
// the stack depth. i.e. '%{callpath:3}' will print '~.a.b.c'
//
// Colors on Windows is unfortunately not supported right now and is currently
// a no-op.
//
// There's also a couple of experimental 'verbs'. These are exposed to get
// feedback and needs a bit of tinkering. Hence, they might change in the
// future.
//
// Experimental:
//     %{longpkg}   Full package path, eg. github.com/go-logging
//     %{shortpkg}  Base package path, eg. go-logging
//     %{longfunc}  Full function name, eg. littleEndian.PutUint32
//     %{shortfunc} Base function name, eg. PutUint32
//     %{callpath}  Call function path, eg. main.a.b.c

// formatFuncName tries to extract certain part of the runtime formatted
// function name to some pre-defined variation.
//
// This function is known to not work properly if the package path or name
// contains a dot.
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
			v += formatFuncName(fmtVerbShortfunc, f.Name())
		}
	}
	return v
}
