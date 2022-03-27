package printer

import (
	"fmt"
	"io"

	"github.com/fatih/color"
)

// Spinner defines interface for terminal spinner.
type Spinner interface {
	Start(stage string)
	Active() bool
	Stop(msg string)
}

// StatusPrinter provides functionality to display steps progress in terminal.
type StatusPrinter struct {
	w       io.Writer
	spinner Spinner
	stage   string
}

// NewStatus returns a new Status instance.
func NewStatus(w io.Writer) *StatusPrinter {
	st := &StatusPrinter{
		w: w,
	}
	if IsSmartTerminal(w) {
		st.spinner = NewDynamicSpinner(w)
	} else {
		st.spinner = NewStaticSpinner(w)
	}

	return st
}

// Step starts spinner for a given step.
func (s *StatusPrinter) Step(stageFmt string, args ...interface{}) {
	// Finish previously started step
	s.End(true)

	started := ""
	s.stage = fmt.Sprintf(stageFmt, args...)
	msg := fmt.Sprintf("%s%s", s.stage, started)
	s.spinner.Start(msg)
}

// End marks started step as completed.
func (s *StatusPrinter) End(success bool) {
	if !s.spinner.Active() {
		return
	}

	var icon string
	if success {
		icon = color.GreenString("✓")
	} else {
		icon = color.RedString("✗")
	}

	msg := fmt.Sprintf(" %s %s\n", icon, s.stage)
	s.spinner.Stop(msg)
}
