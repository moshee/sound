package mp3

import (
	"bufio"
	"io"
	"math"
	"time"

	"ktkr.us/pkg/sound"
	"ktkr.us/pkg/sound/id3/id3v2"
)

// DecodeMeta decodes metadata out of an MP3 stream, attempting to calculate
// the duration and decode the ID3v1 header if there is one.
func DecodeMeta(rr io.Reader, fsize int64) (sound.Metadata, error) {
	r := newReader(rr)
	f, err := r.nextFrame()
	if err != nil {
		//print(6)
		return nil, err
	}

	var numFrames int
	buf := make([]byte, 4)
	_, err = io.ReadFull(f, buf)
	if err != nil {
		return nil, err
	}

	switch string(buf) {
	case "Xing", "Info":
		xing, err := decodeXing(f)
		if err != nil {
			//print(7)
			return nil, err
		}
		numFrames = int(xing.NumFrames)

	case "VBRI":
		vbri, err := decodeVBRI(f)
		if err != nil {
			//print(8)
			return nil, err
		}
		numFrames = int(vbri.NumFrames)
	}

	f.Close()

	var duration time.Duration

	if numFrames == 0 {
		//log.Print(fsize, f.bitrate)
		secs := math.Floor(float64(fsize)/float64(f.bitrate/8) + 0.5)
		duration = time.Second * time.Duration(secs)
	} else {
		var (
			spf        = samplesPerFrame[f.mpegVersion][f.layer]
			numSamples = numFrames * spf
			secs       = math.Floor(float64(numSamples)/float64(f.samplerate) + 0.5)
		)
		duration = time.Duration(secs) * time.Second
	}

	/*
		tags, err := id3v1.Decode(rr)
		if err != nil {
			//print(9)
			return nil, err
		}
		if numFrames == 0 {
			cr := rr.(*countReader)
			if f.bitrate != 0 {
				duration = time.Duration(cr.n) / time.Duration(f.bitrate) * time.Second
			} else {
				// Well, I'm stumped. There's no VBR header or bitrate information
				// to help us calculate the length of the track.
			}
		}
	*/

	m := &meta{
		duration:   duration,
		bitrate:    f.bitrate,
		samplerate: f.samplerate,
		//Tags:       tags,
	}

	if f.channelMode == channelMono {
		m.channels = 1
	} else {
		m.channels = 2
	}
	//print("A")
	return m, nil
}

/*
type countReader struct {
	*bufio.Reader
	n int
}

func (r *countReader) Read(p []byte) (n int, err error) {
	n, err = r.r.Read(p)
	r.n += n
	return
}
*/

// DecodeMetaID3v2 decodes the metadata of an MP3 stream assuming it begins
// with an ID3v2 tag.
func DecodeMetaID3v2(r io.Reader, fsize int64) (sound.Metadata, error) {
	// discount the bytes read from the id3v2 tag before calculating CBR duration
	rr := ensureBufioReader(r)

	tags, err := id3v2.Decode(rr)
	if err != nil {
		//print(3)
		return nil, err
	}
	//br := r.(*bufio.Reader)
	//log.Print(br.Buffered())
	//x, _ := br.Peek(16)
	//log.Printf("%x", x)
	v2tags := tags.(*id3v2.Tags)
	m, err := DecodeMeta(rr, fsize-int64(v2tags.Size))
	if err != nil {
		//print(4)
		return nil, err
	}
	// Prefer id3v2 over id3v1
	mm := m.(*meta)
	mm.Tags = tags
	//print(5)
	return mm, nil
}

func ensureBufioReader(r io.Reader) *bufio.Reader {
	if br, ok := r.(*bufio.Reader); ok {
		return br
	}
	return bufio.NewReader(r)
}
