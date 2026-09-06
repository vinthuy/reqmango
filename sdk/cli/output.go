package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/spf13/cobra"
)

// PrintTable writes aligned columns to w.
func PrintTable(w io.Writer, headers []string, rows [][]string) {
	cols := len(headers)
	widths := make([]int, cols)
	for i, h := range headers {
		widths[i] = len(h)
	}
	for _, row := range rows {
		for i := 0; i < cols && i < len(row); i++ {
			if l := len(row[i]); l > widths[i] {
				widths[i] = l
			}
		}
	}
	writeRow := func(cells []string) {
		for i := 0; i < cols; i++ {
			if i > 0 {
				fmt.Fprint(w, "  ")
			}
			cell := ""
			if i < len(cells) {
				cell = cells[i]
			}
			fmt.Fprintf(w, "%-*s", widths[i], cell)
		}
		fmt.Fprintln(w)
	}
	writeRow(headers)
	sep := make([]string, cols)
	for i := range sep {
		sep[i] = strings.Repeat("-", widths[i])
	}
	writeRow(sep)
	for _, row := range rows {
		writeRow(row)
	}
}

// PrintJSON writes v as indented JSON.
func PrintJSON(w io.Writer, v any) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(w, string(data))
	return err
}

// printResult renders rows as a table, or v as JSON when --output=json.
func printResult(cmd *cobra.Command, headers []string, rows [][]string, v any) error {
	if cmd.Flag("output").Value.String() == "json" {
		return PrintJSON(cmd.OutOrStdout(), v)
	}
	PrintTable(cmd.OutOrStdout(), headers, rows)
	return nil
}
