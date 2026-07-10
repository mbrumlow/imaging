// Command edge runs Canny edge detection on an image and writes the resulting
// black-and-white edge map to out.png.
// Usage: edge -in <path> [-t <threshold>] [-b <blur passes>].
package main

import (
	"flag"
	"fmt"
	"image"
	"image/png"
	"log"
	"os"

	"github.com/mbrumlow/imaging"

	// Register additional decoders so -in can point at JPEG or PPM files
	// (PNG is already registered by image/png above).
	_ "image/jpeg"

	_ "github.com/mbrumlow/ppm"
)

var in = flag.String("in", "", "Path to input file.")
var threshold = flag.Int("t", 100, "Edge strength threshold.")
var blur = flag.Int("b", 0, "Blur count.")

func main() {

	flag.Parse()

	if *in == "" {
		fmt.Printf("Please provide -in flags.\n")
		os.Exit(1)
	}

	img, err := loadImage(*in)
	if err != nil {
		log.Fatalf("Failed to load input image: %v\n", err)
	}

	i := imaging.Edge(img, *threshold, *blur)

	out, err := os.OpenFile("out.png", os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0600)
	if err != nil {
		log.Fatalf("Failed to open out file: %v\n", err)
	}

	if err := png.Encode(out, i); err != nil {
		log.Fatalf("Failed to encode out: %v\n", err)
	}
}

// loadImage opens and decodes the image at path p using whichever registered
// decoder matches the file format.
func loadImage(p string) (image.Image, error) {

	f, err := os.Open(p)
	if err != nil {
		return nil, err
	}

	i, _, err := image.Decode(f)
	if err != nil {
		return nil, err
	}

	return i, nil
}
