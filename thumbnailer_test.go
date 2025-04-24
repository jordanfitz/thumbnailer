package thumbnailer

import (
	"bytes"
	"image"
	"math"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
)

func loadTestImage(t *testing.T, name string) []byte {
	filePath := path.Join("./testdata/", name)
	data, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("failed to load test image: %v", err)
	}
	return data
}

func decode(t *testing.T, img []byte) (image.Image, string) {
	decoded, format, err := image.Decode(bytes.NewReader(img))
	assert.NoError(t, err)
	return decoded, format
}

func dimensions(img image.Image) (width, height int) {
	return img.Bounds().Max.X, img.Bounds().Max.Y
}

func aspectRatio(w, h int) float64 {
	return float64(w) / float64(h)
}

func approximatelyEqual(a, b float64) bool {
	return math.Abs(a-b) <= 1e-3
}

func TestThumbnailer_MaxSize(t *testing.T) {
	t.Parallel()

	testImage := loadTestImage(t, "soccerball.png")

	const maxSize = 200

	original, originalFormat := decode(t, testImage)
	originalWidth, originalHeight := dimensions(original)
	originalRatio := aspectRatio(originalWidth, originalHeight)

	thumbnailData, err := New(Image(testImage), MaxSize(maxSize)).Create()
	assert.NoError(t, err)

	thumbnail, thumbnailFormat := decode(t, thumbnailData)
	thumbnailWidth, thumbnailHeight := dimensions(thumbnail)
	thumbnailRatio := aspectRatio(thumbnailWidth, thumbnailHeight)

	// by default, the output format should be the same as the input format
	assert.Equal(t, originalFormat, thumbnailFormat)
	// since the dimensions of an image are integral, the conversion to and from floating point
	// values will result in slightly different ratios
	assert.True(t, approximatelyEqual(originalRatio, thumbnailRatio),
		"before and after aspect ratios should be approximately equal")

	assert.LessOrEqual(t, thumbnailWidth, maxSize)
	assert.LessOrEqual(t, thumbnailHeight, maxSize)
}

func TestThumbnailer_OutFormat(t *testing.T) {
	t.Parallel()

	testImage := loadTestImage(t, "soccerball.png")

	thumbnailData, err := New(Image(testImage), OutFormat(JPG)).Create()
	assert.NoError(t, err)

	_, thumbnailFormat := decode(t, thumbnailData)
	assert.Equal(t, thumbnailFormat, formatJPG)
}

func TestThumbnailer_Quality(t *testing.T) {
	t.Parallel()

	testImage := loadTestImage(t, "soccerball.png")

	thumb := New(Image(testImage), OutFormat(JPG))
	defaultQuality, err := thumb.Create()
	assert.NoError(t, err)

	lowerQuality, err := thumb.With(Quality(20)).Create()
	assert.NoError(t, err)

	assert.Lessf(t, len(lowerQuality), len(defaultQuality), "lower quality should produce smaller output")
}

func TestThumbnailer_NoImage(t *testing.T) {
	t.Parallel()

	_, err := New().Create()
	assert.Error(t, err)
}

func TestThumbnailer_BadData(t *testing.T) {
	t.Parallel()

	_, err := New(Image([]byte("this is not an image!"))).Create()
	assert.Error(t, err)
}
