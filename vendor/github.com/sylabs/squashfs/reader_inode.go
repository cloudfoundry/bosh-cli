package squashfs

import (
	"errors"
	"io"

	"github.com/sylabs/squashfs/internal/data"
	"github.com/sylabs/squashfs/internal/directory"
	"github.com/sylabs/squashfs/internal/inode"
	"github.com/sylabs/squashfs/internal/metadata"
	"github.com/sylabs/squashfs/internal/toreader"
)

func (r Reader) inodeFromRef(ref uint64) (i inode.Inode, err error) {
	offset, meta := (ref>>16)+r.s.InodeTableStart, ref&0xFFFF
	rdr := metadata.NewReader(toreader.NewReader(r.r, int64(offset)), r.d)
	_, err = rdr.Read(make([]byte, meta))
	if err != nil {
		return
	}
	return inode.Read(rdr, r.s.BlockSize)
}

func (r Reader) inodeFromDir(e directory.Entry) (i inode.Inode, err error) {
	rdr := metadata.NewReader(toreader.NewReader(r.r, int64(uint64(e.BlockStart)+r.s.InodeTableStart)), r.d)
	_, err = rdr.Read(make([]byte, e.Offset))
	if err != nil {
		return
	}
	return inode.Read(rdr, r.s.BlockSize)
}

func (r Reader) getReaders(i inode.Inode) (full *data.FullReader, rdr *data.Reader, err error) {
	var fragOffset uint64
	var blockOffset uint64
	var blockSizes []uint32
	var fragInd uint32
	var fragSize uint32
	if i.Type == inode.Fil {
		fragOffset = uint64(i.Data.(inode.File).FragOffset)
		blockOffset = uint64(i.Data.(inode.File).BlockStart)
		blockSizes = i.Data.(inode.File).BlockSizes
		fragInd = i.Data.(inode.File).FragInd
		fragSize = i.Data.(inode.File).Size % r.s.BlockSize
	} else if i.Type == inode.EFil {
		fragOffset = uint64(i.Data.(inode.EFile).FragOffset)
		blockOffset = i.Data.(inode.EFile).BlockStart
		blockSizes = i.Data.(inode.EFile).BlockSizes
		fragInd = i.Data.(inode.EFile).FragInd
		fragSize = uint32(i.Data.(inode.EFile).Size % uint64(r.s.BlockSize))
	} else {
		return nil, nil, errors.New("getReaders called on non-file type")
	}
	rdr = data.NewReader(toreader.NewReader(r.r, int64(blockOffset)), r.d, blockSizes, r.s.BlockSize)
	full = data.NewFullReader(r.r, uint64(blockOffset), r.d, blockSizes, r.s.BlockSize)
	if fragInd != 0xFFFFFFFF {
		full.AddFragment(func() (io.Reader, error) {
			var fragRdr io.Reader
			fragRdr, err = r.fragReader(fragInd)
			if err != nil {
				return nil, err
			}
			var n, tmpN int
			for n != int(fragOffset) {
				tmpN, err = fragRdr.Read(make([]byte, int(fragOffset)-n))
				if err != nil {
					return nil, err
				}
				n += tmpN
			}
			fragRdr = io.LimitReader(fragRdr, int64(fragSize))
			return fragRdr, nil
		})
		var fragRdr io.Reader
		fragRdr, err = r.fragReader(fragInd)
		if err != nil {
			return nil, nil, err
		}
		var n, tmpN int
		for n != int(fragOffset) {
			tmpN, err = fragRdr.Read(make([]byte, int(fragOffset)-n))
			if err != nil {
				return nil, nil, err
			}
			n += tmpN
		}
		fragRdr = io.LimitReader(fragRdr, int64(fragSize))
		rdr.AddFragment(fragRdr)
	}
	return
}

func (r Reader) readDirectory(i inode.Inode) ([]directory.Entry, error) {
	var offset uint64
	var blockOffset uint16
	var size uint32
	if i.Type == inode.Dir {
		offset = uint64(i.Data.(inode.Directory).BlockStart)
		blockOffset = i.Data.(inode.Directory).Offset
		size = uint32(i.Data.(inode.Directory).Size)
	} else if i.Type == inode.EDir {
		offset = uint64(i.Data.(inode.EDirectory).BlockStart)
		blockOffset = i.Data.(inode.EDirectory).Offset
		size = i.Data.(inode.EDirectory).Size
	} else {
		return nil, errors.New("readDirectory called on non-directory type")
	}
	rdr := metadata.NewReader(toreader.NewReader(r.r, int64(offset+r.s.DirTableStart)), r.d)
	_, err := rdr.Read(make([]byte, blockOffset))
	if err != nil {
		return nil, err
	}
	return directory.ReadEntries(rdr, size)
}
