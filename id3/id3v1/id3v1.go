package id3v1

import (
	"bytes"
	"encoding/binary"
	"io"
	"os"
	"strconv"
	"time"

	"ktkr.us/pkg/sound"
)

const Size = 128

var genres = []string{
	"Blues", "Classic Rock", "Country", "Dance", "Disco", "Funk", "Grunge",
	"Hip-Hop", "Jazz", "Metal", "New Age", "Oldies", "Other", "Pop", "R&B",
	"Rap", "Reggae", "Rock", "Techno", "Industrial", "Alternative", "Ska",
	"Death Metal", "Pranks", "Soundtrack", "Euro-Techno", "Ambient", "Trip-Hop",
	"Vocal", "Jazz+Funk", "Fusion", "Trance", "Classical", "Instrumental", "Acid",
	"House", "Game", "Sound Clip", "Gospel", "Noise", "Alternative Rock", "Bass",
	"Soul", "Punk", "Space", "Meditative", "Instrumental Pop", "Instrumental Rock",
	"Ethnic", "Gothic", "Darkwave", "Techno-Industrial", "Electronic", "Pop-Folk",
	"Eurodance", "Dream", "Southern Rock", "Comedy", "Cult", "Gangsta", "Top 40",
	"Christian Rap", "Pop/Funk", "Jungle", "Native US", "Cabaret", "New Wave",
	"Psychadelic", "Rave", "Showtunes", "Trailer", "Lo-Fi", "Tribal", "Acid Punk",
	"Acid Jazz", "Polka", "Retro", "Musical", "Rock & Roll", "Hard Rock", "Folk",
	"Folk-Rock", "National Folk", "Swing", "Fast Fusion", "Bebop", "Latin",
	"Revival", "Celtic", "Bluegrass", "Avantgarde", "Gothic Rock",
	"Progressive Rock", "Psychedelic Rock", "Symphonic Rock", "Slow Rock",
	"Big Band", "Chorus", "Easy Listening", "Acoustic", "Humour", "Speech",
	"Chanson", "Opera", "Chamber Music", "Sonata", "Symphony", "Booty Bass",
	"Primus", "Porn Groove", "Satire", "Slow Jam", "Club", "Tango", "Samba",
	"Folklore", "Ballad", "Power Ballad", "Rhytmic Soul", "Freestyle", "Duet",
	"Punk Rock", "Drum Solo", "Acapella", "Euro-House", "Dance Hall", "Goa",
	"Drum & Bass",
}

type Tag struct {
	title   string
	artist  string
	album   string
	Year    int
	comment string
	track   int
	genre   string
}

func (t *Tag) Title() string       { return t.title }
func (t *Tag) AlbumArtist() string { return t.artist }
func (t *Tag) Artist() string      { return t.artist }
func (t *Tag) Album() string       { return t.album }
func (t *Tag) Genre() string       { return t.genre }
func (t *Tag) Disc() int           { return 1 }
func (t *Tag) Track() int          { return t.track }
func (t *Tag) Date() time.Time {
	var tm time.Time
	tm.AddDate(t.Year, 0, 0)
	return tm
}
func (t *Tag) Composer() string { return "" }
func (t *Tag) Notes() string    { return t.comment }

type tag struct {
	Title      [30]byte
	Artist     [30]byte
	Album      [30]byte
	Year       [4]byte
	Comment    [29]byte
	AlbumTrack byte
	Genre      byte
}

func Decode(r io.Reader) (sound.Tags, error) {
	var (
		t   tag
		err error
	)

	if seeker, ok := r.(io.Seeker); ok {
		_, err := seeker.Seek(-Size, os.SEEK_END)
		if err != nil {
			return nil, err
		}
	} else {
		// use some other means to find it
		r, err = seekEnd(r, Size)
		if err != nil {
			return nil, err
		}
	}
	buf := make([]byte, 3)
	io.ReadFull(r, buf)
	if string(buf) != "TAG" {
		return nil, nil
	}
	binary.Read(r, binary.LittleEndian, &t)
	var genre string
	if int(t.Genre) < len(genres) {
		genre = genres[t.Genre]
	}
	year, err := strconv.Atoi(string(t.Year[:]))
	if err != nil {
		return nil, err
	}

	return &Tag{
		title:   trimString(t.Title[:]),
		artist:  trimString(t.Artist[:]),
		album:   trimString(t.Album[:]),
		Year:    year,
		comment: trimString(t.Comment[:]),
		track:   int(t.AlbumTrack),
		genre:   genre,
	}, nil
}

func seekEnd(r io.Reader, pos int) (io.Reader, error) {
	var (
		buf  []byte
		buf1 = make([]byte, 1<<15)
		buf2 = make([]byte, 1<<15)
		n    int
		err  error
	)

	for {
		n, err = r.Read(buf1)
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
		buf1, buf2 = buf2, buf1
	}

	if n < Size {
		buf = make([]byte, pos)
		m := copy(buf, buf2[len(buf2)-(pos-n):])
		copy(buf[m:], buf1[:n])
	} else {
		buf = buf1[n-pos : n]
	}

	return bytes.NewReader(buf), nil
}

func trimString(s []byte) string {
	i := bytes.Index(s, []byte{0})
	if i < 0 {
		return string(s)
	}
	return string(s[:i])
}
