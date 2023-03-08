package inode

import (
	"encoding/binary"
	"errors"
	"io"
	"strconv"
)

const (
	Dir = uint16(iota + 1)
	Fil
	Sym
	Block
	Char
	Fifo
	Sock
	EDir
	EFil
	ESym
	EBlock
	EChar
	EFifo
	ESock
)

type Header struct {
	Type    uint16
	Perm    uint16
	UidInd  uint16
	GidInd  uint16
	ModTime uint32
	Num     uint32
}

type Inode struct {
	Header
	Data any
}

func Read(r io.Reader, blockSize uint32) (i Inode, err error) {
	err = binary.Read(r, binary.LittleEndian, &i.Header)
	if err != nil {
		return
	}
	switch i.Type {
	case Dir:
		i.Data, err = ReadDir(r)
	case Fil:
		i.Data, err = ReadFile(r, blockSize)
	case Sym:
		i.Data, err = ReadSym(r)
	case Block:
		fallthrough
	case Char:
		i.Data, err = ReadDevice(r)
	case Fifo:
		fallthrough
	case Sock:
		i.Data, err = ReadIPC(r)
	case EDir:
		i.Data, err = ReadEDir(r)
	case EFil:
		i.Data, err = ReadEFile(r, blockSize)
	case ESym:
		i.Data, err = ReadESym(r)
	case EBlock:
		fallthrough
	case EChar:
		i.Data, err = ReadEDevice(r)
	case EFifo:
		fallthrough
	case ESock:
		i.Data, err = ReadEIPC(r)
	default:
		return i, errors.New("invalid inode type " + strconv.Itoa(int(i.Type)))
	}
	return
}
