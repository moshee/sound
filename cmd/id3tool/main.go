package main

import (
	"flag"
	"log"
	"os"

	"ktkr.us/pkg/fmtutil"
	"ktkr.us/pkg/sound"
	_ "ktkr.us/pkg/sound/mp3"
	_ "ktkr.us/pkg/sound/ogg"
	// _ "ktkr.us/pkg/sound/wave"
)

func main() {
	log.SetFlags(0)
	flag.Parse()

	if flag.NArg() < 1 {
		log.Fatalf("usage: %s <mp3 filename>", os.Args[0])
	}

	f, err := os.Open(flag.Arg(0))
	if err != nil {
		log.Fatal(err)
	}

	defer f.Close()

	meta, name, err := sound.DecodeMeta(f)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("%s, %s, %d kbps", name, fmtutil.HMS(meta.Duration()), meta.BitRate()/1024)

	switch n := meta.NumChannels(); n {
	case 0:
	case 1:
		log.Print("mono")
	case 2:
		log.Print("stereo")
	default:
		log.Printf("%d channels", n)
	}

	f.Seek(0, os.SEEK_SET)

	tags, _, err := sound.DecodeTags(f)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Title:       %q (%d)", tags.Title(), len(tags.Title()))
	log.Printf("AlbumArtist: %q (%d)", tags.AlbumArtist(), len(tags.AlbumArtist()))
	log.Printf("Artist:      %q (%d)", tags.Artist(), len(tags.Artist()))
	log.Printf("Album:       %q (%d)", tags.Album(), len(tags.Album()))
	log.Printf("Genre:       %q (%d)", tags.Genre(), len(tags.Genre()))
	log.Printf("Disc:        %d", tags.Disc())
	log.Printf("Track:       %d", tags.Track())
	log.Printf("Date:        %v", tags.Date())
	log.Printf("Composer:    %q (%d)", tags.Composer(), len(tags.Composer()))
	log.Printf("Notes:       %q (%d)", tags.Notes(), len(tags.Notes()))
}
