package id3v2

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
	"unicode"
)

var txxxEquiv = map[string]string{
	"ALBUM":             "TALB",
	"BPM":               "TBPM",
	"COMPOSER":          "TCOM",
	"GENRE":             "TCON",
	"COPYRIGHT":         "TCOP",
	"ENCODINGTIME":      "TDEN",
	"PLAYLISTDELAY":     "TDLY",
	"ORIGINALDATE":      "TDOR",
	"DATE":              "TDRC",
	"RELEASEDATE":       "TDRL",
	"TAGGINGDATE":       "TDTG",
	"ENCODEDBY":         "TENC",
	"LYRICIST":          "TEXT",
	"FILETYPE":          "TFLT",
	"CONTENTGROUP":      "TIT1",
	"TITLE":             "TIT2",
	"SUBTITLE":          "TIT3",
	"INITIALKEY":        "TKEY",
	"LANGUAGE":          "TLAN",
	"LENGTH":            "TLEN",
	"MEDIA":             "TMED",
	"MOOD":              "TMOO",
	"ORIGINALALBUM":     "TOAL",
	"ORIGINALFILENAME":  "TOFN",
	"ORIGINALLYRICIST":  "TOLY",
	"ORIGINALARTIST":    "TOPE",
	"OWNER":             "TOWN",
	"ARTIST":            "TPE1",
	"ALBUMARTIST":       "TPE2",
	"CONDUCTOR":         "TPE3",
	"REMIXER":           "TPE4",
	"DISCNUMBER":        "TPOS",
	"PRODUCEDNOTICE":    "TPRO",
	"LABEL":             "TPUB",
	"TRACKNUMBER":       "TRCK",
	"RADIOSTATION":      "TRSN",
	"RADIOSTATIONOWNER": "TRSO",
	"ALBUMSORT":         "TSOA",
	"ARTISTSORT":        "TSOP",
	"TITLESORT":         "TSOT",
	"ALBUMARTISTSORT":   "TSO2",
	"ISRC":              "TSRC",
	"ENCODING":          "TSSE",
}

var v22Equiv = map[string]string{
	"BUF": "RBUF", "CNT": "PCNT", "COM": "COMM", "CRA": "AENC",
	"ETC": "ETCO", "GEO": "GEOB", "IPL": "TIPL", "MCI": "MCDI",
	"MLL": "MLLT", "POP": "POPM", "REV": "RVRB", "SLT": "SYLT",
	"STC": "SYTC", "TAL": "TALB", "TBP": "TBPM", "TCM": "TCOM",
	"TCO": "TCON", "TCP": "TCMP", "TCR": "TCOP", "TDY": "TDLY",
	"TEN": "TENC", "TFT": "TFLT", "TKE": "TKEY", "TLA": "TLAN",
	"TLE": "TLEN", "TMT": "TMED", "TOA": "TOAL", "TOF": "TOFN",
	"TOL": "TOLY", "TOR": "TDOR", "TOT": "TOAL", "TP1": "TPE1",
	"TP2": "TPE2", "TP3": "TPE3", "TP4": "TPE4", "TPA": "TPOS",
	"TPB": "TPUB", "TRC": "TSRC", "TRD": "TDRC", "TRK": "TRCK",
	"TS2": "TSO2", "TSA": "TSOA", "TSC": "TSOC", "TSP": "TSOP",
	"TSS": "TSSE", "TST": "TSOT", "TT1": "TIT1", "TT2": "TIT2",
	"TT3": "TIT3", "TXT": "TOLY", "TXX": "TXXX", "TYE": "TDRC",
	"UFI": "UFID", "ULT": "USLT", "WAF": "WOAF", "WAR": "WOAR",
	"WAS": "WOAS", "WCM": "WCOM", "WCP": "WCOP", "WPB": "WPUB",
	"WXX": "WXXX",
}

const (
	encISO8859_1 = 0x00
	encUTF16_BOM = 0x01
	encUTF16BE   = 0x02
	encUTF8      = 0x03
)

var (
	ErrEmptyText    = errors.New("id3v2: empty text field")
	ErrMalformedBOM = errors.New("id3v2: malformed UTF-16 BOM")
)

func readTerminatedString(enc byte, r *bytes.Buffer) (string, error) {
	if enc == encISO8859_1 || enc == encUTF8 {
		s, err := r.ReadString('\x00')
		if err != nil {
			return "", err
		}
		if len(s) <= 1 {
			return "", nil
		}
		return s[:len(s)-1], nil
	}

	// string is null-terminated with two null bytes if unicode encoded

	// disgusting allocation here but what can you do
	buf := []byte{}

	for {
		next := r.Next(2)

		if len(next) != 2 {
			return "", errors.New("id3v2: unexpected eof inside terminated string")
		}

		if bytes.Equal(next, []byte("\x00\x00")) {
			return string(buf), nil
		}

		buf = append(buf, next[0])
		r.UnreadByte()
	}
}

// decode using first byte as encoding identifier
// $00 = ISO-8859-1,    null byte terminated
// $01 = UTF-16 w/ BOM, null word terminated
// $02 = UTF-16BE,      null word terminated
// $03 = UTF-8,         null byte terminated
func decodeTextFrame(enc byte, buf []byte, unsynch bool) (string, error) {
	if len(buf) == 0 {
		return "", nil
	}

	var s string

	if unsynch {
		buf = bytes.Replace(buf, []byte{0xFF, 0x00}, []byte{0xFF}, -1)
	}

	switch enc {
	case encISO8859_1:
		s = decodeLatin1(buf)
	case encUTF16_BOM:
		bom := string(buf[:2])
		buf = buf[2:]

		switch bom {
		case "\xff\xfe":
			s = decodeUTF16BE(buf)
		case "\xfe\xff":
			s = decodeUTF16LE(buf)
		default:
			return "", ErrMalformedBOM
		}
	case encUTF16BE:
		s = decodeUTF16BE(buf)
	case encUTF8:
		s = string(buf)
	default:
		return "", fmt.Errorf("id3v2: unknown encoding 0x%02x", enc)
	}

	return strings.TrimRight(s, "\x00"), nil
}

func decodeLatin1(buf []byte) string {
	r := make([]rune, len(buf))
	for i := range buf {
		r[i] = rune(buf[i])
	}
	return string(r)
}

func decodeUTF16BE(buf []byte) string {
	s := make([]rune, 0, len(buf)/2)
	for i := 0; i < len(buf); i += 2 {
		s = append(s, rune(buf[i])|rune(buf[i+1])<<8)
	}
	return string(s)
}

func decodeUTF16LE(buf []byte) string {
	s := make([]rune, 0, len(buf)/2)
	for i := 0; i < len(buf); i += 2 {
		s = append(s, rune(buf[i])<<8|rune(buf[i+1]))
	}
	return string(s)
}

func decodeTXXX(txxx map[string]string, buf []byte, unsynch bool) error {
	var (
		enc = buf[0]
		b   = bytes.NewBuffer(buf[1:])
	)

	name, err := readTerminatedString(enc, b)
	if err != nil {
		return err
	}

	s, err := decodeTextFrame(enc, b.Bytes(), unsynch)
	if err != nil {
		return err
	}

	txxx[name] = s
	return nil
}

func translateTXXXFrames(frames map[string]string, txxx map[string]string) {
	for key, val := range txxx {
		// log.Printf("TXXX %q: %q", key, val)
		frameID, ok := txxxEquiv[key]
		if ok {
			//log.Printf("TXXX.%s => %s : %s", key, frameID, val)
			frames[frameID] = val
		}
	}
}

func parseMultiNumber(s string) (n1, n2 int, err error) {
	arr := strings.FieldsFunc(s, unicode.IsPunct)

	if len(arr) > 0 {
		n1, err = strconv.Atoi(arr[0])
		if err != nil {
			return
		}
	}

	if len(arr) >= 2 {
		n2, err = strconv.Atoi(arr[1])
		if err != nil {
			return
		}
	}

	return
}

func parseDate(frames map[string]string) (tm time.Time, err error) {
	TYER := frames["TYER"]
	TDAT := frames["TDAT"]
	TIME := frames["TIME"]

	// log.Println(TYER, TDAT, TIME)

	if TYER != "" {
		if TDAT == "" {
			tm, err = time.Parse("2006", TYER)
			if err == nil {
				return
			}
		} else {
			if TIME == "" {
				TIME = "0000"
			}
			tm, err = time.Parse("200602011504", TYER+TDAT+TIME)
			if err != nil {
				tm, err = time.Parse("2006", TYER)
				if err == nil {
					return
				}
			} else {
				return
			}
		}
	}

	// last resort brute force
	dateFrames := []string{"TDRC", "TDRL", "TDOR", "TDAT", "TIME", "TYER"}

	for _, frame := range dateFrames {
		if val, ok := frames[frame]; ok {
			tm, err = tryAllDateFormats(val)
			if err == nil {
				// log.Println(val, tm)
				return
			}
		}
	}

	return time.Time{}, nil
}

var dateFormats = []string{
	"2006-01-02T15:04:05Z0700",
	"2006-01-02T15:04:05Z07:00",
	"2006-01-02T15:04:05",
	"2006-01-02T15:04Z0700",
	"2006-01-02T15:04Z07:00",
	"2006-01-02T15:04",
	"2006-01-02T15",
	"2006-01-02",
	"2006-01",
	"2006",
	"2006/01/02",
	"2006.01.02",
}

func tryAllDateFormats(s string) (tm time.Time, err error) {
	for i, f := range dateFormats {
		tm, err = time.Parse(f, s)
		if err != nil {
			if i == len(dateFormats)-1 {
				return
			}
			err = nil
			continue
		}
		break
	}
	return
}
