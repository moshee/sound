package mp3

import (
	"bufio"
	"encoding/binary"
	"io"
)

type reader struct {
	r   *bufio.Reader
	buf []byte
}

func newReader(r io.Reader) *reader {
	return &reader{r: ensureBufioReader(r)}
}

func (r *reader) nextFrame() (*frame, error) {
	var err error
	// log.Printf("decoding mp3 at %x", z)
	// log.Printf("%d buffered", r.r.Buffered())
	//discard := make([]byte, 32)
	for i := 0; ; i++ {
		x, err := r.r.Peek(2)
		if err != nil {
			return nil, err
		}
		if x[0] == 0xFF && x[1]&0xE0 == 0xE0 {
			// log.Print("skipped ", i, " bytes")
			break
		}
		r.r.ReadByte()
		//r.r.Read(discard[:1])
	}
	var header uint32

	err = binary.Read(r.r, binary.BigEndian, &header)
	if err != nil {
		//print("C")
		return nil, err
	}
	// log.Printf("%x", header)

	// frame sync
	if (header >> 21) != 0x7FF {
		return nil, ErrUnsynced
	}
	var (
		mpegVersion = int(header>>19) & 0x3
		layer       = int(header>>17) & 0x3
	)

	if mpegVersion == versionReserved || layer == layerReserved {
		return nil, ErrReserved
	}

	var (
		bitrate    = int(header>>12) & 0xF
		samplerate = int(header>>10) & 0x3
	)

	h := frameHeader{
		mpegVersion: mpegVersion,
		layer:       layer,
		haveCRC:     ((header >> 16) & 0x1) == 0,
		bitrate:     bitrates[mpegVersion][layer][bitrate] * 1000,
		samplerate:  sampleRates[mpegVersion][samplerate],
		channelMode: int(header>>6) & 0x3,
		havePadding: ((header >> 9) & 0x1) == 1,
	}

	// bit 8: private
	//hModeExtension = (header >> 4) & 0x3
	//hIsCopyrighted = (header >> 3) & 0x1
	//hIsOriginal    = (header >> 2) & 0x1
	//hEmphasis      = header & 0x3
	// ^ we're not going to care about these for now

	if h.bitrate < 0 {
		return nil, ErrBadBitrate
	}
	if h.samplerate < 0 {
		return nil, ErrBadSampleRate
	}

	spf := samplesPerFrame[mpegVersion][layer]
	h.frameSize = ((spf * h.bitrate / 8) / h.samplerate)
	if h.havePadding {
		h.frameSize++
	}

	h.frameSize -= 4

	// Check CRC (TODO: actually do this)
	if h.haveCRC {
		var crc uint16
		// CRC-16 uses the IBM (ANSI, Modbus) polynomial
		err = binary.Read(r.r, binary.BigEndian, &crc)
		if err != nil {
			//print(1)
			return nil, err
		}
		//println(crc)
	}

	var sideInfoSize int
	if h.mpegVersion == version1 {
		if h.channelMode == channelMono {
			sideInfoSize = 17
		} else {
			sideInfoSize = 32
		}
	} else {
		if h.channelMode == channelMono {
			sideInfoSize = 9
		} else {
			sideInfoSize = 17
		}
	}
	if h.haveCRC {
		sideInfoSize -= 2
	}
	// log.Print("side info size: ", sideInfoSize)
	h.frameSize -= sideInfoSize
	/*
		var buf []byte
		if r.buf == nil || len(r.buf) < sideInfoSize {
			r.buf = make([]byte, uint(sideInfoSize))
			buf = r.buf
		} else {
			buf = r.buf[:sideInfoSize]
		}
		io.ReadFull(r.r, buf)
	*/
	r.r.Discard(sideInfoSize)
	//r.r.Read(discard[:sideInfoSize])

	//log.Printf("%#v", h)
	//log.Print(r.r.Peek(4))
	return &frame{h, &frameData{
		io.LimitedReader{r.r, int64(h.frameSize)},
	}}, nil
}
