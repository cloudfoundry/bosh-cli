package ui

import (
	"io"
	"strings"
	"sync"
)

type ComboWriter struct {
	ui     UI
	uiLock sync.Mutex
	onNL   bool
}

type prefixedWriter struct {
	w      *ComboWriter
	prefix string
}

func NewComboWriter(ui UI) *ComboWriter {
	return &ComboWriter{ui: ui, onNL: true}
}

func (w *ComboWriter) Writer(prefix string) io.Writer {
	return prefixedWriter{w: w, prefix: prefix}
}

func (s prefixedWriter) Write(bytes []byte) (int, error) {
	s.w.uiLock.Lock()
	defer s.w.uiLock.Unlock()

	pieces := strings.Split(string(bytes), "\n")

	for i, piece := range pieces {
		if i == len(pieces)-1 {
			if s.w.onNL {
				if len(piece) > 0 {
					s.w.ui.PrintBlock(s.prefix)
					s.w.ui.PrintBlock(piece)
					s.w.onNL = false
				}
			} else {
				s.w.ui.PrintBlock(piece)
			}
		} else {
			if s.w.onNL {
				s.w.ui.PrintBlock(s.prefix)
			}
			s.w.ui.PrintBlock(piece)
			s.w.ui.PrintBlock("\n")
			s.w.onNL = true
		}
	}

	return len(bytes), nil
}
