package printer_test

import (
	"bytes"
	"testing"

	"github.com/sebdah/goldie/v2"
	"github.com/spf13/pflag"
	"github.com/stretchr/testify/require"

	"github.com/mszostok/job-runner/internal/cli/printer"
)

// TestJobPrinterOutput tests that Job outputter works properly in all formats.
//
// This test is based on golden file.
// If the `-test.update-golden` flag is set then the actual content is written
// to the golden file.
//
// To update golden files, run:
//   go test ./internal/cli/printer/... -run "^TestJobPrinterOutput$" -update
func TestJobPrinterOutput(t *testing.T) {
	tests := []struct {
		name   string
		output string
	}{
		{
			name:   "Should print Job in YAML format",
			output: "yaml",
		},
		{
			name:   "Should print Job in JSON format",
			output: "json",
		},
		{
			name:   "Should print Job in Table format",
			output: "table",
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			// given
			buff := &bytes.Buffer{}
			jobPrinter := printer.NewForJob(buff)

			flags := pflag.NewFlagSet("testing", pflag.ContinueOnError)
			jobPrinter.RegisterFlags(flags)

			job := printer.JobDefinition{
				Name:      "YourAdHere",
				CreatedBy: "testing",
				Status:    "SUCCEEDED",
				ExitCode:  0,
			}

			// when
			err := flags.Set("output", test.output)
			require.NoError(t, err)

			err = jobPrinter.Print(job)

			// then
			require.NoError(t, err)
			g := goldie.New(t, goldie.WithNameSuffix(".golden.txt"))
			g.Assert(t, t.Name(), buff.Bytes())
		})
	}
}

// TestStatusPrinterOutput tests that status outputter works properly.
//
// This test is based on golden file. To update golden files, run:
//   go test ./internal/cli/printer/... -run "^TestStatusPrinterOutput$" -update
func TestStatusPrinterOutput(t *testing.T) {
	tests := []struct {
		name    string
		success bool
	}{
		{
			name:    "Should finish with success",
			success: true,
		},
		{
			name:    "Should finish with failure ",
			success: false,
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			// given
			buff := &bytes.Buffer{}
			status := printer.NewStatus(buff)

			// when
			status.Step(test.name)
			status.End(test.success)

			// then
			g := goldie.New(t, goldie.WithNameSuffix(".golden.txt"))
			g.Assert(t, t.Name(), buff.Bytes())
		})
	}
}
