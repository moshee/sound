package mp3

import (
	"encoding/binary"
	"io"
)

type Xing struct {
	NumFrames    uint32
	NumFileBytes uint32
	TOC          []byte
	Quality      uint32
}

const (
	xingFrames = 1 << iota
	xingBytes
	xingTOC
	xingQuality
)

func decodeXing(r io.Reader) (*Xing, error) {
	// log.Print("decoding xing")
	var flags uint32
	err := binary.Read(r, binary.BigEndian, &flags)
	if err != nil {
		return nil, err
	}
	var xing Xing
	if flags&xingFrames != 0 {
		err = binary.Read(r, binary.BigEndian, &xing.NumFrames)
		if err != nil {
			return nil, err
		}
	}
	if flags&xingBytes != 0 {
		err = binary.Read(r, binary.BigEndian, &xing.NumFileBytes)
		if err != nil {
			return nil, err
		}
	}
	if flags&xingTOC != 0 {
		xing.TOC = make([]byte, 100)
		_, err = io.ReadFull(r, xing.TOC)
		if err != nil {
			return nil, err
		}
	}
	if flags&xingQuality != 0 {
		err = binary.Read(r, binary.BigEndian, &xing.Quality)
		if err != nil {
			return nil, err
		}
	}
	// log.Printf("%#v", xing)

	return &xing, nil
}

type VBRI struct {
	Version      uint16
	Delay        uint16
	Quality      uint16
	NumBytes     uint32
	NumFrames    uint32
	TOCSize      uint16
	TOCScale     uint16
	TOCEntrySize uint16
}

func decodeVBRI(r io.Reader) (*VBRI, error) {
	var vbri VBRI
	err := binary.Read(r, binary.BigEndian, &vbri)
	// there is still some TOC data left but whatever, I don't even know what
	// that is and it won't help
	return &vbri, err
}
