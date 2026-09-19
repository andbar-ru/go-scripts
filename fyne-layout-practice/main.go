package main

import (
	"fmt"
	"image/color"
	"log"
	"os"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
)

const (
	argConstraint = "The argument must be an integer between 1 and 6."
)

var faceLayouts = map[int]fyne.Layout{
	1: new(oneDotLayout),
	2: new(twoDotsLayout),
	3: new(threeDotsLayout),
	4: new(fourDotsLayout),
	5: new(fiveDotsLayout),
	6: new(sixDotsLayout),
}

func makeDots(number int) []fyne.CanvasObject {
	dots := make([]fyne.CanvasObject, number)
	for i := 0; i < number; i++ {
		dots[i] = canvas.NewCircle(color.Black)
	}
	return dots
}

func main() {
	if len(os.Args) < 2 {
		log.Fatalf("Argument is required. %s", argConstraint)
	}
	arg := os.Args[1]
	number, err := strconv.Atoi(arg)
	if err != nil || number < 1 || number > 6 {
		log.Fatal(argConstraint)
	}

	a := app.New()
	w := a.NewWindow(fmt.Sprintf("Die face: %d", number))

	layout, ok := faceLayouts[number]
	if !ok {
		log.Fatalf("No layout for number %d", number)
	}

	face := container.New(layout, makeDots(number)...)
	bg := canvas.NewRectangle(color.White)

	w.SetContent(container.NewStack(bg, face))
	w.Resize(fyne.NewSize(360, 360))
	w.ShowAndRun()
}
