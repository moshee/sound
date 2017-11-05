package mp3

import (
	"errors"
	"io"
	"io/ioutil"
	"time"

	"ktkr.us/pkg/sound"
	"ktkr.us/pkg/sound/id3/id3v1"
	"ktkr.us/pkg/sound/id3/id3v2"
)

var (
	ErrUnsynced      = errors.New("mp3: missing frame sync")
	ErrReserved      = errors.New("mp3: layer or MPEG version code has reserved value")
	ErrBadBitrate    = errors.New("mp3: disallowed bitrate code")
	ErrBadSampleRate = errors.New("mp3: disallowed sample rate code")
)

func init() {
	sound.RegisterFormat("MP3 ID3v2.2", "ID3\x02", Decode, id3v2.Decode, DecodeMetaID3v2)
	sound.RegisterFormat("MP3 ID3v2.3", "ID3\x03", Decode, id3v2.Decode, DecodeMetaID3v2)
	sound.RegisterFormat("MP3 ID3v2.4", "ID3\x04", Decode, id3v2.Decode, DecodeMetaID3v2)
	sound.RegisterFormat("MPEG-2 Layer III", "\xFF\xF2", Decode, id3v1.Decode, DecodeMeta)
	sound.RegisterFormat("MPEG-2 Layer III", "\xFF\xF3", Decode, id3v1.Decode, DecodeMeta)
	sound.RegisterFormat("MPEG-2 Layer II", "\xFF\xF4", Decode, id3v1.Decode, DecodeMeta)
	sound.RegisterFormat("MPEG-2 Layer II", "\xFF\xF5", Decode, id3v1.Decode, DecodeMeta)
	sound.RegisterFormat("MPEG-2 Layer I", "\xFF\xF6", Decode, id3v1.Decode, DecodeMeta)
	sound.RegisterFormat("MPEG-2 Layer I", "\xFF\xF7", Decode, id3v1.Decode, DecodeMeta)
	sound.RegisterFormat("MPEG-1 Layer III", "\xFF\xFA", Decode, id3v1.Decode, DecodeMeta)
	sound.RegisterFormat("MPEG-1 Layer III", "\xFF\xFB", Decode, id3v1.Decode, DecodeMeta)
	sound.RegisterFormat("MPEG-1 Layer II", "\xFF\xFC", Decode, id3v1.Decode, DecodeMeta)
	sound.RegisterFormat("MPEG-1 Layer II", "\xFF\xFD", Decode, id3v1.Decode, DecodeMeta)
	sound.RegisterFormat("MPEG-1 Layer I", "\xFF\xFE", Decode, id3v1.Decode, DecodeMeta)
	sound.RegisterFormat("MPEG-1 Layer I", "\xFF\xFF", Decode, id3v1.Decode, DecodeMeta)
}

// AAAAAAAA AAABBCCD EEEEFFGH IIJJKLMM
// 11111111 1111001X

func Decode(r io.Reader) (sound.Sound, error) {
	panic("kek")
}

const (
	version2_5      = 0
	versionReserved = 1
	version2        = 2
	version1        = 3

	layerReserved = 0
	layerIII      = 1
	layerII       = 2
	layerI        = 3

	channelStereo      = 0
	channelJointStereo = 1
	channelDualChannel = 2
	channelMono        = 3
)

var (
	bitrates = [4][4][16]int{
		version1: {
			layerI:   {0, 32, 64, 96, 128, 160, 192, 224, 256, 288, 320, 352, 384, 416, 448, -1},
			layerII:  {0, 32, 48, 56, 64, 80, 96, 112, 128, 160, 192, 224, 256, 320, 384, -1},
			layerIII: {0, 32, 40, 48, 56, 64, 80, 96, 112, 128, 160, 192, 224, 256, 320, -1},
		},
		version2: {
			layerI:   {0, 32, 48, 56, 64, 80, 96, 112, 128, 144, 160, 176, 192, 224, 256, -1},
			layerII:  {0, 8, 16, 24, 32, 40, 48, 56, 64, 80, 96, 112, 128, 144, 160, -1},
			layerIII: {0, 8, 16, 24, 32, 40, 48, 56, 64, 80, 96, 112, 128, 144, 160, -1},
		},
		version2_5: {
			layerI:   {0, 32, 48, 56, 64, 80, 96, 112, 128, 144, 160, 176, 192, 224, 256, -1},
			layerII:  {0, 8, 16, 24, 32, 40, 48, 56, 64, 80, 96, 112, 128, 144, 160, -1},
			layerIII: {0, 8, 16, 24, 32, 40, 48, 56, 64, 80, 96, 112, 128, 144, 160, -1},
		},
	}

	sampleRates = [4][4]int{
		version1:   {44100, 48000, 32000, -1},
		version2:   {22050, 24000, 16000, -1},
		version2_5: {11025, 12000, 8000, -1},
	}

	samplesPerFrame = [4][4]int{
		version1: {
			layerI:   384,
			layerII:  1152,
			layerIII: 1152,
		},
		version2: {
			layerI:   384,
			layerII:  1152,
			layerIII: 576,
		},
		version2_5: {
			layerI:   384,
			layerII:  1152,
			layerIII: 576,
		},
	}
)

type meta struct {
	duration   time.Duration
	channels   int
	bitrate    int
	samplerate int
	sound.Tags
}

func (m *meta) Duration() time.Duration { return m.duration }
func (m *meta) NumChannels() int        { return m.channels }
func (m *meta) BitRate() int            { return m.bitrate }
func (m *meta) SampleRate() int         { return m.samplerate }

type frameHeader struct {
	mpegVersion int
	layer       int
	haveCRC     bool
	bitrate     int
	samplerate  int
	channelMode int
	havePadding bool
	frameSize   int
}

type frameData struct {
	io.LimitedReader
}

func (fd *frameData) Close() error {
	_, err := io.Copy(ioutil.Discard, fd)
	//print("B")
	return err
}

type frame struct {
	frameHeader
	*frameData
}
