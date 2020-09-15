package logging

import (
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/apex/log"
)

// NewLogStdHandler - new custom log handler for apex log lib.
func NewLogStdHandler() *StdHandler {
	return &StdHandler{}
}

// StdHandler implementation.
type StdHandler struct {
	mu     sync.Mutex
	Writer io.Writer
}

// HandleLog implements log.Handler.
func (h *StdHandler) HandleLog(e *log.Entry) error {

	h.mu.Lock()
	defer h.mu.Unlock()

	if e.Level >= log.ErrorLevel {
		h.Writer = os.Stderr
	} else {
		h.Writer = os.Stdout
	}

	fmt.Fprint(h.Writer, logFormatter(e))

	return nil
}
