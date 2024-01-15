package printer

import (
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/spf13/pflag"
)

type JobDefinition struct {
	Name      string `json:"name"`
	CreatedBy string `json:"createdBy"`
	Status    string `json:"status"`
	ExitCode  int    `json:"exitCode"`
}

// Printer is an interface that knows how to print objects.
type Printer interface {
	// Print receives an object, formats it and prints it to a writer.
	Print(in JobDefinition, w io.Writer) error
}

// JobPrinter provides functionality to print a given resource in requested format.
// Can be configured with pflag.FlagSet.
type JobPrinter struct {
	writer       io.Writer
	outputFormat PrintFormat

	printers map[PrintFormat]Printer
}

// NewForJob returns a new JobPrinter instance.
func NewForJob(w io.Writer) *JobPrinter {
	return &JobPrinter{
		writer: w,
		printers: map[PrintFormat]Printer{
			JSONFormat:  &JSON{},
			YAMLFormat:  &YAML{},
			TableFormat: &Table{},
		},
		outputFormat: TableFormat,
	}
}

// RegisterFlags registers JobPrinter terminal flags.
func (r *JobPrinter) RegisterFlags(flags *pflag.FlagSet) {
	flags.VarP(&r.outputFormat, "output", "o", fmt.Sprintf("Output format. One of: %s", r.availablePrinters()))
}

// Print prints received object in requested format.
func (r *JobPrinter) Print(in JobDefinition) error {
	printer, found := r.printers[r.outputFormat]
	if !found {
		return fmt.Errorf("printer %q is not available", r.outputFormat)
	}

	return printer.Print(in, r.writer)
}

func (r *JobPrinter) availablePrinters() string {
	var out []string
	for key := range r.printers {
		out = append(out, key.String())
	}

	// We generate doc automatically, so it needs to be deterministic
	sort.Strings(out)

	return strings.Join(out, " | ")
}
