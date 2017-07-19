package table

import (
	"fmt"
	"strings"
	"unicode"
)

func NewHeader(title string) Header {
	return Header{
		Key:    KeyifyHeader(title),
		Title:  title,
		Hidden: false,
	}
}

func (t *Table) SetColumnVisibility(headers []Header) error {
	for i, _ := range t.Header {
		t.Header[i].Hidden = true
	}

	for _, header := range headers {
		found := false
		foundHeaders := []string{}

		for i, tableHeader := range t.Header {
			if tableHeader.Key == header.Key || tableHeader.Title == header.Title {
				t.Header[i].Hidden = false
				found = true
				break
			}
			foundHeaders = append(foundHeaders, "'"+tableHeader.Key+"'")
		}

		if !found {
			return fmt.Errorf("Failed to find header '%s' (found headers: %s)", header.Key, strings.Join(foundHeaders, ", "))
		}
	}

	return nil
}

func KeyifyHeader(header string) string {
	pieces := []string{}

	for _, s := range strings.Split(cleanHeader(header), " ") {
		if s != "" {
			pieces = append(pieces, s)
		}
	}

	return strings.Join(pieces, "_")
}

func cleanHeader(header string) string {
	mapFunc := func(r rune) rune {
		if unicode.IsLetter(r) || unicode.IsNumber(r) {
			return unicode.ToLower(r)
		} else if r == '(' || r == ')' {
			return -1
		}
		return ' '
	}
	return strings.Map(mapFunc, header)
}
