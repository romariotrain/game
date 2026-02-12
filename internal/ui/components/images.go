package components

import (
	"fmt"
	"image"
	"math"
	"os"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	xdraw "golang.org/x/image/draw"

	_ "image/jpeg"
	_ "image/png"
)

// RoundedImageFromFile loads an image, stretches it to the requested size,
// and applies rounded corners via alpha mask.
func RoundedImageFromFile(path string, size fyne.Size, cornerRadius float32) (*canvas.Image, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open image: %w", err)
	}
	defer file.Close()

	src, _, err := image.Decode(file)
	if err != nil {
		return nil, fmt.Errorf("decode image: %w", err)
	}

	width := int(math.Round(float64(size.Width)))
	height := int(math.Round(float64(size.Height)))
	if width < 1 || height < 1 {
		return nil, fmt.Errorf("invalid target size: %dx%d", width, height)
	}

	radius := int(math.Round(float64(cornerRadius)))
	dst := image.NewNRGBA(image.Rect(0, 0, width, height))
	xdraw.CatmullRom.Scale(dst, dst.Bounds(), src, src.Bounds(), xdraw.Src, nil)
	applyRoundedMask(dst, radius)

	img := canvas.NewImageFromImage(dst)
	img.FillMode = canvas.ImageFillStretch
	img.SetMinSize(size)
	return img, nil
}

func applyRoundedMask(dst *image.NRGBA, radius int) {
	if radius <= 0 {
		return
	}

	width := dst.Bounds().Dx()
	height := dst.Bounds().Dy()
	if width < 2 || height < 2 {
		return
	}

	if radius > width/2 {
		radius = width / 2
	}
	if radius > height/2 {
		radius = height / 2
	}
	if radius <= 0 {
		return
	}

	r := float64(radius)
	rr := r * r
	left := float64(radius - 1)
	right := float64(width - radius)
	top := float64(radius - 1)
	bottom := float64(height - radius)

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			if (x >= radius && x < width-radius) || (y >= radius && y < height-radius) {
				continue
			}

			cx := left
			if x >= width-radius {
				cx = right
			}
			cy := top
			if y >= height-radius {
				cy = bottom
			}

			dx := float64(x) - cx
			dy := float64(y) - cy
			if dx*dx+dy*dy <= rr {
				continue
			}

			offset := dst.PixOffset(x, y)
			dst.Pix[offset+0] = 0
			dst.Pix[offset+1] = 0
			dst.Pix[offset+2] = 0
			dst.Pix[offset+3] = 0
		}
	}
}
