package imaging

import (
	"image"
	"math"
)

// Edge returns a black-and-white image with detected edges drawn in white and
// everything else black. Edges are found using the Canny edge detection
// algorithm (https://en.wikipedia.org/wiki/Canny_edge_detector).
//
// t is the edge-strength threshold: only gradients stronger than t are kept as
// edges, so a higher t keeps only stronger edges. b is the number of Gaussian
// pre-blur passes applied before detection to reduce noise.
func Edge(img image.Image, t, b int) image.Image {

	// A degenerate image has no pixels to detect edges in.
	if img.Bounds().Dx() == 0 || img.Bounds().Dy() == 0 {
		return image.NewRGBA(img.Bounds())
	}

	out := Gaussian(img, b)

	hyp, deg := intensityGradient(out)
	suppressed := nonMaximumSuppression(hyp, deg, img.Bounds().Dx())

	for y, i := img.Bounds().Min.Y, 0; y < img.Bounds().Max.Y; y++ {
		for x := img.Bounds().Min.X; x < img.Bounds().Max.X; x, i = x+1, i+1 {

			o := out.PixOffset(x, y)

			if suppressed[i] > t {
				out.Pix[o+0] = 0xFF
				out.Pix[o+1] = 0xFF
				out.Pix[o+2] = 0xFF
				out.Pix[o+3] = 0xFF
			} else {
				out.Pix[o+0] = 0x00
				out.Pix[o+1] = 0x00
				out.Pix[o+2] = 0x00
				out.Pix[o+3] = 0xFF
			}
		}
	}

	return out
}

// intensityGradient returns the image intensities and their direction.
// Image intensities are processed using the Sobel operator.
// https://en.wikipedia.org/wiki/Sobel_operator
func intensityGradient(img image.Image) ([]int, []int) {

	hyp := make([]int, 0, img.Bounds().Dx()*img.Bounds().Dy())
	deg := make([]int, 0, img.Bounds().Dx()*img.Bounds().Dy())

	for y := img.Bounds().Min.Y; y < img.Bounds().Max.Y; y++ {
		for x := img.Bounds().Min.X; x < img.Bounds().Max.X; x++ {

			gX := xG(img, x, y)
			gY := yG(img, x, y)
			g := math.Hypot(float64(gX), float64(gY))
			o := math.Atan2(float64(gY), float64(gX))
			o = math.Abs(o * 180 / math.Pi)

			if o > 0 && o <= 22.5 || o > 157.5 && o <= 180 {
				o = 0
			}

			if o > 22.5 && o <= 67.5 {
				o = 45
			}

			if o > 67.5 && o <= 112.5 {
				o = 90
			}

			if o > 112.5 && o <= 157.5 {
				o = 135
			}

			hyp = append(hyp, int(g))
			deg = append(deg, int(o))
		}
	}

	return hyp, deg
}

// nonMaximumSuppression thins the edge.
// See Non-maximum suppression at https://en.wikipedia.org/wiki/Canny_edge_detector
func nonMaximumSuppression(hyp, deg []int, width int) (out []int) {

	out = make([]int, 0, len(hyp))

	for i := range hyp {

		z := deg[i]

		switch z {

		case 0: // 0 - east and west

			w, e := 0, 0

			// Don't wrap and don't overflow.
			if i%width != 0 && i-1 >= 0 {
				w = hyp[i-1]
			}

			// Don't wrap and don't overflow.
			if (i+1)%width != 0 && i+1 < len(hyp) {
				e = hyp[i+1]
			}

			if hyp[i] > w && hyp[i] > e {
				out = append(out, hyp[i])
			} else {
				out = append(out, 0)
			}

		case 90: // 90 - north south

			n, s := 0, 0

			if i-width >= 0 {
				n = hyp[i-width]
			}

			if i+width < len(hyp) {
				s = hyp[i+width]
			}

			if hyp[i] > n && hyp[i] > s {
				out = append(out, hyp[i])
			} else {
				out = append(out, 0)
			}

		case 135: // 135 - north west and south east

			nw, se := 0, 0

			// Don't wrap (nw is on the left edge, se on the right) and don't overflow.
			if i%width != 0 && i-width-1 >= 0 {
				nw = hyp[i-width-1]
			}

			if (i+1)%width != 0 && i+width+1 < len(hyp) {
				se = hyp[i+width+1]
			}

			if hyp[i] > nw && hyp[i] > se {
				out = append(out, hyp[i])
			} else {
				out = append(out, 0)
			}

		case 45: // 45 - north east and south west

			ne, sw := 0, 0

			// Don't wrap (ne is on the right edge, sw on the left) and don't overflow.
			if (i+1)%width != 0 && i-width+1 >= 0 {
				ne = hyp[i-width+1]
			}

			if i%width != 0 && i+width-1 < len(hyp) {
				sw = hyp[i+width-1]
			}

			if hyp[i] > ne && hyp[i] > sw {
				out = append(out, hyp[i])
			} else {
				out = append(out, 0)
			}

		default:
			out = append(out, 0)
		}
	}

	return
}

// xG calculates the horizontal derivative approximation at x,y coordinates.
// The horizontal Sobel kernel.
func xG(a image.Image, x, y int) int {
	k := []int{-1, 0, 1, -2, 0, 2, -1, 0, 1}
	return processKern(k, a, x, y)
}

// yG calculates the vertical derivative approximation at x,y coordinates.
// The vertical Sobel kernel.
func yG(a image.Image, x, y int) int {
	k := []int{-1, -2, -1, 0, 0, 0, 1, 2, 1}
	return processKern(k, a, x, y)
}

// processKern convolves the given 3x3 kernel with the luminance of the pixels
// around x,y. Coordinates outside the image are clamped to the nearest edge
// pixel by CalcLum, so the kernel indices always line up with the sampled
// neighbors, including on the border.
func processKern(k []int, img image.Image, x, y int) int {

	c, xg := 0, 0
	for dy := -1; dy <= 1; dy++ {
		for dx := -1; dx <= 1; dx++ {
			xg += k[c] * CalcLum(img, x+dx, y+dy)
			c++
		}
	}

	return xg
}
