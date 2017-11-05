package mp3

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestDuration(t *testing.T) {
	badTracks := []string{
		"kubivunuv",
		"nolutimed",
		"vamobovup",
		"vusavetuy",
		"wagonobil",
	}

	for _, file := range badTracks {
		g, err := filepath.Glob("/home/moshee/.pls/" + file + "*")
		if err != nil {
			t.Fatal(err)
		}
		f, err := os.Open(g[0])
		if err != nil {
			t.Fatal(err)
		}
		fi, err := f.Stat()
		if err != nil {
			t.Fatal(err)
		}
		m, err := DecodeMetaID3v2(f, fi.Size())
		if err != nil {
			t.Fatal(err)
		}

		fmt.Printf("%v\t%s\n", m.Duration(), m.Title())

		f.Close()
	}
}
