# imaging

A small Go image-processing package. It provides two operations — a **Gaussian
blur** and **Canny edge detection** — built on a handful of pixel helpers. It
was written to support another project, so the scope is intentionally narrow.

The package operates on the standard library's [`image.Image`](https://pkg.go.dev/image#Image)
interface and returns `*image.RGBA`, so it composes with `image/png`,
`image/jpeg`, and any other codec registered with the `image` package.

## Install

```sh
go get github.com/mbrumlow/imaging
```

```go
import "github.com/mbrumlow/imaging"
```

## API

### `Gaussian(img image.Image, n int) *image.RGBA`

Applies a 3×3 Gaussian blur to `img`. `n` is the number of passes: each pass
blurs the output of the previous one, so a larger `n` produces a wider, softer
blur. As a special case, `n == 0` copies the image without blurring.

```go
blurred := imaging.Gaussian(src, 10)
```

### `Edge(img image.Image, t, b int) image.Image`

Runs [Canny edge detection](https://en.wikipedia.org/wiki/Canny_edge_detector)
and returns a black-and-white image in which detected edges are white and
everything else is black.

- `t` — gradient-strength threshold; only pixels whose (thinned) gradient
  magnitude exceeds `t` are marked as edges. Raise it to keep only strong edges.
- `b` — number of Gaussian blur passes applied first to suppress noise. Use
  `0` to skip pre-blurring.

```go
edges := imaging.Edge(src, 100, 0)
```

### Helpers

- `CalcBounds(img, x, y, b)` — returns the clamped start/end coordinates of a
  square window of radius `b` centered on `(x, y)`, kept inside the image.
- `CalcLum(img, x, y)` — returns the Rec. 709 luminance (0–255) of a pixel.

## Examples

The [`example/`](example) directory contains two runnable command-line tools.
Both read `-in <path>` (PNG, JPEG, or PPM) and write the result to `out.png`.

Blur — `-b` sets the number of blur passes (default 10):

```sh
go run ./example/blur -in input.png -b 10
```

Edge — `-t` sets the threshold (default 100) and `-b` the pre-blur passes
(default 0):

```sh
go run ./example/edge -in input.png -t 100 -b 0
```

### Original
![original](https://cloud.githubusercontent.com/assets/1245807/18030068/a0d37df4-6c6f-11e6-8d7d-b8b8eb18b7e1.png)

### Edges
![edges](https://cloud.githubusercontent.com/assets/1245807/18030071/acfcc45a-6c6f-11e6-8388-12123e3c2fae.png)

### Blur
![blur](https://cloud.githubusercontent.com/assets/1245807/18030580/5f2d9554-6c80-11e6-98b7-7a685b7b4a5e.png)

## License

MIT — see [LICENSE](LICENSE).
