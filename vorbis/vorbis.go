package vorbis

import (
	"encoding/binary"
	"errors"
	"io"
	"log"
	"strconv"
	"strings"
	"time"

	"ktkr.us/pkg/sound"
	"ktkr.us/pkg/sound/ogg"
)

func init() {
	sound.RegisterFormat("Ogg Vorbis", "OggS????????????????????????\x01vorbis", Decode, DecodeTags, DecodeMeta)
}

const (
	idPreamble      = "\x01vorbis"
	commentPreamble = "\x03vorbis"
	setupPreamble   = "\x05vorbis"
)

var (
	ErrMissingFramingBit = errors.New("vorbis: missing framing bit")
	ErrBadPreamble       = errors.New("vorbis: malformed packet preamble")
	ErrBadComment        = errors.New("vorbis: malformed comment vector")
)

/*
Vorbis I Spec §4.2.2
1) [vorbis_version]    = read 32 bits as unsigned integer
2) [audio_channels]    = read 8 bit integer as unsigned
3) [audio_sample_rate] = read 32 bits as unsigned integer
4) [bitrate_maximum]   = read 32 bits as signed integer
5) [bitrate_nominal]   = read 32 bits as signed integer
6) [bitrate_minimum]   = read 32 bits as signed integer
7) [blocksize_0]       = 2 exponent (read 4 bits as unsigned integer)
8) [blocksize_1]       = 2 exponent (read 4 bits as unsigned integer)
9) [framing_flag]      = read one bit
*/
type header struct {
	VorbisVersion   uint32
	AudioChannels   uint8
	AudioSampleRate uint32
	BitrateMaximum  int32
	BitrateNominal  int32
	BitrateMinimum  int32
	BlockSizes      uint8
	FramingBit      uint8
}

type meta struct {
	header
	numSamples int64
	Comment
}

func (m *meta) Duration() time.Duration {
	// Avoid overflowing int64 to get milliseconds if we have a really really
	// long track
	return time.Millisecond * time.Duration(1e3*float64(m.numSamples)/float64(m.AudioSampleRate))
}

func (m *meta) NumChannels() int {
	return int(m.AudioChannels)
}

func (m *meta) BitRate() int {
	if m.BitrateNominal == 0 {
		// approximate using file size and duration
		return 0
	}
	return int(m.BitrateNominal)
}

func (m *meta) SampleRate() int {
	return int(m.AudioSampleRate)
}

func Decode(rr io.Reader) (sound.Sound, error) {
	return nil, nil
}

func DecodeTags(rr io.Reader) (sound.Tags, error) {
	r := ogg.NewReader(rr)
	r.NextPage()
	r.NextPage()

	err := readPacketPreamble(r, commentPreamble)
	if err != nil {
		return nil, errors.New("malformed Vorbis Comment preamble")
	}
	_, comment, err := ReadComment(r)
	return comment, err
}

func DecodeMeta(rr io.Reader, fsize int64) (sound.Metadata, error) {
	r := ogg.NewReader(rr)
	err := readPacketPreamble(r, idPreamble)
	if err != nil {
		return nil, err
	}

	var h header
	err = binary.Read(r, binary.LittleEndian, &h)
	if err != nil {
		return nil, err
	}

	if h.FramingBit != 1 {
		return nil, ErrMissingFramingBit
	}

	err = readPacketPreamble(r, commentPreamble)
	if err != nil {
		return nil, errors.New("malformed Vorbis Comment preamble")
	}
	_, comment, err := ReadComment(r)
	if err != nil {
		return nil, err
	}

	var page, lastPage *ogg.Page
	for {
		page, err = r.NextPage()
		if err != nil {
			return nil, err
		}

		if page == nil {
			break
		}

		lastPage = page
	}

	return &meta{h, lastPage.GranulePos, comment}, nil
}

func decode(r io.Reader) (sound.Sound, error) {
	return nil, nil
}

func decodeMeta(r io.Reader) (sound.Metadata, error) {
	panic("aaa")
}

func ReadComment(r io.Reader) (string, Comment, error) {
	vendor, err := readString(r)
	if err != nil {
		return "", nil, err
	}

	var numComments uint32
	err = binary.Read(r, binary.LittleEndian, &numComments)
	if err != nil {
		return "", nil, err
	}

	c := make(Comment, numComments)

	for i := uint32(0); i < numComments; i++ {
		comment, err := readString(r)
		if err != nil {
			return "", nil, err
		}

		parts := strings.SplitN(comment, "=", 2)
		if len(parts) < 2 {
			return "", nil, ErrBadComment
		}
		key := strings.ToUpper(parts[0])

		// again, we're gonna skip album art for now
		if key == "METADATA_BLOCK_PICTURE" {
			continue
		}
		val := parts[1]

		if _, ok := c[key]; ok {
			c[key] = append(c[key], val)
		} else {
			c[key] = []string{val}
		}
	}

	return vendor, c, nil
}

func readPacketPreamble(r io.Reader, preamble string) error {
	buf := make([]byte, len(preamble))
	_, err := r.Read(buf)
	if err != nil {
		return err
	}
	if string(buf) != preamble {
		log.Printf("%q", string(buf))
		return ErrBadPreamble
	}
	return nil
}

func readString(r io.Reader) (string, error) {
	var length uint32
	err := binary.Read(r, binary.LittleEndian, &length)
	if err != nil {
		return "", err
	}

	s := make([]byte, length)
	_, err = io.ReadFull(r, s)
	if err != nil {
		return "", err
	}

	return string(s), nil
}

type Comment map[string][]string

func (c Comment) Get(key string) string {
	val := c[key]
	if val != nil && len(val) > 0 {
		return val[0]
	}
	return ""
}

func (c Comment) GetAll(key string) string {
	val, ok := c[key]
	if ok {
		return strings.Join(val, ", ")
	}
	return ""
}

var dateFormats = []string{
	"2006-01-02",
	"2006-01",
	"2006",
}

func (c Comment) Title() string       { return c.GetAll("TITLE") }
func (c Comment) AlbumArtist() string { return c.GetAll("ALBUMARTIST") }
func (c Comment) Artist() string      { return c.GetAll("ARTIST") }
func (c Comment) Album() string       { return c.GetAll("ALBUM") }
func (c Comment) Genre() string       { return c.GetAll("GENRE") }
func (c Comment) Composer() string    { return c.GetAll("COMPOSER") }
func (c Comment) Notes() string       { return c.Get("DESCRIPTION") }

func (c Comment) Disc() int {
	n, _ := strconv.Atoi(c.Get("DISCNUMBER"))
	return n
}

func (c Comment) Track() int {
	n, _ := strconv.Atoi(c.Get("TRACKNUMBER"))
	return n
}

func (c Comment) Date() time.Time {
	s := c.Get("DATE")
	for _, dateFormat := range dateFormats {
		t, err := time.Parse(dateFormat, s)
		if err != nil {
			continue
		}
		return t
	}
	return time.Time{}
}
