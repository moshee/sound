package id3v2

import (
	"fmt"
	"os"
	"testing"
)

func TestTextFrame(t *testing.T) {
	f, err := os.Open("/home/moshee/.pls/pihupaxog.06 Mitch Murder - Montage.mp3")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	tags, err := Decode(f)
	if err != nil {
		t.Fatal(err)
	}

	tt := tags.(*Tags)

	for k, v := range tt.Frames {
		fmt.Printf("%s\t%s\n", k, v)
	}
}
