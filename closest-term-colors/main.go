/*
closest-term-colors outputs `n` (default 1) term color codes closest to the color specified via
command-line argument.
*/
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"

	"github.com/andbar-ru/closest_colors"
)

const (
	colorsUrl = "https://jonasjacek.github.io/colors/data.json"
)

var (
	colors                []color
	colorRegex            = regexp.MustCompile(`(?i)^#([0-9a-f]{2})([0-9a-f]{2})([0-9a-f]{2})$`)
	numberRegex           = regexp.MustCompile(`^\d+$`)
	srcRgb                rgb
	numberOfClosestColors = 1
)

type rgb struct {
	Red   uint8 `json:"r"`
	Green uint8 `json:"g"`
	Blue  uint8 `json:"b"`
}

type color struct {
	Id   int    `json:"colorId"`
	Rgb  rgb    `json:"rgb"`
	Hex  string `json:"hexString"`
	Name string `json:"name"`
}

func check(err error) {
	if err != nil {
		log.Panic(err)
	}
}

func printHelpAndExit() {
	fmt.Printf("Usage: %s <color> [<number of closest colors to print>]\n", filepath.Base(os.Args[0]))
	fmt.Println("color must be given in hex format, e.g. '#131723'")
	fmt.Printf("number of closest colors is optional (1 by default), must be integer and less than number of available colors (%d).\n", len(colors))
	os.Exit(1)
}

func fillColors() {
	response, err := http.Get(colorsUrl)
	check(err)
	defer response.Body.Close()
	if response.StatusCode != 200 {
		log.Panicf("%s: status code error: %d %s", colorsUrl, response.StatusCode, response.Status)
	}
	err = json.NewDecoder(response.Body).Decode(&colors)
	check(err)
	if len(colors) != 256 {
		log.Panicf("Number of colors must be 256, but got %d", len(colors))
	}
}

func validateArgs() {
	args := os.Args[1:]
	if len(args) < 1 || len(args) > 2 {
		printHelpAndExit()
	}

	colorArg := args[0]
	if !colorRegex.MatchString(colorArg) {
		fmt.Printf("ERROR: color does not match regexp %s\n\n", colorRegex)
		printHelpAndExit()
	}
	submatches := colorRegex.FindStringSubmatch(colorArg)
	red, err := strconv.ParseUint(submatches[1], 16, 8)
	check(err)
	green, err := strconv.ParseUint(submatches[2], 16, 8)
	check(err)
	blue, err := strconv.ParseUint(submatches[3], 16, 8)
	check(err)

	srcRgb = rgb{Red: uint8(red), Green: uint8(green), Blue: uint8(blue)}

	if len(args) == 2 {
		numberArg := args[1]
		if !numberRegex.MatchString(numberArg) {
			fmt.Printf("ERROR: number does not match regexp %s\n\n", numberRegex)
			printHelpAndExit()
		}
		numberStr := numberRegex.FindStringSubmatch(numberArg)[0]
		numberOfClosestColors, err = strconv.Atoi(numberStr)
		if numberOfClosestColors > len(colors) {
			fmt.Printf("ERROR: number is greater than number of available colors: %d > %d", numberOfClosestColors, len(colors))
			printHelpAndExit()
		}
		check(err)
	}
}

func (rgb rgb) RGB() (uint8, uint8, uint8) {
	return rgb.Red, rgb.Green, rgb.Blue
}

func (c color) RGB() (uint8, uint8, uint8) {
	return c.Rgb.Red, c.Rgb.Green, c.Rgb.Blue
}

func init() {
	fillColors()
}

func main() {
	validateArgs()

	rgbColors := make([]closest_colors.RGBColor, len(colors))
	for i, c := range colors {
		rgbColors[i] = c
	}

	results, err := closest_colors.FindClosestRGBColors(srcRgb, numberOfClosestColors, rgbColors)
	check(err)
	for _, result := range results {
		resultColor := result.Color.(color)
		fmt.Printf("%d (%s: %s) (distance %.2f)\n", resultColor.Id, resultColor.Hex, resultColor.Name, result.Distance)
	}
}
