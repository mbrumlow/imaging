package imaging

import (
	"image"
	"math"
)

// Edge returns a new black-and-white image in which detected edges are white
// (0xFFFFFFFF) and everything else is black, using the Canny edge detection
// algorithm: https://en.wikipedia.org/wiki/Canny_edge_detector
//
// img is the input image. t is the gradient-strength threshold: only pixels
// whose (suppressed) intensity exceeds t are marked as edges, so a higher t
// keeps only stronger edges. b is the number of Gaussian blur passes applied
// before detection to reduce noise (see Gaussian); b == 0 disables blurring.
func Edge(img image.Image, t, b int) image.Image {

	// 1. Smooth the input to suppress noise before differentiating.
	out := Gaussian(img, b)

	// 2. Compute the gradient magnitude and (quantized) direction per pixel.
	hyp, deg := intencityGradient(out)

	// 3. Thin the gradient response to single-pixel-wide edges.
	max := nonMaximumSuppression(hyp, deg, img.Bounds().Max.X)

	// 4. Threshold: paint surviving pixels white, everything else black.
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

// intencityGradient returns two parallel slices, indexed in row-major order:
// the gradient magnitude ("hyp") and the gradient direction in degrees ("deg")
// for every pixel. The gradient is computed with the Sobel operator
// (https://en.wikipedia.org/wiki/Sobel_operator), and each direction is snapped
// to the nearest of the four axes an image grid supports: 0, 45, 90 or 135.
func intencityGradient(img image.Image) ([]int, []int) {

	hyp := make([]int, 0, img.Bounds().Max.X*img.Bounds().Max.Y)
	deg := make([]int, 0, img.Bounds().Max.X*img.Bounds().Max.Y)

	for y := img.Bounds().Min.Y; y < img.Bounds().Max.Y; y++ {
		for x := img.Bounds().Min.X; x < img.Bounds().Max.X; x++ {

			// Horizontal and vertical Sobel responses.
			gX := xG(img, x, y)
			gY := yG(img, x, y)

			// Magnitude via hypot; direction folded into the 0–180 range.
			g := math.Hypot(float64(gX), float64(gY))
			o := math.Atan2(float64(gY), float64(gX))
			o = math.Abs(o * 180 / math.Pi)

			// Quantize the angle to the nearest grid direction.
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

// nonMaximumSuppression thins wide gradient responses down to single-pixel
// edges: each pixel is kept only if its magnitude is strictly greater than the
// two neighbors that lie along its gradient direction; otherwise it is zeroed.
// width is the image width, used to convert the flat index into row/column
// neighbors. See the non-maximum suppression step of the Canny algorithm:
// https://en.wikipedia.org/wiki/Canny_edge_detector
func nonMaximumSuppression(hyp, deg []int, width int) (out []int) {

	out = make([]int, 0, len(hyp))

	for i := range hyp {

		// The gradient direction selects which pair of neighbors to compare.
		z := deg[i]

		switch z {

		case 0: // horizontal gradient: compare west and east neighbors

			w, e := 0, 0

			// Guard against reading across a row boundary or off the slice.
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

		case 90: // vertical gradient: compare north and south neighbors

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

		case 135: // diagonal gradient: compare north-west and south-east

			nw, se := 0, 0

			// NW moves left a column, so guard against the left edge.
			if i%width != 0 && i-width-1 >= 0 {
				nw = hyp[i-width-1]
			}

			// SE moves right a column, so guard against the right edge.
			if (i+1)%width != 0 && i+width+1 < len(hyp) {
				se = hyp[i+width+1]
			}

			if hyp[i] > nw && hyp[i] > se {
				out = append(out, hyp[i])
			} else {
				out = append(out, 0)
			}

		case 45: // diagonal gradient: compare north-east and south-west

			ne, sw := 0, 0

			// NE moves right a column, so guard against the right edge.
			if (i+1)%width != 0 && i-width+1 >= 0 {
				ne = hyp[i-width+1]
			}

			// SW moves left a column, so guard against the left edge.
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

// xG approximates the horizontal derivative at (x, y) using the horizontal
// Sobel kernel.
func xG(a image.Image, x, y int) int {
	k := []int{-1, 0, 1, -2, 0, 2, -1, 0, 1}
	return processKern(k, a, x, y)
}

// yG approximates the vertical derivative at (x, y) using the vertical Sobel
// kernel.
func yG(a image.Image, x, y int) int {
	k := []int{-1, -2, -1, 0, 0, 0, 1, 2, 1}
	return processKern(k, a, x, y)
}

// processKern convolves the 3x3 kernel k with the luminance of the neighborhood
// around (x, y) and returns the accumulated response. CalcBounds keeps the
// window inside the image, so pixels on the border sample a clamped window.
func processKern(k []int, img image.Image, x, y int) int {

	// c walks the kernel; xg accumulates the weighted luminance response.
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
