package average_color

import (
	"image"
	"image/color"
	"math"

	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
)

func AverageColor(img image.Image) color.NRGBA {
	bounds := img.Bounds()
	pxNumber := bounds.Max.X * bounds.Max.Y
	var reds, greens, blues, alphas uint64
	for x := 0; x < bounds.Max.X; x++ {
		for y := 0; y < bounds.Max.Y; y++ {
			red, green, blue, alpha := img.At(x, y).RGBA()
			reds += uint64(red)
			greens += uint64(green)
			blues += uint64(blue)
			alphas += uint64(alpha)
		}
	}

	var red, green, blue, alpha uint8
	k := uint64(0xffff / 0xff)
	red = uint8(math.Round(float64(reds/k) / float64(pxNumber)))
	green = uint8(math.Round(float64(greens/k) / float64(pxNumber)))
	blue = uint8(math.Round(float64(blues/k) / float64(pxNumber)))
	alpha = uint8(math.Round(float64(alphas/k) / float64(pxNumber)))

	return color.NRGBA{red, green, blue, alpha}
}
