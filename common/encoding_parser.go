package common

import (
	"bytes"
	"crypto/md5"
	"encoding/binary"
	"io"

	"github.com/pkg/errors"
)

type EncodingCPageEntry struct {
	KeyCount uint16
	FileSize uint32
	Ckey     []uint8   /*CHashSize*/
	Ekey     [][]uint8 /*[EHashSize][KeyCount]*/
}

type EncodingPageIndex struct {
	Hash     []uint8 /*XHashSize*/
	Checksum [0x10]uint8
}

type EncodingCTableEntry struct {
	Index   EncodingPageIndex
	Entries []EncodingCPageEntry
}

type EncodingHeader struct {
	Signature      uint16
	Version        uint8
	CHashSize      uint8
	EHashSize      uint8
	CPageSize      uint16
	EPageSize      uint16
	CPageCount     uint32
	EPageCount     uint32
	Unknown        uint8
	EspecBlockSize uint32
}

type Encoding struct {
	EncCTable []EncodingCTableEntry
}

func ParseEncoding(r io.Reader) (Encoding, error) {
	h := &EncodingHeader{}
	if err := binary.Read(r, binary.BigEndian, h); err != nil {
		return Encoding{}, errors.WithStack(err)
	}
	if h.Signature != 0x454e {
		return Encoding{}, errors.WithStack(errors.New("invalid encoding header"))
	}
	if _, err := io.ReadFull(r, make([]uint8, h.EspecBlockSize)); err != nil {
		return Encoding{}, errors.WithStack(err)
	}
	var cPageIndices []EncodingPageIndex
	for i := uint32(0); i < h.CPageCount; i++ {
		idx := EncodingPageIndex{}
		idx.Hash = make([]uint8, h.CHashSize)
		if err := binary.Read(r, binary.BigEndian, &idx.Hash); err != nil {
			return Encoding{}, errors.WithStack(err)
		}
		if err := binary.Read(r, binary.BigEndian, &idx.Checksum); err != nil {
			return Encoding{}, errors.WithStack(err)
		}
		cPageIndices = append(cPageIndices, idx)
	}
	encoding := Encoding{}
	for _, idx := range cPageIndices {
		CTableData := make([]byte, int(h.CPageSize)*1024)
		if err := binary.Read(r, binary.BigEndian, &CTableData); err != nil {
			return Encoding{}, errors.WithStack(err)
		}
		if hash := md5.Sum(CTableData); bytes.Compare(hash[:], idx.Checksum[:]) != 0 {
			return Encoding{}, errors.WithStack(errors.New("encoding file invalid checksum"))
		}
		entries := []EncodingCPageEntry{}
		CTableDataBuf := bytes.NewBuffer(CTableData)
		for i := uint32(0); ; /*until EOF or until padding (cEntry.KeyCount == 0)*/ i++ {
			cEntry := EncodingCPageEntry{}
			if err := binary.Read(CTableDataBuf, binary.LittleEndian, &cEntry.KeyCount); err != nil {
				if err == io.EOF {
					break
				}
				return Encoding{}, errors.WithStack(err)
			}
			if cEntry.KeyCount == 0 {
				//a page is zero padded once entries have filled it
				break
			}
			if err := binary.Read(CTableDataBuf, binary.BigEndian, &cEntry.FileSize); err != nil {
				return Encoding{}, errors.WithStack(err)
			}
			cEntry.Ckey = make([]uint8, h.CHashSize)
			if err := binary.Read(CTableDataBuf, binary.BigEndian, &cEntry.Ckey); err != nil {
				return Encoding{}, errors.WithStack(err)
			}
			for i := uint16(0); i < cEntry.KeyCount; i++ {
				ekey := make([]uint8, h.EHashSize)
				if err := binary.Read(CTableDataBuf, binary.BigEndian, &ekey); err != nil {
					return Encoding{}, errors.WithStack(err)
				}
				cEntry.Ekey = append(cEntry.Ekey, ekey)
			}
			entries = append(entries, cEntry)
		}
		encoding.EncCTable = append(encoding.EncCTable, EncodingCTableEntry{Index: idx, Entries: entries})
	}
	return encoding, nil
}
