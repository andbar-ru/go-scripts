package main

import (
	"fyne.io/fyne/v2"
)

const (
	dotRadius = float32(50)
	dotSize   = dotRadius * 2
)

var (
	dotFyneSize = fyne.NewSize(dotSize, dotSize)
)

type oneDotLayout struct{}

func (l *oneDotLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	if len(objects) != 1 {
		return
	}
	o := objects[0]
	o.Resize(dotFyneSize) // never scales with the container
	o.Move(fyne.NewPos((size.Width-dotSize)/2, (size.Height-dotSize)/2))
}

func (l *oneDotLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	return dotFyneSize
}

type twoDotsLayout struct{}

func (l *twoDotsLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	if len(objects) != 2 {
		return
	}
	for _, o := range objects {
		o.Resize(dotFyneSize)
	}
	objects[0].Move(fyne.NewPos(size.Width-dotSize, 0))
	objects[1].Move(fyne.NewPos(0, size.Height-dotSize))
}

func (l *twoDotsLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	return fyne.NewSize(dotSize*2, dotSize*2)
}

type threeDotsLayout struct{}

func (l *threeDotsLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	if len(objects) != 3 {
		return
	}
	for _, o := range objects {
		o.Resize(dotFyneSize)
	}
	objects[0].Move(fyne.NewPos(size.Width-dotSize, 0))
	objects[1].Move(fyne.NewPos((size.Width-dotSize)/2, (size.Height-dotSize)/2))
	objects[2].Move(fyne.NewPos(0, size.Height-dotSize))
}

func (l *threeDotsLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	return fyne.NewSize(dotSize*3, dotSize*3)
}

type fourDotsLayout struct{}

func (l *fourDotsLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	if len(objects) != 4 {
		return
	}
	for _, o := range objects {
		o.Resize(dotFyneSize)
	}
	objects[0].Move(fyne.NewPos(0, 0))
	objects[1].Move(fyne.NewPos(size.Width-dotSize, 0))
	objects[2].Move(fyne.NewPos(0, size.Height-dotSize))
	objects[3].Move(fyne.NewPos(size.Width-dotSize, size.Height-dotSize))
}

func (l *fourDotsLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	return fyne.NewSize(dotSize*2+10, dotSize*2+10)
}

type fiveDotsLayout struct{}

func (l *fiveDotsLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	if len(objects) != 5 {
		return
	}
	for _, o := range objects {
		o.Resize(dotFyneSize)
	}
	objects[0].Move(fyne.NewPos(0, 0))
	objects[1].Move(fyne.NewPos(size.Width-dotSize, 0))
	objects[2].Move(fyne.NewPos((size.Width-dotSize)/2, (size.Height-dotSize)/2))
	objects[3].Move(fyne.NewPos(0, size.Height-dotSize))
	objects[4].Move(fyne.NewPos(size.Width-dotSize, size.Height-dotSize))
}

func (l *fiveDotsLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	return fyne.NewSize(dotSize*3, dotSize*3)
}

type sixDotsLayout struct{}

func (l *sixDotsLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	if len(objects) != 6 {
		return
	}
	for _, o := range objects {
		o.Resize(dotFyneSize)
	}
	objects[0].Move(fyne.NewPos(0, 0))
	objects[1].Move(fyne.NewPos(size.Width-dotSize, 0))
	objects[2].Move(fyne.NewPos(0, (size.Height-dotSize)/2))
	objects[3].Move(fyne.NewPos(size.Width-dotSize, (size.Height-dotSize)/2))
	objects[4].Move(fyne.NewPos(0, size.Height-dotSize))
	objects[5].Move(fyne.NewPos(size.Width-dotSize, size.Height-dotSize))
}

func (l *sixDotsLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	return fyne.NewSize(dotSize*2+10, dotSize*3+20)
}
