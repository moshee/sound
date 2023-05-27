package flac

import (
	"bufio"
	"encoding/binary"
	"io"
	"time"

	"github.com/pkg/errors"

	"ktkr.us/pkg/sound"
	"ktkr.us/pkg/sound/vorbis"
)

const Magic = "fLaC"

func init() {
	sound.RegisterFormat("FLAC", "fLaC", Decode, DecodeTags, DecodeMeta)
}

type reader struct {
	r *bufio.Reader
}

func newReader(rr io.Reader) *reader {
	ret := &reader{}
	if br, ok := rr.(*bufio.Reader); ok {
		ret.r = br
	} else {
		ret.r = bufio.NewReader(rr)
	}
	ret.r.Discard(len(Magic))
	return ret
}

const (
	blockTypeStreaminfo = iota
	blockTypePadding
	blockTypeApplication
	blockTypeSeektable
	blockTypeVorbisComment
	blockTypeCuesheet
	blockTypePicture
	blockTypeInvalid = 127
)

type metadataBlock interface {
	header() metadataBlockHeader
}

type metadataBlockHeader struct {
	Header byte
	Length uint24
}

func (m metadataBlockHeader) header() metadataBlockHeader {
	return m
}

type streaminfo struct {
	MinBlockSize uint16
	MaxBlockSize uint16
	MinFrameSize uint24
	MaxFrameSize uint24
	SampleRate   uint64
	MD5          [16]byte
}

// info from STREAMINFO block to satisfy the sound.Metadata interface
type Metadata struct {
	MinBlockSize  uint16
	MaxBlockSize  uint16
	MinFrameSize  uint32
	MaxFrameSize  uint32
	sampleRate    int
	numChannels   int
	BitsPerSample int
	NumSamples    uint64
	MD5           [16]byte
}

func (m Metadata) Duration() time.Duration {
	sampleDurationSec := 1 / float64(m.sampleRate)
	durationSec := sampleDurationSec * float64(m.NumSamples)
	return time.Duration(durationSec * float64(time.Second))
}

func (m Metadata) NumChannels() int {
	return m.numChannels
}

func (m Metadata) BitRate() int {
	return m.BitsPerSample * m.sampleRate
}

func (m Metadata) SampleRate() int {
	return m.sampleRate
}

type uint24 [3]byte

func (n uint24) Uint32() uint32 {
	return uint32(n[0])<<16 | uint32(n[1])<<8 | uint32(n[2])
}

func Decode(rr io.Reader) (sound.Sound, error) {
	panic("x")
}

func DecodeTags(rr io.Reader) (sound.Tags, error) {
	var (
		lastMeta = false
		h        metadataBlockHeader
		comment  vorbis.Comment
	)

	r := newReader(rr)

	for !lastMeta {
		err := binary.Read(r.r, binary.BigEndian, &h)
		if err != nil {
			return nil, err
		}

		lastMeta = (h.Header>>7)&1 == 1
		blockType := h.Header & 0x7F
		blockSize := int(h.Length.Uint32())

		switch blockType {
		case blockTypeStreaminfo, blockTypePadding, blockTypeApplication, blockTypeSeektable, blockTypeCuesheet, blockTypePicture:
			// fmt.Printf("metadata block: %d (%d bytes)\n", blockType, blockSize)
			r.r.Discard(blockSize)

		case blockTypeVorbisComment:
			_, comment, err = vorbis.ReadComment(r.r)
			return comment, err

		case blockTypeInvalid:
			return nil, errors.New("invalid metadata block type")

		default:
			return nil, errors.Errorf("reserved metadata block type: %q", blockType)
		}
	}

	return vorbis.Comment{}, nil
}

func DecodeMeta(rr io.Reader, fsize int64) (sound.Metadata, error) {
	var (
		lastMeta = false
		h        metadataBlockHeader
	)

	r := newReader(rr)

	for !lastMeta {
		err := binary.Read(r.r, binary.BigEndian, &h)
		if err != nil {
			return nil, err
		}

		lastMeta = (h.Header>>7)&1 == 1
		blockType := h.Header & 0x7F
		blockSize := int(h.Length.Uint32())

		switch blockType {
		case blockTypeVorbisComment, blockTypePadding, blockTypeApplication, blockTypeSeektable, blockTypeCuesheet, blockTypePicture:
			// fmt.Printf("metadata block: %d (%d bytes)\n", blockType, blockSize)
			r.r.Discard(blockSize)

		case blockTypeStreaminfo:
			var b streaminfo
			err = binary.Read(r.r, binary.BigEndian, &b)
			if err != nil {
				return nil, err
			}

			sampleRate := int((b.SampleRate >> 44) & 0x3FFFF)
			numChannels := int((b.SampleRate>>41)&0x7) + 1
			bitsPerSample := int((b.SampleRate>>36)&0x1F) + 1
			numSamples := b.SampleRate & 0xFFFFFFFFF

			m := Metadata{
				MinBlockSize:  b.MinBlockSize,
				MaxBlockSize:  b.MaxBlockSize,
				MinFrameSize:  b.MinFrameSize.Uint32(),
				MaxFrameSize:  b.MaxFrameSize.Uint32(),
				sampleRate:    sampleRate,
				numChannels:   numChannels,
				BitsPerSample: bitsPerSample,
				NumSamples:    numSamples,
				MD5:           b.MD5,
			}
			return m, nil

		case blockTypeInvalid:
			return nil, errors.New("invalid metadata block type")

		default:
			return nil, errors.Errorf("reserved metadata block type: %q", blockType)
		}
	}

	return nil, errors.New("no STREAMINFO metadata block found")
}
