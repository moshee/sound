// Package sound implements routines for decoding audio files.
//
// Currently, this package only aims to provide functionality for metadata such
// as tags.
package sound

import (
	"bufio"
	"errors"
	"io"
	"os"
	"time"
)

var (
	ErrFormat = errors.New("sound: unknown format")
)

// TODO: album art?

type Sound interface{}

type Metadata interface {
	Duration() time.Duration
	NumChannels() int // Number of audio channels.
	// Number of bits per second. This is more of an advisory value than a hard
	// number, but it should be correct for CBR.
	BitRate() int
	SampleRate() int // Number of samples per second.
}

type Tags interface {
	Title() string
	AlbumArtist() string
	Artist() string
	Album() string
	Genre() string
	Disc() int
	Track() int
	Date() time.Time
	Composer() string
	Notes() string
}

var formats []format

type format struct {
	name       string
	magic      string
	decode     func(io.Reader) (Sound, error)
	decodeTags func(io.Reader) (Tags, error)
	decodeMeta func(io.Reader, int64) (Metadata, error)
}

// RegisterFormat lets the package know how to decode a sound file format
// identified by a magic number. Each decode function is provided with
// the the input as an io.Reader. In addition, the decodeMeta function
// will get the filesize as an additional parameter to aid in calculating
// duration, if the filesize can be calculated.
func RegisterFormat(name, magic string,
	decode func(io.Reader) (Sound, error),
	decodeTags func(io.Reader) (Tags, error),
	decodeMeta func(io.Reader, int64) (Metadata, error)) {
	formats = append(formats, format{name, magic, decode, decodeTags, decodeMeta})
}

func Decode(r io.Reader) (Sound, string, error) {
	panic("unimplemented")
}

func DecodeMeta(r io.Reader) (Metadata, string, error) {
	rr := bufio.NewReader(r)

	f := sniff(rr)
	if f.decodeMeta == nil {
		return nil, "", ErrFormat
	}
	var (
		n   int64
		err error
	)
	if seeker, ok := r.(io.Seeker); ok {
		n, err = seeker.Seek(0, os.SEEK_END)
		if err != nil {
			return nil, f.name, err
		}
		seeker.Seek(0, os.SEEK_SET)
		rr.Reset(r)
	} else {
		// how else can we determine the file size easily?
	}

	m, err := f.decodeMeta(rr, n)
	return m, f.name, err
}

func DecodeTags(r io.Reader) (Tags, string, error) {
	rr := bufio.NewReader(r)
	f := sniff(rr)
	if f.decodeTags == nil {
		return nil, "", ErrFormat
	}
	m, err := f.decodeTags(rr)
	return m, f.name, err
}

// Match reports whether magic matches b. Magic may contain "?" wildcards.
func match(magic string, b []byte) bool {
	if len(magic) != len(b) {
		return false
	}
	for i, c := range b {
		if magic[i] != c && magic[i] != '?' {
			return false
		}
	}
	return true
}

// Sniff determines the format of r's data.
func sniff(r *bufio.Reader) format {
	for _, f := range formats {
		b, err := r.Peek(len(f.magic))
		if err == nil && match(f.magic, b) {
			return f
		}
	}
	return format{}
}
