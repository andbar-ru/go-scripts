package main

import (
	"fyne.io/fyne/v2"
)

const (
	dotRadius = float32(50)
	dotSize   = dotRadius * 2 // 100x100 bounding box
)

type oneDotLayout struct{}

func (l *oneDotLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	for _, o := range objects {
		o.Resize(fyne.NewSize(dotSize, dotSize)) // never scales with the container
		o.Move(fyne.NewPos(
			(size.Width-dotSize)/2,
			(size.Height-dotSize)/2,
		))
	}
}

func (l *oneDotLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	return fyne.NewSize(dotSize, dotSize)
}
