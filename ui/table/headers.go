package table

import (
	"fmt"
	"strings"
	"unicode"
)

const UNKNOWN_HEADER_MAPPING rune = '_'

func NewHeader(title string) Header {
	return Header{
		Key:    KeyifyHeader(title),
		Title:  title,
		Hidden: false,
	}
}

func NewHeadersFromStrings(titles []string) (headers []Header) {
	for _, t := range titles {
		headers = append(headers, NewHeader(t))
	}
	return
}

func (t *Table) SetColumnVisibility(headers []Header) error {
	for tableHeaderIdx := range t.Header {
		t.Header[tableHeaderIdx].Hidden = true
	}

	for _, header := range headers {
		foundHeader := false

		for tableHeaderIdx, tableHeader := range t.Header {
			if tableHeader.Key == header.Key || tableHeader.Title == header.Title {
				t.Header[tableHeaderIdx].Hidden = false
				foundHeader = true

				break
			}
		}

		if !foundHeader {
			// key may be empty; if title is present
			return fmt.Errorf("Failed to find header: %s", header.Key) //nolint:staticcheck
		}
	}

	return nil
}

func (t *Table) SetColumnVisibilityFiltered(headers []Header, filterHeaders []Header) error {
	for tableHeaderIdx := range t.Header {
		t.Header[tableHeaderIdx].Hidden = true
	}

	for _, header := range headers {
		foundHeader := false

		for tableHeaderIdx, tableHeader := range t.Header {
			if tableHeader.Key == header.Key || tableHeader.Title == header.Title {
				t.Header[tableHeaderIdx].Hidden = false
				foundHeader = true

				break
			}
		}

		if !foundHeader {
			for _, filterHeader := range filterHeaders {
				if filterHeader.Key == header.Key || filterHeader.Title == header.Title {
					foundHeader = true

					break
				}
			}
		}

		if !foundHeader {
			// key may be empty; if title is present
			return fmt.Errorf("Failed to find header: %s", header.Key) //nolint:staticcheck
		}
	}

	return nil
}

func KeyifyHeader(header string) string {
	splittedStrings := strings.Split(cleanHeader(header), " ")
	splittedTrimmedStrings := []string{}
	for _, s := range splittedStrings {
		if s != "" {
			splittedTrimmedStrings = append(splittedTrimmedStrings, s)
		}
	}

	join := strings.Join(splittedTrimmedStrings, "_")
	if len(join) == 0 {
		return string(UNKNOWN_HEADER_MAPPING)
	}
	return join
}

func cleanHeader(header string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsLetter(r) || unicode.IsNumber(r) {
			return unicode.ToLower(r)
		} else {
			return ' '
		}
	}, header)
}
