package table

import (
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/dustin/go-humanize"

	boshuifmt "github.com/cloudfoundry/bosh-init/ui/fmt"
)

func (t ValueString) String() string { return t.S }
func (t ValueString) Value() Value   { return t }

func (t ValueString) Compare(other Value) int {
	otherS := other.(ValueString).S
	switch {
	case t.S == otherS:
		return 0
	case t.S < otherS:
		return -1
	default:
		return 1
	}
}

func (t ValueStrings) String() string { return strings.Join(t.S, "\n") }
func (t ValueStrings) Value() Value   { return t }

func (t ValueStrings) Compare(other Value) int {
	otherS := other.(ValueStrings).S
	switch {
	case len(t.S) == len(otherS):
		return 0
	case len(t.S) < len(otherS):
		return -1
	default:
		return 1
	}
}

func (t ValueInt) String() string { return strconv.Itoa(t.I) }
func (t ValueInt) Value() Value   { return t }

func (t ValueInt) Compare(other Value) int {
	otherI := other.(ValueInt).I
	switch {
	case t.I == otherI:
		return 0
	case t.I < otherI:
		return -1
	default:
		return 1
	}
}

func (t ValueBytes) String() string { return humanize.Bytes(t.I) }
func (t ValueBytes) Value() Value   { return t }

func (t ValueBytes) Compare(other Value) int {
	otherI := other.(ValueBytes).I
	switch {
	case t.I == otherI:
		return 0
	case t.I < otherI:
		return -1
	default:
		return 1
	}
}

func (t ValueTime) String() string { return t.T.Format(boshuifmt.TimeFullFmt) }
func (t ValueTime) Value() Value   { return t }

func (t ValueTime) Compare(other Value) int {
	otherT := other.(ValueTime).T
	switch {
	case t.T.Equal(otherT):
		return 0
	case t.T.Before(otherT):
		return -1
	default:
		return 1
	}
}

func (t ValueBool) String() string { return fmt.Sprintf("%t", t.B) }
func (t ValueBool) Value() Value   { return t }

func (t ValueBool) Compare(other Value) int {
	otherB := other.(ValueBool).B
	switch {
	case t.B == otherB:
		return 0
	case t.B == false && otherB == true:
		return -1
	default:
		return 1
	}
}

func (t ValueVersion) String() string { return t.V.String() }
func (t ValueVersion) Value() Value   { return t }

func (t ValueVersion) Compare(other Value) int {
	return t.V.Compare(other.(ValueVersion).V)
}

func (t ValueError) String() string {
	if t.E != nil {
		return t.E.Error()
	}
	return ""
}

func (t ValueError) Value() Value            { return t }
func (t ValueError) Compare(other Value) int { panic("Never callled") }

func (t ValueNone) String() string          { return "" }
func (t ValueNone) Value() Value            { return t }
func (t ValueNone) Compare(other Value) int { panic("Never callled") }

func (t ValueFmt) String() string          { return t.V.String() }
func (t ValueFmt) Value() Value            { return t.V }
func (t ValueFmt) Compare(other Value) int { panic("Never called") }

func (t ValueFmt) Fprintf(w io.Writer, pattern string, rest ...interface{}) (int, error) {
	if t.Func == nil {
		return fmt.Fprintf(w, pattern, rest...)
	}
	return fmt.Fprintf(w, "%s", t.Func(pattern, rest...))
}

func (t ValueSuffix) String() string {
	str := t.V.String()
	if len(str) > 0 {
		return str + t.Suffix
	}

	return ""
}

func (t ValueSuffix) Value() Value            { return t.V }
func (t ValueSuffix) Compare(other Value) int { panic("Never called") }
