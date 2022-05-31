package logging

import (
	"fmt"
	"time"

	"github.com/apex/log"
	"github.com/shalb/cluster.dev/pkg/colors"
)

// color function.
type colorFunc func(string, ...interface{}) string

// Colors mapping.
var Colors = []colorFunc{
	log.DebugLevel: colors.Fmt(colors.Purple).Sprintf,
	log.InfoLevel:  colors.Fmt(colors.Cyan).Sprintf,
	log.WarnLevel:  colors.Fmt(colors.Yellow).Sprintf,
	log.ErrorLevel: colors.Fmt(colors.LightRed).Sprintf,
	log.FatalLevel: colors.Fmt(colors.Red).Sprintf,
}

// Strings mapping.
var Strings = [...]string{
	log.DebugLevel: "DEBUG",
	log.InfoLevel:  "INFO",
	log.WarnLevel:  "WARN",
	log.ErrorLevel: "ERROR",
	log.FatalLevel: "FATAL",
}

// utilStartTime time.
var utilStartTime = time.Now()

// loggingInit - initial function for logging subsystem.
func init() {
	log.SetHandler(NewLogStdHandler())
}

var traceLog bool

func InitLogLevel(ll string, trace bool) {
	lvl, err := log.ParseLevel(ll)
	if err != nil {
		log.Fatalf("Can't parse logging level '%s': %s", ll, err.Error())
	}
	log.SetLevel(lvl)
	traceLog = trace
	Colors = []colorFunc{
		log.DebugLevel: colors.Fmt(colors.Purple).Sprintf,
		log.InfoLevel:  colors.Fmt(colors.Cyan).Sprintf,
		log.WarnLevel:  colors.Fmt(colors.Yellow).Sprintf,
		log.ErrorLevel: colors.Fmt(colors.LightRed).Sprintf,
		log.FatalLevel: colors.Fmt(colors.Red).Sprintf,
	}
}

func logFormatter(e *log.Entry) string {
	color := Colors[e.Level]
	level := Strings[e.Level]
	names := e.Fields.Names()

	// ts := time.Since(utilStartTime) / time.Second
	tsR := time.Now().Format("15:04:05")

	output := fmt.Sprintf("%s %s", colors.Fmt(colors.LightWhite).Sprint(tsR), color("[%s]", level))

	if len(names) > 0 {
		output += " "
	}
	for _, name := range names {
		output += fmt.Sprintf("[%v]", e.Fields.Get(name))
	}
	if traceLog {
		traceMsg := colors.Fmt(colors.LightWhite).Sprintf("<%s>", FormatCallPath(6, 15))
		output = fmt.Sprintf("%s %s", output, traceMsg)
	}
	output = fmt.Sprintf("%s %-25s", output, e.Message)

	return fmt.Sprintln(output)
}
