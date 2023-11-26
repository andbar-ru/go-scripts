package main

import (
	"bytes"
	"fmt"
	"os"
)

type XY [2]int

var (
	// Печатаемые символы ASCII в порядке увеличения кодов.
	characters = []byte(" !\"#$%&'()*+,-./0123456789:;<=>?@ABCDEFGHIJKLMNOPQRSTUVWXYZ[\\]^_`abcdefghijklmnopqrstuvwxyz{|}~")
	// rows и cols выбраны так, чтобы rows*cols >= len(characters) и близко к нему
	rows = 10
	cols = 10
	// Размер ячейки
	cellSize = 10
	// Размеры поля
	sizeX = rows * cellSize
	sizeY = cols * cellSize
	// Координаты байтов
	byte2xy = make(map[byte]XY, len(characters))
)

// init заполняет byte2xy.
func init() {
	for r := 0; r < rows; r++ {
		for c := 0; c < cols; c++ {
			i := r*rows + c
			if i >= len(characters) {
				return
			}
			xy := XY{c*cols + cellSize/2, r*rows + cellSize/2}
			byte2xy[characters[i]] = xy
		}
	}
}

func genSvg(points []XY) string {
	buf := new(bytes.Buffer)
	buf.WriteString(fmt.Sprintf(`<svg version="1.0" xmlns="http://www.w3.org/2000/svg" width="%dpx" height="%dpx">`, sizeX, sizeY))
	buf.WriteByte('\n')
	buf.WriteString("<polyline points=\"")

	for i, p := range points {
		if i != 0 {
			buf.WriteByte(' ')
		}
		buf.WriteString(fmt.Sprintf("%d %d", p[0], p[1]))
	}

	buf.WriteString(`" fill="none" stroke="black" stroke-width="1" stroke-linecap="round" />`)
	buf.WriteByte('\n')
	buf.WriteString("</svg>\n")

	return buf.String()
}

func main() {
	args := os.Args[1:]
	if len(args) != 1 {
		fmt.Fprintln(os.Stderr, "You must specify phrase as the only argument")
		os.Exit(1)
	}
	phrase := args[0]

	points := make([]XY, 0, len(phrase))

	for i := 0; i < len(phrase); i++ {
		b := phrase[i]
		xy, ok := byte2xy[b]
		if !ok {
			fmt.Fprintf(os.Stderr, "Unallowed character %q in phrase. Printable ASCII characters are allowed only.\n", b)
			os.Exit(1)
		}
		points = append(points, xy)
	}

	svg := genSvg(points)

	fmt.Println(svg)
}
