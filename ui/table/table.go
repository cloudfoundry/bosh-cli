package table

import (
	"fmt"
	"io"
	"sort"
)

func (t Table) AsRows() [][]Value {
	rows := [][]Value{}

	totalRows := 0

	if len(t.Sections) > 0 {
		for _, s := range t.Sections {
			if s.FirstColumn != nil && len(s.FirstColumn.String()) > 0 {
				if len(s.Rows) > 0 && len(s.Rows[0]) > 0 {
					for i, _ := range s.Rows {
						s.Rows[i][0] = s.FirstColumn
					}
				}
			}

			totalRows += len(s.Rows)

			for _, r := range s.Rows {
				rows = append(rows, r)
			}
		}
	}

	if len(t.Rows) > 0 {
		totalRows += len(t.Rows)

		for _, r := range t.Rows {
			rows = append(rows, r)
		}
	}

	// Fill in nils
	for i, r := range rows {
		for j, c := range r {
			if c == nil {
				rows[i][j] = ValueNone{}
			}
		}
	}

	// Sort all rows
	sort.Sort(Sorting{t.SortBy, rows})

	// Dedup first column
	if !t.FillFirstColumn {
		var lastVal Value

		for _, r := range rows {
			if lastVal == nil {
				lastVal = r[0]
			} else if lastVal.String() == r[0].String() {
				r[0] = ValueString{"~"}
			} else {
				lastVal = r[0]
			}
		}
	}

	return rows
}

func (t Table) Print(w io.Writer) error {
	err := t.printHeader(w)
	if err != nil {
		return err
	}

	if len(t.BackgroundStr) == 0 {
		t.BackgroundStr = " "
	}

	if len(t.BorderStr) == 0 {
		t.BorderStr = "  "
	}

	writer := NewWriter(w, "-", t.BackgroundStr, t.BorderStr)

	if t.Transpose {
		var newRows [][]Value

		headerVals := buildHeaderVals(t)

		for _, row := range t.Rows {
			for i, val := range row {
				newRows = append(newRows, []Value{headerVals[i], val})
			}
		}

		t.Rows = newRows
	} else {
		if len(t.Header) > 0 {
			writer.Write(buildHeaderVals(t))
		}
	}

	rows := t.AsRows()

	for _, row := range rows {
		writer.Write(row)
	}

	err = writer.Flush()
	if err != nil {
		return err
	}

	return t.printFooter(w, len(rows))
}

func buildHeaderVals(t Table) []Value {
	var headerVals []Value

	if len(t.Header) > 0 {
		for _, h := range t.Header {
			headerVals = append(headerVals, ValueFmt{
				V:    ValueString{h},
				Func: t.HeaderFormatFunc,
			})
		}
	}
	return headerVals
}

func (t Table) printHeader(w io.Writer) error {
	if len(t.Title) > 0 {
		_, err := fmt.Fprintf(w, "%s\n\n", t.Title)
		if err != nil {
			return err
		}
	}

	return nil
}

func (t Table) printFooter(w io.Writer, num int) error {
	if len(t.Notes) > 0 {
		_, err := fmt.Fprintf(w, "\n")
		if err != nil {
			return err
		}

		for _, n := range t.Notes {
			_, err := fmt.Fprintf(w, "%s\n", n)
			if err != nil {
				return err
			}
		}
	}

	if len(t.Header) > 0 {
		_, err := fmt.Fprintf(w, "\n%d %s\n", num, t.Content)
		if err != nil {
			return err
		}
	}

	return nil
}
