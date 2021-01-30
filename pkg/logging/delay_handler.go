package logging

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"github.com/apex/log"
	"github.com/tj/go-spin"
)

// BannerPrintStyle - describe how to print banner when log is hidden.
type BannerPrintStyle int

const (
	// SpinnerStyle show spinet with timer.
	SpinnerStyle BannerPrintStyle = iota
	// LineStile - print banner in newline with time 10 sec. (like terraform)
	LineStile
)

// DelayHandler implementation.
type DelayHandler struct {
	entries       chan *log.Entry
	start         time.Time
	spin          *spin.Spinner
	done          chan struct{}
	w             io.Writer
	afterTimeOut  bool
	spinMux       sync.Mutex
	tmOut         time.Duration
	banner        string
	bannerPrinter func(h *DelayHandler)
	bannerTicker  time.Duration
}

// NewDelayHandler handler.
func NewDelayHandler(timeout time.Duration, banner string, bannerStyle BannerPrintStyle) *DelayHandler {
	h := &DelayHandler{
		entries:      make(chan *log.Entry),
		done:         make(chan struct{}),
		start:        time.Now(),
		spin:         spin.New(),
		w:            &bytes.Buffer{},
		afterTimeOut: false,
		tmOut:        timeout,
		banner:       banner,
	}
	switch bannerStyle {
	case SpinnerStyle:
		h.bannerTicker = 100 * time.Millisecond
		h.bannerPrinter = spinBannerPrinter
	case LineStile:
		h.bannerTicker = 10 * time.Second
		h.bannerPrinter = newlineBannerPrinter
	default:
		h.bannerTicker = 10 * time.Second
		h.bannerPrinter = newlineBannerPrinter
	}
	h.bannerPrinter(h)
	go h.loop()

	return h
}

// Close the handler.
func (h *DelayHandler) Close() error {
	h.done <- struct{}{}
	close(h.done)
	close(h.entries)
	return nil
}

// loop for rendering.
func (h *DelayHandler) loop() {
	ticker := time.NewTicker(h.bannerTicker)
	timeOut := time.NewTimer(h.tmOut)
	for {
		select {
		case e := <-h.entries:
			h.spinMux.Lock()
			fmt.Fprint(h.w, logFormatter(e))
			h.spinMux.Unlock()
		case <-ticker.C:
			h.spinMux.Lock()
			if !h.afterTimeOut {
				h.bannerPrinter(h)
			}
			h.spinMux.Unlock()
		case <-timeOut.C:
			h.spinMux.Lock()
			h.afterTimeOut = true
			timeOut.Stop()
			fmt.Fprintf(os.Stdout, "\033[2K\rTimeout reached, Printing hidden log: \n%s", h.w)
			h.w = os.Stdout
			h.spinMux.Unlock()
		case <-h.done:
			ticker.Stop()
			if !h.afterTimeOut {
				fmt.Println("\n\n\n\n")
			}
			return
		}
	}
}

// HandleLog implements log.Handler.
func (h *DelayHandler) HandleLog(e *log.Entry) error {
	h.entries <- e
	return nil
}

func spinBannerPrinter(h *DelayHandler) {
	fmt.Printf("\r    %s %s %-7s", h.banner, h.spin.Next(), time.Since(h.start).Round(time.Second))
}

func newlineBannerPrinter(h *DelayHandler) {
	fmt.Printf("    %s %-7s\n", h.banner, time.Since(h.start).Round(time.Second))
}
