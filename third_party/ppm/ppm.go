// Package ppm implements a decoder for Netpbm images and registers it with the
// standard library's image package. Importing it for its side effects makes
// image.Decode / image.DecodeConfig transparently handle Netpbm files:
//
//	import _ "github.com/mbrumlow/ppm"
//
// Both the ASCII (P2/P3) and binary (P5/P6) encodings are supported, for
// grayscale (PGM: P2/P5) and color (PPM: P3/P6). Every image is decoded into an
// *image.RGBA; grayscale samples are replicated across R, G and B, and samples
// are rescaled from the file's maxval to 8 bits per channel.
package ppm

import (
	"bufio"
	"fmt"
	"image"
	"image/color"
	"io"
	"strconv"
)

func init() {
	image.RegisterFormat("ppm", "P6", Decode, DecodeConfig)
	image.RegisterFormat("ppm", "P3", Decode, DecodeConfig)
	image.RegisterFormat("pgm", "P5", Decode, DecodeConfig)
	image.RegisterFormat("pgm", "P2", Decode, DecodeConfig)
}

// Decode reads a Netpbm image from r and returns it as an *image.RGBA.
func Decode(r io.Reader) (image.Image, error) {
	br := bufio.NewReader(r)

	magic, w, h, maxval, err := readHeader(br)
	if err != nil {
		return nil, err
	}

	binary := magic == "P5" || magic == "P6"
	color3 := magic == "P3" || magic == "P6"

	scale := func(v int) uint8 {
		if maxval == 255 {
			return uint8(v)
		}
		// Rescale the sample from [0, maxval] onto [0, 255].
		return uint8(v * 255 / maxval)
	}

	readSample := func() (int, error) {
		if !binary {
			return readInt(br)
		}
		if maxval < 256 {
			b, err := br.ReadByte()
			return int(b), err
		}
		hi, err := br.ReadByte()
		if err != nil {
			return 0, err
		}
		lo, err := br.ReadByte()
		if err != nil {
			return 0, err
		}
		return int(hi)<<8 | int(lo), nil
	}

	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			rr, err := readSample()
			if err != nil {
				return nil, err
			}
			gg, bb := rr, rr
			if color3 {
				if gg, err = readSample(); err != nil {
					return nil, err
				}
				if bb, err = readSample(); err != nil {
					return nil, err
				}
			}
			img.SetRGBA(x, y, color.RGBA{scale(rr), scale(gg), scale(bb), 0xFF})
		}
	}

	return img, nil
}

// DecodeConfig reads just enough of a Netpbm header to report the image's
// dimensions and color model.
func DecodeConfig(r io.Reader) (image.Config, error) {
	_, w, h, _, err := readHeader(bufio.NewReader(r))
	if err != nil {
		return image.Config{}, err
	}
	return image.Config{ColorModel: color.RGBAModel, Width: w, Height: h}, nil
}

// readHeader parses the magic number, dimensions and maximum sample value from
// a Netpbm header, skipping whitespace and #-comments between fields.
func readHeader(r *bufio.Reader) (magic string, w, h, maxval int, err error) {
	if magic, err = readToken(r); err != nil {
		return
	}
	switch magic {
	case "P2", "P3", "P5", "P6":
	default:
		return "", 0, 0, 0, fmt.Errorf("ppm: unsupported magic number %q", magic)
	}

	if w, err = readInt(r); err != nil {
		return
	}
	if h, err = readInt(r); err != nil {
		return
	}
	if maxval, err = readInt(r); err != nil {
		return
	}

	if w <= 0 || h <= 0 {
		return "", 0, 0, 0, fmt.Errorf("ppm: invalid dimensions %dx%d", w, h)
	}
	if maxval <= 0 || maxval > 65535 {
		return "", 0, 0, 0, fmt.Errorf("ppm: invalid maxval %d", maxval)
	}
	return
}

// readInt reads the next whitespace-delimited token and parses it as an integer.
func readInt(r *bufio.Reader) (int, error) {
	tok, err := readToken(r)
	if err != nil {
		return 0, err
	}
	n, err := strconv.Atoi(tok)
	if err != nil {
		return 0, fmt.Errorf("ppm: invalid integer %q", tok)
	}
	return n, nil
}

// readToken returns the next whitespace-delimited token, skipping any leading
// whitespace and #-comments. It consumes exactly one trailing whitespace byte
// as the delimiter, which is what separates a binary header's maxval from the
// raw pixel data that follows.
func readToken(r *bufio.Reader) (string, error) {
	var b byte
	var err error

	for {
		if b, err = r.ReadByte(); err != nil {
			return "", err
		}
		if b == '#' {
			if err = skipComment(r); err != nil {
				return "", err
			}
			continue
		}
		if !isSpace(b) {
			break
		}
	}

	buf := []byte{b}
	for {
		if b, err = r.ReadByte(); err != nil {
			if err == io.EOF {
				return string(buf), nil
			}
			return "", err
		}
		if b == '#' {
			_ = r.UnreadByte()
			return string(buf), nil
		}
		if isSpace(b) {
			return string(buf), nil
		}
		buf = append(buf, b)
	}
}

// skipComment consumes bytes up to and including the next newline.
func skipComment(r *bufio.Reader) error {
	for {
		b, err := r.ReadByte()
		if err != nil {
			return err
		}
		if b == '\n' {
			return nil
		}
	}
}

func isSpace(b byte) bool {
	switch b {
	case ' ', '\t', '\n', '\r', '\v', '\f':
		return true
	}
	return false
}
