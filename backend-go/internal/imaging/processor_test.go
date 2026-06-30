package imaging

import (
	"bytes"
	"errors"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"testing"
)

// makeTestImage builds an opaque RGBA gradient at w×h. We use a gradient (not a
// solid color) so the JPEG encoder produces non-trivial output that the decoder
// can round-trip without quantizing to a single color.
func makeTestImage(w, h int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, color.RGBA{
				R: uint8((x * 255) / w),
				G: uint8((y * 255) / h),
				B: uint8(((x + y) * 255) / (w + h)),
				A: 255,
			})
		}
	}
	return img
}

func encodeJPEG(t *testing.T, img image.Image, q int) []byte {
	t.Helper()
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: q}); err != nil {
		t.Fatalf("encode jpeg: %v", err)
	}
	return buf.Bytes()
}

func encodePNG(t *testing.T, img image.Image) []byte {
	t.Helper()
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatalf("encode png: %v", err)
	}
	return buf.Bytes()
}

// decodeDimensions returns the width/height of the JPEG bytes by re-decoding
// them. We don't compare pixel values because lossy JPEG round-tripping makes
// exact equality brittle; the dimensions are what we actually care about.
func decodeDimensions(t *testing.T, b []byte) (int, int) {
	t.Helper()
	cfg, _, err := image.DecodeConfig(bytes.NewReader(b))
	if err != nil {
		t.Fatalf("decode dims: %v", err)
	}
	return cfg.Width, cfg.Height
}

func TestProcess_JPEGLandscape_DownscalesAndProducesThumb(t *testing.T) {
	src := makeTestImage(4000, 3000) // landscape, larger than full max
	in := encodeJPEG(t, src, 90)

	p := NewProcessor()
	out, err := p.Process(in, "image/jpeg")
	if err != nil {
		t.Fatalf("process: %v", err)
	}

	if out.ContentType != "image/jpeg" {
		t.Errorf("content type = %q", out.ContentType)
	}
	if out.Width != 4000 || out.Height != 3000 {
		t.Errorf("original dims = %dx%d, want 4000x3000", out.Width, out.Height)
	}

	fw, fh := decodeDimensions(t, out.Full)
	if fw != DefaultFullMaxDim {
		t.Errorf("full width = %d, want %d", fw, DefaultFullMaxDim)
	}
	// 1920 * 3000/4000 = 1440
	if fh != 1440 {
		t.Errorf("full height = %d, want 1440 (preserve aspect)", fh)
	}

	tw, th := decodeDimensions(t, out.Thumb)
	if tw != DefaultThumbMaxDim {
		t.Errorf("thumb width = %d, want %d", tw, DefaultThumbMaxDim)
	}
	if th != 240 { // 320 * 3000/4000
		t.Errorf("thumb height = %d, want 240 (preserve aspect)", th)
	}

	if len(out.Thumb) >= len(out.Full) {
		t.Errorf("thumb (%d B) should be smaller than full (%d B)", len(out.Thumb), len(out.Full))
	}
}

func TestProcess_JPEGPortrait_FitsLongestEdge(t *testing.T) {
	src := makeTestImage(1500, 4000) // portrait
	in := encodeJPEG(t, src, 90)

	out, err := NewProcessor().Process(in, "image/jpeg")
	if err != nil {
		t.Fatalf("process: %v", err)
	}

	fw, fh := decodeDimensions(t, out.Full)
	if fh != DefaultFullMaxDim {
		t.Errorf("full height = %d, want %d (longest edge)", fh, DefaultFullMaxDim)
	}
	if fw != 720 { // 1500 * 1920/4000
		t.Errorf("full width = %d, want 720", fw)
	}
}

func TestProcess_SmallImage_NotUpscaled(t *testing.T) {
	src := makeTestImage(200, 150) // smaller than thumb max
	in := encodeJPEG(t, src, 90)

	out, err := NewProcessor().Process(in, "image/jpeg")
	if err != nil {
		t.Fatalf("process: %v", err)
	}

	fw, fh := decodeDimensions(t, out.Full)
	if fw != 200 || fh != 150 {
		t.Errorf("full dims = %dx%d, want 200x150 (no upscale)", fw, fh)
	}
	tw, th := decodeDimensions(t, out.Thumb)
	if tw != 200 || th != 150 {
		t.Errorf("thumb dims = %dx%d, want 200x150 (no upscale)", tw, th)
	}
}

func TestProcess_PNGInput_OutputsJPEG(t *testing.T) {
	src := makeTestImage(800, 600)
	in := encodePNG(t, src)

	out, err := NewProcessor().Process(in, "image/png")
	if err != nil {
		t.Fatalf("process: %v", err)
	}
	if out.ContentType != "image/jpeg" {
		t.Errorf("content type = %q, want image/jpeg (we always re-encode)", out.ContentType)
	}
	// Verify the output is decodable as JPEG.
	if _, _, err := image.Decode(bytes.NewReader(out.Full)); err != nil {
		t.Errorf("full not decodable: %v", err)
	}
}

func TestProcess_RejectsHEIC(t *testing.T) {
	_, err := NewProcessor().Process([]byte("not really heic"), "image/heic")
	if !errors.Is(err, ErrUnsupportedType) {
		t.Errorf("got %v, want ErrUnsupportedType", err)
	}
}

func TestProcess_RejectsHEIF(t *testing.T) {
	_, err := NewProcessor().Process([]byte("not really heif"), "image/heif")
	if !errors.Is(err, ErrUnsupportedType) {
		t.Errorf("got %v, want ErrUnsupportedType", err)
	}
}

func TestProcess_RejectsUnknownType(t *testing.T) {
	_, err := NewProcessor().Process([]byte("garbage"), "image/gif")
	if !errors.Is(err, ErrUnsupportedType) {
		t.Errorf("got %v, want ErrUnsupportedType", err)
	}
}

func TestProcess_RejectsCorruptInput(t *testing.T) {
	_, err := NewProcessor().Process([]byte("not actually a jpeg"), "image/jpeg")
	if err == nil {
		t.Error("expected decode error for garbage input")
	}
	if errors.Is(err, ErrUnsupportedType) {
		t.Error("decode failure should not surface as ErrUnsupportedType")
	}
}

func TestProcess_CustomDimensionsAndQuality(t *testing.T) {
	src := makeTestImage(2000, 2000)
	in := encodeJPEG(t, src, 95)

	p := &Processor{FullMaxDim: 800, ThumbMaxDim: 100, FullQuality: 60, ThumbQuality: 50}
	out, err := p.Process(in, "image/jpeg")
	if err != nil {
		t.Fatalf("process: %v", err)
	}
	fw, _ := decodeDimensions(t, out.Full)
	if fw != 800 {
		t.Errorf("full width = %d, want 800 (custom)", fw)
	}
	tw, _ := decodeDimensions(t, out.Thumb)
	if tw != 100 {
		t.Errorf("thumb width = %d, want 100 (custom)", tw)
	}
}
