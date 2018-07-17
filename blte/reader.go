package blte

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"fmt"
	"io"
	"io/ioutil"
)

type header struct {
	Sig  uint32
	Size uint32
}

type chunkInfo struct {
	Unknown uint16
	Count   uint16
}

type chunkInfoEntry struct {
	Csize    uint32
	USize    uint32
	Checksum [16]uint8
}

func Decode(r io.Reader, w io.Writer) error {
	h := header{}
	if err := binary.Read(r, binary.BigEndian, &h); err != nil {
		return err
	}

	if h.Sig != 0x424c5445 /*BLTE*/ {
		return fmt.Errorf("invalid signature %x", h.Sig)
	}

	if h.Size == 0 {
		all, err := ioutil.ReadAll(r)
		if err != nil {
			return err
		}
		l := uint32(len(all))
		return decodeData(bytes.NewBuffer(all), w, l, l)
	}

	ci := chunkInfo{}
	if err := binary.Read(r, binary.BigEndian, &ci); err != nil {
		return err
	}

	entries := []chunkInfoEntry{}
	for i := uint16(0); i < uint16(ci.Count); i++ {
		entry := chunkInfoEntry{}
		if err := binary.Read(r, binary.BigEndian, &entry); err != nil {
			return err
		}
		entries = append(entries, entry)
	}

	for _, e := range entries {
		if err := decodeData(r, w, e.Csize-1, e.USize); err != nil {
			return err
		}
	}
	return nil
}

func decodeData(r io.Reader, w io.Writer, csize, usize uint32) error {
	var typ uint8
	if err := binary.Read(r, binary.BigEndian, &typ); err != nil {
		return err
	}

	if typ == 'N' {
		buf := make([]uint8, int(usize))
		if err := binary.Read(r, binary.BigEndian, &buf); err != nil {
			return err
		}
		_, err := w.Write(buf)
		return err
	}

	if typ == 'Z' {
		tmp := make([]uint8, int(csize))
		if _, err := io.ReadFull(r, tmp); err != nil {
			return err
		}

		z, err := zlib.NewReader(bytes.NewBuffer(tmp))
		if err != nil {
			return err
		}

		utmp := make([]uint8, int(usize))
		if _, err = io.ReadFull(z, utmp); err != nil {
			return err
		}
		_, err = w.Write(utmp)
		return err
	}

	return fmt.Errorf("unsuported encoding type %+q", typ)
}