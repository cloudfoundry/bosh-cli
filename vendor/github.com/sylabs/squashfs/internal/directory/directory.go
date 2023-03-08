package directory

import (
	"bytes"
	"encoding/binary"
	"io"
)

type header struct {
	Entries    uint32
	InodeStart uint32
	Num        uint32
}

type entryInit struct {
	Offset    uint16
	NumOffset int16
	Type      uint16
	NameSize  uint16
}

type entry struct {
	entryInit
	Name []byte
}

type Entry struct {
	Name       string
	BlockStart uint32
	Type       uint16
	Offset     uint16
}

func readEntry(r io.Reader) (e entry, err error) {
	err = binary.Read(r, binary.LittleEndian, &e.entryInit)
	if err != nil {
		return
	}
	e.Name = make([]byte, e.NameSize+1)
	err = binary.Read(r, binary.LittleEndian, &e.Name)
	return
}

func ReadEntries(rdr io.Reader, size uint32) (e []Entry, err error) {
	dat := make([]byte, size-3)
	rdr.Read(dat)
	r := bytes.NewReader(dat)
	var h header
	var en entry
	for {
		err = binary.Read(r, binary.LittleEndian, &h)
		if err == io.EOF {
			err = nil
			return
		} else if err != nil {
			return
		}
		h.Entries++
		for i := 0; i < int(h.Entries); i++ {
			if i != 0 && i%256 == 0 {
				err = binary.Read(r, binary.LittleEndian, &h)
				if err != nil {
					return
				}
			}
			en, err = readEntry(r)
			if err != nil {
				return
			}
			e = append(e, Entry{
				Name:       string(en.Name),
				BlockStart: h.InodeStart,
				Type:       en.Type,
				Offset:     en.Offset,
			})
		}
	}
}
