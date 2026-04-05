package output

import (
	"encoding/json"
	"fmt"
	"io"
)

func WriteTable(w io.Writer, headers []string, rows [][]string) {
	widths := make([]int, len(headers))
	for i, header := range headers {
		widths[i] = len(header)
	}
	for _, row := range rows {
		for i, cell := range row {
			if len(cell) > widths[i] {
				widths[i] = len(cell)
			}
		}
	}
	for i, header := range headers {
		fmt.Fprintf(w, "% -*s", widths[i], header)
		if i < len(headers)-1 {
			fmt.Fprint(w, "  ")
		}
	}
	fmt.Fprintln(w)
	for _, row := range rows {
		for i, cell := range row {
			fmt.Fprintf(w, "% -*s", widths[i], cell)
			if i < len(row)-1 {
				fmt.Fprint(w, "  ")
			}
		}
		fmt.Fprintln(w)
	}
}

func WriteJSON(w io.Writer, value any) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(value)
}
