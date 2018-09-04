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
	mux    sync.Mutex
	reds   uint64
	greens uint64
	blues  uint64
	alphas uint64
}

func (c *colors) Inc(r, g, b, a uint64) {
	c.mux.Lock()
	defer c.mux.Unlock()

	c.reds += r
	c.greens += g
	c.blues += b
	c.alphas += a
}

func AverageColor(img image.Image) color.NRGBA {
	bounds := img.Bounds()
	pxNumber := bounds.Max.X * bounds.Max.Y
	clrs := &colors{reds: 0, greens: 0, blues: 0, alphas: 0}

	var wg sync.WaitGroup
	wg.Add(bounds.Max.X)
	for x := 0; x < bounds.Max.X; x++ {
		go func(x int) {
			defer wg.Done()
			var lReds, lGreens, lBlues, lAlphas uint64
			for y := 0; y < bounds.Max.Y; y++ {
				red, green, blue, alpha := img.At(x, y).RGBA()
				lReds += uint64(red)
				lGreens += uint64(green)
				lBlues += uint64(blue)
				lAlphas += uint64(alpha)
			}
			clrs.Inc(lReds, lGreens, lBlues, lAlphas)
		}(x)
	}
	wg.Wait()

	var red, green, blue, alpha uint8
	k := uint64(0xffff / 0xff)
	red = uint8(math.Round(float64(clrs.reds/k) / float64(pxNumber)))
	green = uint8(math.Round(float64(clrs.greens/k) / float64(pxNumber)))
	blue = uint8(math.Round(float64(clrs.blues/k) / float64(pxNumber)))
	alpha = uint8(math.Round(float64(clrs.alphas/k) / float64(pxNumber)))

	return color.NRGBA{red, green, blue, alpha}
}
