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

func newDot() *canvas.Circle {
	return canvas.NewCircle(color.Black)
}

var number2layout = map[int]fyne.Layout{
	1: new(oneDotLayout),
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
	w := a.NewWindow(fmt.Sprintf("Die face: %d dots", number))

	layout := number2layout[number]
	if layout == nil {
		log.Fatalf("No layout for number %d", number)
	}

	face := container.New(layout, newDot())
	bg := canvas.NewRectangle(color.White)

	w.SetContent(container.NewStack(bg, face))
	w.Resize(fyne.NewSize(360, 360))
	w.ShowAndRun()
}
