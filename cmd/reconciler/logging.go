package main

import (
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"github.com/apex/log"
)

// utilStartTime time.
var utilStartTime = time.Now()

// colors.
const (
	none   = 0
	red    = 31
	green  = 32
	yellow = 33
	blue   = 34
	purple = 35
	cyan   = 36
	gray   = 37
)

// Colors mapping.
var Colors = [...]int{
	log.DebugLevel: purple,
	log.InfoLevel:  blue,
	log.WarnLevel:  yellow,
	log.ErrorLevel: red,
	log.FatalLevel: red,
}

// Strings mapping.
var Strings = [...]string{
	log.DebugLevel: "DEBUG",
	log.InfoLevel:  "INFO",
	log.WarnLevel:  "WARN",
	log.ErrorLevel: "ERROR",
	log.FatalLevel: "FATAL",
}

// Handler implementation.
type Handler struct {
	mu     sync.Mutex
	Writer io.Writer
}

// NewLogHandler - new custom log handler for apex log lib.
func NewLogHandler() *Handler {
	return &Handler{}
}

// HandleLog implements log.Handler.
func (h *Handler) HandleLog(e *log.Entry) error {
	color := Colors[e.Level]
	level := Strings[e.Level]
	names := e.Fields.Names()

	h.mu.Lock()
	defer h.mu.Unlock()

	if e.Level >= log.ErrorLevel {
		h.Writer = os.Stderr
	} else {
		h.Writer = os.Stdout
	}

	ts := time.Since(utilStartTime) / time.Second
	tsR := time.Now().Format("15:04:05.000")

	fmt.Fprintf(h.Writer, "%s\033[%dm%6s\033[0m[%04d] %-25s", tsR, color, level, ts, e.Message)

	for _, name := range names {
		fmt.Fprintf(h.Writer, " \033[%dm%s\033[0m=%v", color, name, e.Fields.Get(name))
	}

	fmt.Fprintln(h.Writer)

	return nil
}

// loggingInit - initial function for logging subsystem.
func loggingInit() {
	log.SetHandler(NewLogHandler())
	lvl, err := log.ParseLevel(globalConfig.LogLevel)
	if err != nil {
		log.Fatalf("Can't parse logging level '%s': %s", globalConfig.LogLevel, err.Error())
	}
	log.SetLevel(lvl)
}
