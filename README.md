# imaging

A small Go package for basic image processing. It currently provides Canny
edge detection and Gaussian blur, along with a couple of pixel-level helpers
used to build them.

The package operates on the standard library's [`image.Image`](https://pkg.go.dev/image#Image)
interface and returns `*image.RGBA` (or `image.Image`) values, so it composes
cleanly with `image/png`, `image/jpeg`, and the rest of the `image` ecosystem.

## Install

```sh
go get github.com/mbrumlow/imaging
```

```go
import "github.com/mbrumlow/imaging"
```

## Usage

### Gaussian blur

`Gaussian` applies a 3x3 Gaussian kernel to the image. The blur is applied `n`
times, so a larger `n` produces a stronger blur. Passing `n == 0` copies the
image through a single pass without blurring.

```go
img := imaging.Gaussian(src, 10) // blur src ten times
```

### Edge detection

`Edge` runs [Canny edge detection](https://en.wikipedia.org/wiki/Canny_edge_detector)
and returns a black-and-white image where detected edges are white. It takes a
threshold `t` (edge strength cutoff — higher keeps only stronger edges) and a
blur count `b` (how many Gaussian passes to apply before detecting edges, which
reduces noise).

```go
img := imaging.Edge(src, 100, 0) // threshold 100, no pre-blur
```

## API

| Function | Description |
| --- | --- |
| `Gaussian(img image.Image, n int) *image.RGBA` | Returns a copy of `img` blurred `n` times with a Gaussian kernel. `n == 0` copies without blurring. |
| `Edge(img image.Image, t, b int) image.Image` | Returns a black/white image with edges (above strength `t`) drawn in white, after `b` Gaussian pre-blur passes. |
| `CalcBounds(img image.Image, x, y, b int) (int, int, int, int)` | Returns the clamped start/end `x,y` of a box of radius `b` centered at `x,y`, guaranteed to stay within the image bounds. |
| `CalcLum(img image.Image, x, y int) int` | Returns the luminance of the pixel at `x,y` (`0.2126*R + 0.7152*G + 0.0722*B`). |

## Notes

- `CalcLum` clamps coordinates that fall outside the image bounds to the
  nearest edge pixel instead of panicking. Sampling an out-of-range `x,y` now
  returns the luminance of the closest in-bounds pixel rather than triggering
  an out-of-range access.
- The blur and Sobel convolutions use edge replication at the borders: a pixel
  on the edge is convolved with the full 3x3 kernel, sampling the nearest
  in-bounds pixel for any neighbor that lies outside the image. This keeps the
  kernel weights aligned with the sampled pixels, so borders are no longer
  under-weighted (which previously darkened them). Blurring a uniform image is
  a no-op everywhere, including on the edges.

## Examples

Runnable programs live under [`example/`](./example). Each reads an image with
`-in`, processes it, and writes `out.png`.

```sh
# Blur an image (defaults to 10 passes; override with -b)
go run ./example/blur -in input.jpg -b 10

# Detect edges (defaults to threshold 100, no pre-blur; override with -t / -b)
go run ./example/edge -in input.jpg -t 100 -b 0
```

The examples decode PNG, JPEG, and [PPM](https://github.com/mbrumlow/ppm) input.
The PPM/PGM decoder is bundled under [`third_party/ppm`](./third_party/ppm) and
wired up with a `replace` directive in `go.mod`, so the examples build without
network access; point the `replace` at the upstream module if you prefer it.

### Original
![original image](https://cloud.githubusercontent.com/assets/1245807/18030068/a0d37df4-6c6f-11e6-8d7d-b8b8eb18b7e1.png)

### Edges
![edge-detected image](https://cloud.githubusercontent.com/assets/1245807/18030071/acfcc45a-6c6f-11e6-8388-12123e3c2fae.png)

### Blur
![blurred image](https://cloud.githubusercontent.com/assets/1245807/18030580/5f2d9554-6c80-11e6-98b7-7a685b7b4a5e.png)

## License

See [LICENSE](./LICENSE).
