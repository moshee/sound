package mp4

import (
	"bufio"
	"encoding/binary"
	"errors"
	"io"

	"ktkr.us/pkg/sound"
)

var (
	ErrInvalidFormat = errors.New("mp4: invalid format")
)

const (
	atomHeaderSize = 8
)

type AtomHeader struct {
	Size uint32
	Name [4]byte
}

type Atom struct {
	Name     string
	Content  []byte
	Parent   *Atom
	Children map[string][]*Atom
}

func (a *Atom) Get(path ...string) *Atom {
	if a.Children == nil || len(a.Children) == 0 || len(path) == 0 {
		return nil
	}

	current := a

	for _, name := range path {
		children := current.Children[name]
		if children == nil || len(children) == 0 {
			return nil
		}

		current = children[0]
	}

	return current
}

type Reader struct {
	Type string
	r    *bufio.Reader
}

func NewReader(r io.Reader) (*Reader, error) {
	br := bufio.NewReader(r)
	peek, err := br.Peek(8)
	if err != nil {
		return nil, err
	}

	// look for ftyp atom
	if string(peek[4:]) != "ftyp" {
		return nil, ErrInvalidFormat
	}

	rr := &Reader{r}

	a, err := rr.ReadAtom()
	if err != nil {
		return nil, err
	}
}

/*
root:
	read atom
	if parent atom
	  goto root
	else
	  read data
	  if dual atom
	    goto root
      else
	    if exhausted child atoms
		  go up a level
		else
	      read next child
*/
func (r *Reader) ReadAtom() (*Atom, error) {
	var h AtomHeader
	err := binary.Read(r.r, binary.BigEndian, &h)
	if err != nil {
		return nil, err
	}

	buf := make([]byte, h.Size-atomHeaderSize)

}

func ReadTags(r io.Reader) (sound.Tags, error) {
	return nil, nil
}
