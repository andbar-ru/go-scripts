package main

import (
	"bufio"
	"fmt"
	"image"
	"log"
	"os"

	"github.com/andbar-ru/go-scripts/average_color"
)

func check(err error) {
	if err != nil {
		log.Panic(err)
	}
}

func main() {
	if len(os.Args) != 2 {
		fmt.Printf("usage: %s <image path>\n", os.Args[0])
		os.Exit(1)
	}

	imagePath := os.Args[1]

	if _, err := os.Stat(imagePath); os.IsNotExist(err) {
		fmt.Printf("file %s doen not exist\n", imagePath)
		os.Exit(1)
	}

	f, err := os.Open(imagePath)
	check(err)
	defer f.Close()
	img, _, err := image.Decode(bufio.NewReader(f))
	check(err)

	averageColor := average_color.AverageColor(img)

	if averageColor.A == 0xff {
		fmt.Printf("#%02x%02x%02x\n", averageColor.R, averageColor.G, averageColor.B)
	} else {
		fmt.Printf("#%02x%02x%02x%02x\n", averageColor.R, averageColor.G, averageColor.B, averageColor.A)
	}
}
