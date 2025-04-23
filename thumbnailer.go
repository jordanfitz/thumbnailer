// Package thumbnailer contains a utility with which thumbnails can be generated from an image.
package thumbnailer

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"math"

	"golang.org/x/image/draw"
)

type OutputFormat uint8

const (
	OriginalFormat OutputFormat = iota
	JPG
	PNG
)

const (
	formatJPG      = "jpeg"
	formatPNG      = "png"
	DefaultMaxSize = 300
)

var (
	ErrInvalidImage = errors.New("invalid image")
)

type Option func(t *Thumbnailer)

// Image sets the JPG or PNG image data from which thumbnails can be generated.
func Image(value []byte) Option {
	img := make([]byte, len(value))
	copy(img, value)
	return func(t *Thumbnailer) {
		t.img = img
	}
}

// MaxSize sets a size which the scaled image's largest dimension will not exceed.
func MaxSize(value int) Option {
	return func(t *Thumbnailer) {
		t.maxSize = value
	}
}

// Quality sets the JPG quality used by Create. It has no effect if the output format is not JPG.
// By default, [jpeg.DefaultQuality] is used.
func Quality(value int) Option {
	return func(t *Thumbnailer) {
		t.jpgQuality = value
	}
}

// OutFormat sets the output image format used by Create.
// By default, the format of the original image is used.
func OutFormat(value OutputFormat) Option {
	if value > PNG {
		value = OriginalFormat
	}
	return func(t *Thumbnailer) {
		t.outFormat = value
	}
}

// Scaler sets the [draw.Scaler] used by Create.
// By default, the [draw.ApproxBiLinear] scaler is used.
func Scaler(value draw.Scaler) Option {
	return func(t *Thumbnailer) {
		t.scaler = value
	}
}

type Thumbnailer struct {
	scaler     draw.Scaler
	img        []byte
	options    []Option
	maxSize    int
	jpgQuality int
	outFormat  OutputFormat
}

// New creates a new instance of [Thumbnailer] with which thumbnails can be generated.
func New(options ...Option) Thumbnailer {
	return Thumbnailer{
		scaler:     draw.ApproxBiLinear,
		img:        nil,
		options:    options,
		maxSize:    DefaultMaxSize,
		jpgQuality: jpeg.DefaultQuality,
		outFormat:  OriginalFormat,
	}
}

// With applies an option to the Thumbnailer instance.
func (t Thumbnailer) With(o Option) Thumbnailer {
	t.options = append(t.options, o)
	return t
}

func (t Thumbnailer) encodeJPG(img *image.RGBA) ([]byte, error) {
	var buffer bytes.Buffer
	if err := jpeg.Encode(&buffer, img, &jpeg.Options{
		Quality: t.jpgQuality,
	}); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

func (t Thumbnailer) encodePNG(img *image.RGBA) ([]byte, error) {
	var buffer bytes.Buffer
	if err := png.Encode(&buffer, img); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

func (t Thumbnailer) encode(img *image.RGBA) ([]byte, error) {
	switch t.outFormat {
	case JPG:
		return t.encodeJPG(img)
	case PNG:
		return t.encodePNG(img)
	}
	return nil, fmt.Errorf("unexpected output format")
}

// Create generates a thumbnail, returning the encoded thumbnail image or an error.
func (t Thumbnailer) Create() ([]byte, error) {
	for _, option := range t.options {
		option(&t)
	}

	originalImage, format, err := image.Decode(bytes.NewReader(t.img))
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	if t.outFormat == OriginalFormat {
		switch format {
		case formatJPG:
			t.outFormat = JPG
		case formatPNG:
			t.outFormat = PNG
		default:
			return nil, fmt.Errorf("invalid image format '%s'", format)
		}
	}

	bounds := originalImage.Bounds().Max
	newWidth, newHeight := scaleDimensions(t.maxSize, bounds.X, bounds.Y)

	scaledRect := image.Rect(0, 0, newWidth, newHeight)
	scaledImage := image.NewRGBA(scaledRect)

	t.scaler.Scale(scaledImage, scaledRect, originalImage, originalImage.Bounds(), draw.Over, nil)

	return t.encode(scaledImage)
}

func scaleDimensions(maxSize, width, height int) (newWidth, newHeight int) {
	hRatio := float64(1)
	if width > maxSize {
		hRatio = float64(maxSize) / float64(width)
	}

	vRatio := float64(1)
	if height > maxSize {
		vRatio = float64(maxSize) / float64(height)
	}

	scale := min(hRatio, vRatio)
	if scale == 1 {
		return width, height
	}

	newWidth = int(math.Round(float64(width) * scale))
	newHeight = int(math.Round(float64(height) * scale))

	return newWidth, newHeight
}
