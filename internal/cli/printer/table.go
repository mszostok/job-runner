package printer

import (
	"io"
	"strconv"

	"github.com/olekukonko/tablewriter"
)

var _ Printer = &Table{}

// Table prints data in table format.
type Table struct{}

// Print creates table with provided data and writes it to a given writer.
func (p *Table) Print(in JobDefinition, w io.Writer) error {
	table := tablewriter.NewWriter(w)
	table.SetAutoWrapText(true)
	table.SetColumnSeparator(" ")
	table.SetBorder(false)
	table.SetRowLine(true)

	table.SetHeader([]string{"Name", "Created by", "Status", "Exit code"})
	table.Append([]string{
		in.Name,
		in.CreatedBy,
		in.Status,
		strconv.Itoa(in.ExitCode),
	})

	table.Render()

	return nil
}
