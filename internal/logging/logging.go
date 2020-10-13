package logging

import (
	"fmt"
	"time"

	"github.com/apex/log"
	"github.com/aybabtme/rgbterm"
	"github.com/shalb/cluster.dev/internal/config"
)

// color function.
type colorFunc func(string) string

// purple string.
func purple(s string) string {
	return rgbterm.FgString(s, 186, 85, 211)
}

// gray string.
func gray(s string) string {
	return rgbterm.FgString(s, 150, 150, 150)
}

// blue string.
func blue(s string) string {
	return rgbterm.FgString(s, 77, 173, 247)
}

// cyan string.
func cyan(s string) string {
	return rgbterm.FgString(s, 34, 184, 207)
}

// green string.
func green(s string) string {
	return rgbterm.FgString(s, 0, 200, 255)
}

// red string.
func red(s string) string {
	return rgbterm.FgString(s, 194, 37, 92)
}

// yellow string.
func yellow(s string) string {
	return rgbterm.FgString(s, 252, 196, 25)
}

// Colors mapping.
var Colors = [...]colorFunc{
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

// utilStartTime time.
var utilStartTime = time.Now()

// loggingInit - initial function for logging subsystem.
func init() {
	log.SetHandler(NewLogStdHandler())

	lvl, err := log.ParseLevel(config.Global.LogLevel)
	if err != nil {
		log.Fatalf("Can't parse logging level '%s': %s", config.Global.LogLevel, err.Error())
	}
	log.SetLevel(lvl)
}

func logFormatter(e *log.Entry) string {
	color := Colors[e.Level]
	level := Strings[e.Level]
	names := e.Fields.Names()

	// ts := time.Since(utilStartTime) / time.Second
	tsR := time.Now().Format("15:04:05.000")
	output := fmt.Sprintf("%s %s %-25s", rgbterm.FgString(tsR, 204, 204, 204), color(fmt.Sprintf("[%s]", level)), e.Message)

	for _, name := range names {
		output += fmt.Sprintf(" %s=%v", color(name), e.Fields.Get(name))
	}

	return fmt.Sprintln(output)
}
