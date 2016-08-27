package imaging

import (
	"image"
	"math"
)

// Edge returns a new image that has had all the edges within the given
// threshold set to 0xFFFFFFFF.
// Img is the input image.
// Image edges are detected using the Canny edge detection algorithm defined at
// https://en.wikipedia.org/wiki/Canny_edge_detector .
func Edge(img image.Image, t, b int) image.Image {

	out := Gaussian(img, b)

	hyp, deg := intencityGradient(out)
	max := nonMaximumSuppression(hyp, deg, img.Bounds().Max.X)

	for y, i := img.Bounds().Min.Y, 0; y < img.Bounds().Max.Y; y++ {
		for x := img.Bounds().Min.X; x < img.Bounds().Max.X; x, i = x+1, i+1 {

			o := (y*img.Bounds().Max.X + x) * 4

			if max[i] > t {
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

// IntencityGradient returns the image intensities and their direction.
// Image intensities are processed using the Sorbel operator.
// https://en.wikipedia.org/wiki/Sobel_operator
func intencityGradient(img image.Image) ([]int, []int) {

	hyp := make([]int, 0, img.Bounds().Max.X*img.Bounds().Max.Y)
	deg := make([]int, 0, img.Bounds().Max.X*img.Bounds().Max.Y)

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

// NonMaximumSuppression thins the edge.
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

			// Don't wrap and don't overlow.
			if i+1%width != 0 && i+1 < len(hyp) {
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

			if i-1%width != 0 && i-width-1 >= 0 {
				nw = hyp[i-width-1]
			}

			if i+1%width != 0 && i+width+1 < len(hyp) {
				se = hyp[i+width+1]
			}

			if hyp[i] > nw && hyp[i] > se {
				out = append(out, hyp[i])
			} else {
				out = append(out, 0)
			}

		case 45: // 45 - north east and south west

			ne, sw := 0, 0

			if i-width+1 >= 0 {
				ne = hyp[i-width+1]
			}

			if i+width-1 < len(hyp) {
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

// XG calculates the horizontal derivative approximation at x,y coordinates.
// The horizontal Sorbel kernel.
func xG(a image.Image, x, y int) int {
	k := []int{-1, 0, 1, -2, 0, 2, -1, 0, 1}
	return processKern(k, a, x, y)
}

// XG calculates the vertical derivative approximation at x,y coordinates.
// The vertical Sorbel kernel.
func yG(a image.Image, x, y int) int {
	k := []int{-1, -2, -1, 0, 0, 0, 1, 2, 1}
	return processKern(k, a, x, y)
}

// ProcessKern processes the given kernel on the x,y coordinates while respecting image boundaries.
func processKern(k []int, img image.Image, x, y int) int {

	c, xg := 0, 0
	sx, sy, ex, ey := CalcBounds(img, x, y, 1)

	for y := sy; y <= ey; y++ {
		for x := sx; x <= ex; x++ {
			avg := CalcLum(img, x, y)
			xg += k[c] * avg
			c++
		}
	}

	return xg
}
