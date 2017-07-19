package table

import (
	"strings"
	"unicode"
)

func NewHeader(title string) Header {
	return Header{
		Key:   keyifyHeader(title),
		Title: title,
	}
}

func keyifyHeader(header string) string {
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
