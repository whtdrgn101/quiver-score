// Package imaging produces sized JPEG renditions of user-uploaded photos.
//
// The plan originally called for WebP output, but the only credible pure-Go
// WebP encoder (HugoSmits86/nativewebp) is lossless-only, which produces
// larger files than lossy JPEG for continuous-tone photos. JPEG Q80/Q70 hits
// the bandwidth target with the standard library and no CGO dependency.
package imaging

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	_ "image/png" // register PNG decoder

	"golang.org/x/image/draw"
	_ "golang.org/x/image/webp" // register WebP decoder
)

// Default rendition sizes and JPEG quality levels.
const (
	DefaultFullMaxDim    = 1920
	DefaultThumbMaxDim   = 320
	DefaultFullQuality   = 80
	DefaultThumbQuality  = 70
	OutputContentType    = "image/jpeg"
)

// ErrUnsupportedType is returned when the caller-declared content type is not
// one we accept. Callers can errors.Is it to distinguish from decode failures.
var ErrUnsupportedType = errors.New("imaging: unsupported image type")

// Result holds the JPEG bytes and metadata for a processed upload.
type Result struct {
	Full        []byte // full-size JPEG (≤ FullMaxDim on the longer edge)
	Thumb       []byte // thumbnail JPEG (≤ ThumbMaxDim on the longer edge)
	ContentType string // always image/jpeg
	Width       int    // original decoded width in pixels
	Height      int    // original decoded height in pixels
}

// Processor renders uploads to JPEG full + thumb pairs. The zero value works;
// NewProcessor sets the defaults explicitly for clarity.
type Processor struct {
	FullMaxDim   int
	ThumbMaxDim  int
	FullQuality  int
	ThumbQuality int
}

func NewProcessor() *Processor {
	return &Processor{
		FullMaxDim:   DefaultFullMaxDim,
		ThumbMaxDim:  DefaultThumbMaxDim,
		FullQuality:  DefaultFullQuality,
		ThumbQuality: DefaultThumbQuality,
	}
}

// Process decodes input and returns JPEG-encoded full and thumbnail renditions.
// inputContentType is the client-declared MIME type from the upload; HEIC is
// rejected up front because there is no pure-Go HEIC decoder — clients must
// convert HEIC to JPEG before upload.
func (p *Processor) Process(input []byte, inputContentType string) (*Result, error) {
	switch inputContentType {
	case "image/jpeg", "image/png", "image/webp":
		// supported
	case "image/heic", "image/heif":
		return nil, fmt.Errorf("%w: %s — clients must convert HEIC/HEIF to JPEG before upload", ErrUnsupportedType, inputContentType)
	default:
		return nil, fmt.Errorf("%w: %s", ErrUnsupportedType, inputContentType)
	}

	src, _, err := image.Decode(bytes.NewReader(input))
	if err != nil {
		return nil, fmt.Errorf("imaging: decode %s: %w", inputContentType, err)
	}
	bounds := src.Bounds()

	full, err := p.render(src, p.fullMaxDim(), p.fullQuality())
	if err != nil {
		return nil, fmt.Errorf("imaging: encode full: %w", err)
	}
	thumb, err := p.render(src, p.thumbMaxDim(), p.thumbQuality())
	if err != nil {
		return nil, fmt.Errorf("imaging: encode thumb: %w", err)
	}

	return &Result{
		Full:        full,
		Thumb:       thumb,
		ContentType: OutputContentType,
		Width:       bounds.Dx(),
		Height:      bounds.Dy(),
	}, nil
}

func (p *Processor) render(src image.Image, maxDim, quality int) ([]byte, error) {
	out := fit(src, maxDim)
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, out, &jpeg.Options{Quality: quality}); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// fit returns src downscaled so its longest edge is at most maxDim. Smaller
// images are returned unchanged (we never upscale — uploads stay at their
// native resolution if already below the cap).
func fit(src image.Image, maxDim int) image.Image {
	b := src.Bounds()
	w, h := b.Dx(), b.Dy()
	if w <= maxDim && h <= maxDim {
		return src
	}
	var nw, nh int
	if w >= h {
		nw = maxDim
		nh = h * maxDim / w
	} else {
		nh = maxDim
		nw = w * maxDim / h
	}
	dst := image.NewRGBA(image.Rect(0, 0, nw, nh))
	// CatmullRom is the highest-quality of the stdlib resamplers — slower than
	// ApproxBiLinear but the difference is irrelevant at our throughput.
	draw.CatmullRom.Scale(dst, dst.Bounds(), src, b, draw.Over, nil)
	return dst
}

func (p *Processor) fullMaxDim() int {
	if p.FullMaxDim > 0 {
		return p.FullMaxDim
	}
	return DefaultFullMaxDim
}

func (p *Processor) thumbMaxDim() int {
	if p.ThumbMaxDim > 0 {
		return p.ThumbMaxDim
	}
	return DefaultThumbMaxDim
}

func (p *Processor) fullQuality() int {
	if p.FullQuality > 0 {
		return p.FullQuality
	}
	return DefaultFullQuality
}

func (p *Processor) thumbQuality() int {
	if p.ThumbQuality > 0 {
		return p.ThumbQuality
	}
	return DefaultThumbQuality
}
