package progress

import (
	"fmt"
	"io"
	"os"
	"strings"
)

var spinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

// Renderer displays Docker image pull progress.
// In TTY mode each layer occupies one line, updated in-place via ANSI codes,
// with a braille spinner on in-progress layers.
// In non-TTY mode only meaningful state transitions are printed.
type Renderer struct {
	w          io.Writer
	isTTY      bool
	order      []string
	states     map[string]*layerState
	drawn      int
	spinnerIdx int
}

type layerState struct {
	status   string
	progress string // Docker's pre-formatted bar, e.g. "[======>   ]  45MB/89MB"
}

// New returns a Renderer that writes to w.
func New(w io.Writer) *Renderer {
	return &Renderer{
		w:      w,
		isTTY:  isTerminal(w),
		states: make(map[string]*layerState),
	}
}

// newForTest returns a Renderer with an explicit TTY setting, for use in tests.
func newForTest(w io.Writer, isTTY bool) *Renderer {
	return &Renderer{
		w:      w,
		isTTY:  isTTY,
		states: make(map[string]*layerState),
	}
}

// Update records a layer event and refreshes the display.
func (r *Renderer) Update(id, status, progress string) {
	if r.isTTY {
		if _, ok := r.states[id]; !ok {
			r.order = append(r.order, id)
			r.states[id] = &layerState{}
		}
		r.states[id].status = status
		r.states[id].progress = progress
		r.redraw()
	} else {
		if isMeaningfulStatus(status) {
			_, _ = fmt.Fprintf(r.w, "%s: %s\n", id, status)
		}
	}
}

// Print writes a summary line with no layer ID (e.g. "Digest:", "Status:").
// In TTY mode it finalises the layer display first.
func (r *Renderer) Print(line string) {
	if r.isTTY {
		r.Flush()
	}
	_, _ = fmt.Fprintf(r.w, "%s\n", line)
}

// Flush finalises the in-place display: redraws all layers as static output
// and resets state. Safe to call when nothing has been drawn yet.
func (r *Renderer) Flush() {
	if !r.isTTY || r.drawn == 0 {
		return
	}
	r.clearLines(r.drawn)
	for _, id := range r.order {
		_, _ = fmt.Fprintf(r.w, "%s: %s\n", id, r.states[id].status)
	}
	// Reset so a subsequent pull (e.g. multi-stage FROM) starts fresh.
	r.drawn = 0
	r.order = nil
	r.states = make(map[string]*layerState)
}

// Clear removes the in-progress display without printing a final state.
func (r *Renderer) Clear() {
	r.clearLines(r.drawn)
	r.drawn = 0
	r.order = nil
	r.states = make(map[string]*layerState)
}

func (r *Renderer) redraw() {
	r.clearLines(r.drawn)
	r.drawn = 0
	spin := spinnerFrames[r.spinnerIdx%len(spinnerFrames)]
	r.spinnerIdx++
	for _, id := range r.order {
		s := r.states[id]
		switch {
		case isLayerDone(s.status) || strings.HasPrefix(s.status, "Pulling from"):
			// Done or header line: static, no spinner.
			_, _ = fmt.Fprintf(r.w, "  %s: %s\n", id, s.status)
		case s.progress != "":
			// Active with a progress bar from Docker.
			_, _ = fmt.Fprintf(r.w, "%s %s: %s %s\n", spin, id, s.status, s.progress)
		default:
			// Active without a progress bar: show spinner + status only.
			_, _ = fmt.Fprintf(r.w, "%s %s: %s\n", spin, id, s.status)
		}
		r.drawn++
	}
}

func isLayerDone(status string) bool {
	return status == "Pull complete" || status == "Already exists"
}

func (r *Renderer) clearLines(n int) {
	for i := 0; i < n; i++ {
		_, _ = fmt.Fprint(r.w, "\033[1A\033[2K")
	}
}

// isMeaningfulStatus reports whether a pull status is worth printing in
// non-TTY mode — i.e. it signals a real state change rather than churn.
func isMeaningfulStatus(status string) bool {
	switch status {
	case "Pull complete", "Already exists", "Download complete":
		return true
	}
	return strings.HasPrefix(status, "Pulling from") ||
		strings.HasPrefix(status, "Digest:") ||
		strings.HasPrefix(status, "Status:")
}

// isTerminal reports whether w is a character device (i.e. a terminal).
func isTerminal(w io.Writer) bool {
	f, ok := w.(*os.File)
	if !ok {
		return false
	}
	fi, err := f.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice != 0
}
