// Package id3v2 provides facilities for reading ID3v2 tags. Supported versions
// are 2.2, 2.3, and 2.4.
package id3v2

import (
	"bufio"
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"io"
	"io/ioutil"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/pkg/errors"

	"ktkr.us/pkg/sound"
)

// Tags is an ID3v2 tag set.
type Tags struct {
	*Header
	Frames map[string]string

	disc  int
	track int
	date  time.Time

	TotalTracks int
	TotalDiscs  int
}

func (t *Tags) Title() string       { return t.Frames["TIT2"] }
func (t *Tags) AlbumArtist() string { return t.Frames["TPE2"] }
func (t *Tags) Artist() string      { return t.Frames["TPE1"] }
func (t *Tags) Album() string       { return t.Frames["TALB"] }
func (t *Tags) Genre() string       { return t.Frames["TCON"] }
func (t *Tags) Disc() int           { return t.disc }
func (t *Tags) Track() int          { return t.track }
func (t *Tags) Date() time.Time     { return t.date }
func (t *Tags) Composer() string    { return t.Frames["TCOM"] }
func (t *Tags) Notes() string       { return t.Frames["COMM"] }

type Header struct {
	Magic [3]byte
	Major uint8
	Minor uint8
	Flags uint8
	Size  uint32
}

func synchsafe32(n uint32) uint32 {
	m := n & 0x7f
	m |= ((n & 0x7f00) >> 1)
	m |= ((n & 0x7f0000) >> 2)
	m |= ((n & 0x7f000000) >> 3)
	return m
}

type extHeader23 struct {
	Size    uint32
	Flags   uint16
	PadSize uint32
}
type extHeader24 struct {
	Size      uint32
	FlagBytes uint8
	Flags     byte
}

type frameHeader struct {
	Size  uint32
	Flags uint16
}

const Magic = "ID3"

const (
	// header flags
	flagUnsynchronisation = 1 << 7
	flagExtendedHeader    = 1 << 6
	flagExperimental      = 1 << 5
	flagFooterPresent     = 1 << 4

	// extended header flags
	extFlagTagIsUpdate     = 1 << 6
	extFlagCRCDataPresent  = 1 << 5
	extFlagTagRestrictions = 1 << 4

	// frame status flags
	frameTagAlterPreservation  = 1 << 14
	frameFileAlterPreservation = 1 << 13
	frameReadOnly              = 1 << 12

	// frame format flags
	frameGroupingIdentity    = 1 << 6
	frameCompressed          = 1 << 3
	frameEncrypted           = 1 << 2
	frameUnsynchronisation   = 1 << 1
	frameDataLengthIndicator = 1 << 0

	footerSize = 10
)

var (
	ErrBadHeader   = errors.New("id3v2: bad magic")
	ErrGarbage     = errors.New("id3v2: garbage in frame name")
	ErrUnknownFlag = errors.New("id3v2: unknown header flag")
	ErrEncryption  = errors.New("id3v2: frame encryption not supported")
	ErrCompression = errors.New("id3v2: frame compression not supported")
)

type countReader struct {
	r io.Reader
	n int64
}

func (r *countReader) Read(p []byte) (n int, err error) {
	n, err = io.ReadFull(r.r, p)
	r.n += int64(n)
	//log.Printf("read %d bytes", n)
	return
}

type limitedBufioReader struct {
	*bufio.Reader
	n int64
}

func (lpr *limitedBufioReader) Read(p []byte) (n int, err error) {
	if lpr.n <= 0 {
		return 0, io.EOF
	}
	if lpr.n >= int64(len(p)) {
		n, err = io.ReadFull(lpr.Reader, p)
	} else {
		n, err = io.ReadFull(lpr.Reader, p[:lpr.n])
	}
	lpr.n -= int64(n)
	return
}

// Decode decodes an ID3v2 header out of an MP3 stream. It only reads as many
// bytes as it needs to, no more and no less.
//
// The underlying type of the sound.Tags returned will be (*Tag).
func Decode(r io.Reader) (sound.Tags, error) {
	// log.Print("decode id3 header")
	//r := &countReader{r: rr}
	h, padding, err := readHeader(r)
	if err != nil {
		return nil, err
	}

	// log.Printf("%d bytes of padding", padding)

	// calls from mp3 package should always be bufio.Reader
	var br *bufio.Reader

	if _br, ok := r.(*bufio.Reader); ok {
		// log.Print("reader is *bufio.Reader")
		br = _br
	} else {
		br = bufio.NewReader(r)
	}

	// Make sure we don't read more than we need to, so that subsequent readers
	// can get the audio data from the file
	// totalSize := int64(h.Size)
	if h.Flags&flagFooterPresent != 0 {
		// we're not using the footer (which is used only to aid in searching
		// for the ID3 tags backwards from EOF), so discard it
		h.Size -= footerSize
		padding += footerSize
	}

	// Operate on entire tag in memory. Simplest way we can reliably limit
	// bytes read from *bufio.Reader (wrapping with limited reader will take
	// chunks at a time and not be accurate).
	tag := make([]byte, h.Size)
	io.ReadFull(br, tag)
	lr := bytes.NewReader(tag)
	// lr := &limitedBufioReader{br, int64(h.Size)}

	//log.Print("data left: ", lr.N)

	// log.Print("reading frames")
	frames, err := readFrames(lr, h)
	if err != nil {
		return nil, errors.Wrap(err, "read frames")
	}

	if padding > 0 {
		_, err = io.CopyN(ioutil.Discard, br, int64(padding))
		if err != nil {
			return nil, errors.Wrap(err, "discard padding")
		}
	}

	return makeTags(h, frames)
}

func readHeader(r io.Reader) (*Header, uint32, error) {
	var (
		esize   uint32
		h       Header
		padding uint32
	)

	err := binary.Read(r, binary.BigEndian, &h)
	if err != nil {
		return nil, 0, err
	}
	if string(h.Magic[:]) != Magic {
		return nil, 0, ErrBadHeader
	}
	h.Size = synchsafe32(h.Size)

	if (h.Flags & flagExtendedHeader) != 0 {
		switch h.Major {
		case 2:
			return nil, 0, ErrUnknownFlag

		case 3:
			var hh extHeader23
			err = binary.Read(r, binary.BigEndian, &hh)
			if err != nil {
				return nil, 0, err
			}
			if hh.Size > 6 {
				// discard CRC if present
				_, err = io.CopyN(ioutil.Discard, r, int64(hh.Size-6))
				if err != nil {
					return nil, 0, err
				}
			}
			// header size field in id3v2.3 doesn't include itself
			esize = hh.Size + 4
			h.Size -= hh.PadSize
			padding = hh.PadSize

		case 4:
			var hh extHeader24
			err = binary.Read(r, binary.BigEndian, &hh)
			if err != nil {
				return nil, 0, err
			}
			hh.Size = synchsafe32(hh.Size)

			////log.Printf("%#v", hh)

			// for now we're just gonna skip it
			// (the fixed header size is 6)
			buf := make([]byte, hh.Size-6)
			_, err = io.ReadFull(r, buf)
			if err != nil {
				return nil, 0, err
			}

			esize = hh.Size
		}
	}

	h.Size -= esize

	return &h, padding, nil
}

var validFramePat = regexp.MustCompile(`^[A-Z0-9]+\x00*$`)

func validFrameName(name []byte) bool {
	return validFramePat.Match(name)
	// apparently less than 4 (for ≥2.3) is ok if the end is zero padded
	// for _, c := range name {
	// 	if !(('A' <= c && c <= 'Z') || ('0' <= c && c <= '9')) {
	// 		return false
	// 	}
	// }
	// return true
}

func readFrames(rr *bytes.Reader, h *Header) (map[string]string, error) {
	var (
		frames     = make(map[string]string)
		txxx       = make(map[string]string)
		fh         frameHeader
		frameID    []byte
		headerSize uint32
		pos        = uint32(0)

		// needed to decode 3-byte size dscriptor in id3v2.2
		sizeBuf    = make([]byte, 4)
		frameSize  uint32
		allUnsynch = h.Flags&flagUnsynchronisation != 0
	)

	if h.Major == 2 {
		headerSize = 6
		frameID = make([]byte, 3)
	} else {
		headerSize = 10
		frameID = make([]byte, 4)
	}

frameloop:
	for ; pos < h.Size; pos += frameSize + headerSize {
		_, err := io.ReadFull(rr, frameID)
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}

		// next, err := rr.Peek(16)
		// if err != nil {
		// 	return nil, err
		// }
		// log.Printf("next 16: %q", next)

		// startPos := pos

		// seek through any padding etc that wasn't reported anywhere
		//XXX unfortunately peek doesn't advance the reader so we have to
		//manually implement a moving window to detect frames in a sea of nulls
		for !validFrameName(frameID) && pos < h.Size {
			b, err := rr.ReadByte()
			if err != nil {
				if err == io.EOF {
					break frameloop
				}
				return nil, err
			}

			copy(frameID[:len(frameID)-1], frameID[1:])
			frameID[len(frameID)-1] = b
			pos++
		}

		if pos >= h.Size {
			break
		}

		var (
			frameUnsynch = false
			// dataLengthIndicator uint32
			s           string
			frameIDStr            = string(frameID)
			frameReader io.Reader = rr
		)

		if h.Major == 2 {
			_, err = io.ReadFull(rr, sizeBuf[1:])
			if err != nil {
				return nil, err
			}

			frameSize = uint32(binary.BigEndian.Uint32(sizeBuf))
		} else {
			err = binary.Read(rr, binary.BigEndian, &fh)
			if err != nil {
				return nil, err
			}
			if h.Major >= 4 {
				frameSize = synchsafe32(fh.Size)
			} else {
				frameSize = fh.Size
			}

			if fh.Flags&frameEncrypted != 0 {
				return nil, ErrEncryption
			}

			frameUnsynch = allUnsynch || fh.Flags&frameUnsynchronisation != 0

			if fh.Flags&frameDataLengthIndicator != 0 {
				_, err = io.ReadFull(rr, sizeBuf)
				if err != nil {
					if err == io.EOF {
						return nil, errors.New("unexpected eof in frame header")
					}
					return nil, err
				}

				frameSize -= 4
				// dataLengthIndicator = synchsafe32(binary.BigEndian.Uint32(sizeBuf))
			}

			if fh.Flags&frameCompressed != 0 {
				zr, err := zlib.NewReader(rr)
				if err != nil {
					return nil, err
				}
				frameReader = zr
			}
		}
		// log.Printf("Frame %s (size %d/$%[2]x) at $%x", string(frameID), frameSize, pos)

		// if frameUnsynch {
		// 	log.Printf("frame %q is unsynchronised", frameIDStr)
		// }

		if frameID[0] == 'T' {
			buf := make([]byte, frameSize)
			_, err = io.ReadFull(rr, buf)
			if err != nil {
				return nil, err
			}

			if frameIDStr == "TXXX" {
				err = decodeTXXX(txxx, buf, frameUnsynch)
				if err != nil {
					log.Print(errors.Wrap(err, "decode TXXX"))
				} else {
					// log.Printf("  %q", truncate(s, 40))
				}
				continue
			}

			s, err = decodeTextFrame(buf[0], buf[1:], frameUnsynch)
			if err != nil {
				return nil, err
			}

			j := strings.IndexByte(s, '\x00')
			if j > -1 {
				s = s[:j]
			}
		} else {
			switch frameIDStr {
			case "APIC", "PIC", "PRIV":
				// skip album arts for now
				//log.Print("skipping album art")

				io.CopyN(ioutil.Discard, rr, int64(frameSize))
				continue

			case "COMM":
				buf := make([]byte, frameSize)
				_, err = io.ReadFull(rr, buf)
				if err != nil {
					return nil, err
				}

				b := bytes.NewBuffer(buf)
				enc, err := b.ReadByte()
				if err != nil {
					return nil, err
				}

				b.Next(3) // discard lang code
				// log.Printf("comment lang: %q", lang)
				readTerminatedString(enc, b)
				s, err = decodeTextFrame(enc, b.Bytes(), frameUnsynch)
				if err != nil {
					return nil, err
				}

			default:
				buf := make([]byte, frameSize)
				_, err = io.ReadFull(rr, buf)
				if err != nil {
					return nil, err
				}
				// TODO: other special frames
				s = string(buf)
			}
		}

		// log.Printf("  %q", truncate(s, 40))

		if len(frameID) == 3 {
			newID, ok := v22Equiv[frameIDStr]
			if ok {
				//log.Printf("%s => %s", frameIDStr, newID)
				frameIDStr = newID
			}
		}

		//log.Printf("%s: %s", frameIDStr, s)
		frames[frameIDStr] = s

		if zr, ok := frameReader.(io.ReadCloser); ok {
			zr.Close()
		}
	}

	// for k, v := range frames {
	// 	log.Printf("%q: %q", k, v)
	// }
	translateTXXXFrames(frames, txxx)

	return frames, nil
}

func truncate(s string, limit int) string {
	if len(s) <= limit {
		return s
	} else {
		half := len(s) / 2
		remove := (len(s) - limit) / 2
		end := half - remove
		start := half + remove

		return s[:end] + "..." + s[start:]
	}
}

func makeTags(h *Header, frames map[string]string) (sound.Tags, error) {
	t := Tags{
		Header: h,
		Frames: frames,
	}
	var err error

	TPOS := t.Frames["TPOS"]
	if TPOS != "" {
		t.disc, t.TotalDiscs, err = parseMultiNumber(TPOS)
		if err != nil {
			return nil, err
		}
	}

	TRCK := t.Frames["TRCK"]
	if TRCK != "" {
		t.track, t.TotalTracks, err = parseMultiNumber(TRCK)
		if err != nil {
			return nil, err
		}
	}

	t.date, err = parseDate(t.Frames)
	if err != nil {
		return nil, err
	}

	return &t, nil
}
