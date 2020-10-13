package logging

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strings"

	"github.com/apex/log"
)

// LogWriter io writer for logging driver.
type LogWriter struct {
	logFunc       func(string)
	trimmedString string
	ctx           *log.Entry
}

// NewLogWriter create log io writer.
func NewLogWriter(logLevel log.Level, fielder log.Fielder) (*LogWriter, error) {
	ctx := log.WithFields(fielder)
	//log.WithFields(fielder)
	var logFunctionsMap = map[log.Level]func(string){
		log.DebugLevel: ctx.Debug,
		log.InfoLevel:  ctx.Info,
		log.WarnLevel:  ctx.Warn,
		log.ErrorLevel: ctx.Error,
		log.FatalLevel: ctx.Fatal,
	}

	fn, ok := logFunctionsMap[logLevel]
	if !ok {
		return nil, fmt.Errorf("failed log level '%s'", logLevel)
	}

	lw := LogWriter{
		fn,
		"",
		ctx,
	}
	return &lw, nil
}

// Write func - standart io write interface.
func (l *LogWriter) Write(p []byte) (n int, err error) {
	n = 0
	reader := bufio.NewReader(bytes.NewReader(p))
	for {
		line, err := reader.ReadString('\n')

		if err != nil && err != io.EOF {
			log.Error(err.Error())
			break
		}

		n += len(line)
		if err == io.EOF {
			l.trimmedString = line
			break
		}
		if len(l.trimmedString) > 0 {
			line = fmt.Sprintf("%s%s", l.trimmedString, line)
			l.trimmedString = ""
		}
		line = strings.TrimSuffix(line, "\n")
		l.logFunc(line)

	}
	return n, nil
}
