package average_color

import (
	"image"
	"image/color"
	"math"
	"sync"

	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
)

type colors struct {
	mux sync.Mutex
	// sum of color values ∈ [0,255]
	redSum   float64
	greenSum float64
	blueSum  float64
	alphaSum float64
}

func (c *colors) Inc(r, g, b, a float64) {
	c.mux.Lock()
	defer c.mux.Unlock()

	c.redSum += r
	c.greenSum += g
	c.blueSum += b
	c.alphaSum += a
}

func AverageColor(img image.Image) color.NRGBA {
	bounds := img.Bounds()
	pxNumber := bounds.Max.X * bounds.Max.Y
	clrs := &colors{}

	var wg sync.WaitGroup
	wg.Add(bounds.Max.Y)
	for y := 0; y < bounds.Max.Y; y++ {
		go func(y int) {
			defer wg.Done()
			var reds, greens, blues, alphas float64
			for x := 0; x < bounds.Max.X; x++ {
				redAP, greenAP, blueAP, alphaAP := img.At(x, y).RGBA() // alpha-premultiplied values
				red := float64(redAP*0xff) / float64(alphaAP)
				green := float64(greenAP*0xff) / float64(alphaAP)
				blue := float64(blueAP*0xff) / float64(alphaAP)
				alpha := float64(alphaAP * 0xff / 0xffff)
				reds += red
				greens += green
				blues += blue
				alphas += alpha
			}
			clrs.Inc(reds, greens, blues, alphas)
		}(y)
	}
	wg.Wait()

	var red, green, blue, alpha uint8
	red = uint8(math.Round(clrs.redSum / float64(pxNumber)))
	green = uint8(math.Round(clrs.greenSum / float64(pxNumber)))
	blue = uint8(math.Round(clrs.blueSum / float64(pxNumber)))
	alpha = uint8(math.Round(clrs.alphaSum / float64(pxNumber)))

	return color.NRGBA{red, green, blue, alpha}
}
