// Package ogg implements the Ogg multimedia container format, used in media
// formats such as Vorbis, Theora, Opus, and Speex.
//
// An Ogg file is comprised of a sequence of pages, generally around 8k each in
// size, to facilitate easy stream syncing. Each page has a header with
// position and size information and a data segment.
//
// Pages are a transport stream for abstract data packets. Packets may span
// multiple pages. Each contained format has its own specifications on how
// packets are defined.
//
// Detailed information on the format can be found at
// http://www.xiph.org/ogg/
package ogg

import (
	"bufio"
	"encoding/binary"
	"errors"
	"hash/crc32"
	"io"
)

const (
	CapturePattern  = "OggS"
	CRC32Polynomial = 0x04c11db7
)

var (
	ErrBadHeader = errors.New("ogg: malformed header")
	crcTable     = crc32.MakeTable(CRC32Polynomial)
)

type Header struct {
	Version            byte
	HeaderType         byte
	GranulePos         int64
	StreamSerialNumber uint32
	PageCounter        uint32
	PageChecksum       uint32
	SegmentCount       uint8
}

const (
	headerTypeContinued = 1 << 1
	headerTypeBOS       = 1 << 2
	headerTypeEOS       = 1 << 3
)

type Page struct {
	Header
	// Segments is a table of data segments pointing into the Data field.
	//Segments [][]byte
	// Data is the raw page data.
	Data []byte
}

// Decode returns an io.Reader that provides raw data, transparently decoding
// pages internally. It will be of type *Reader.
func Decode(r io.Reader) io.Reader {
	return NewReader(r)
}

type Reader struct {
	r *bufio.Reader
	// page is the current page
	validPage  bool
	page       Page
	ptr        int
	buf        []byte
	segmentTab []byte
	h          Header
}

func NewReader(r io.Reader) *Reader {
	if br, ok := r.(*bufio.Reader); ok {
		return &Reader{r: br}
	}
	return &Reader{r: bufio.NewReader(r)}
}

// Read reads raw packet data into a buffer, decoding packets across pages as
// needed.
//
// Read uses ReadPage but does not buffer pages. This means that reading a page
// using NextPage will cause Read to lose its position in the stream.
// Subsequent calls to Read will begin at the start of the last page read.
func (r *Reader) Read(p []byte) (n int, err error) {
	for n < len(p) {
		// Decode a new page from the stream if this is the first read or we've
		// exhausted the current page.
		if !r.validPage || r.ptr >= len(r.page.Data)-1 {
			var page *Page
			page, err = r.NextPage()
			if err != nil {
				if page == nil {
					return n, io.EOF
				}
				return
			}
		}

		remainingInPage := len(r.page.Data) - r.ptr
		bytesToCopy := len(p)
		var m int
		if remainingInPage < bytesToCopy {
			m = copy(p[n:], r.page.Data[r.ptr:])
		} else {
			m = copy(p[n:], r.page.Data[r.ptr:r.ptr+bytesToCopy])
		}
		n += m
		r.ptr += m
	}
	return
}

// NextPage decodes and returns a page from the Ogg stream, and any decoding
// error occurred. It only returns an EOF error if an unexpected EOF occurred
// in the middle of a page. If the last page has already been read out, both
// return values will be nil.
//
// Page data is only valid until the next call to NextPage.
func (r *Reader) NextPage() (*Page, error) {
	// should seek until we find an OggS
	err := r.capture(false)
	if err != nil {
		if err == io.EOF {
			return nil, nil
		}
		r.validPage = false
		return nil, err
	}

	err = binary.Read(r.r, binary.LittleEndian, &r.page.Header)
	if err != nil {
		r.validPage = false
		return nil, err
	}

	var segmentTab []byte
	if r.segmentTab == nil || len(r.segmentTab) < int(r.page.Header.SegmentCount) {
		r.segmentTab = make([]byte, r.page.Header.SegmentCount)
		segmentTab = r.segmentTab
	} else {
		segmentTab = r.segmentTab[:r.page.Header.SegmentCount]
	}
	_, err = io.ReadFull(r.r, segmentTab)
	if err != nil {
		r.validPage = false
		return nil, err
	}

	pageSize := 0
	for _, l := range segmentTab {
		pageSize += int(l)
	}

	// Only allocate a new buffer if we need more space
	var buf []byte
	if r.buf == nil || len(r.buf) < pageSize {
		r.buf = make([]byte, uint(pageSize))
		buf = r.buf
	} else {
		buf = r.buf[:pageSize]
	}
	_, err = io.ReadFull(r.r, buf)
	if err != nil {
		r.validPage = false
		return nil, err
	}

	/*
		segments := make([][]byte, h.SegmentCount)
		pos := 0
		for i := range segments {
			segmentSize := int(segmentTab[i])
			segments[i] = r.buf[pos : pos+segmentSize]
			pos += segmentSize
		}
	*/

	//page := &Page{h, segments, r.buf[:pageSize]}
	r.page.Data = buf
	r.validPage = true
	//page := &Page{h, buf}
	//r.page = page
	r.ptr = 0
	return &r.page, nil
}

// capture should ensure that there is an 'OggS' in the stream. If seek is true
// then it should read forward and look for one.
func (r *Reader) capture(seek bool) error {
	// TODO: make it actually seek
	buf := make([]byte, 4)
	_, err := io.ReadFull(r.r, buf)
	if err != nil {
		return err
	}

	if string(buf) != CapturePattern {
		return ErrBadHeader
	}

	return nil
}
