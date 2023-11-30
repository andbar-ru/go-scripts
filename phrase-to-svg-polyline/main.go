package main

import (
	"bytes"
	"fmt"
	"math"
	"os"
)

// Полярные координаты
type PolarCoords struct {
	// радиус
	r int
	// угол в радианах
	phi float64
}

// Прямоугольные координаты для svg
type SvgCoords [2]float64

const debug = true

var (
	// Печатаемые символы ASCII в порядке увеличения кода.
	characters = []byte(" !\"#$%&'()*+,-./0123456789:;<=>?@ABCDEFGHIJKLMNOPQRSTUVWXYZ[\\]^_`abcdefghijklmnopqrstuvwxyz{|}~")
	// Отображение символа на угол в радианах
	char2phi = make(map[byte]float64, len(characters))
	// Базовый радиус в пикселях
	r0 = 100
	// Инкремент на каждый следующее появляение символа в пикселях
	inc = 10
)

// init заполняет char2phi, равномерно распределяя углы по окружности.
func init() {
	l := len(characters)
	for i, c := range characters {
		char2phi[c] = 2.0 * math.Pi / float64(l) * float64(i)
	}
}

// Генерирует и возвращает svg-код.
// Точки points преобразуются в полилинию, а максимальный радиус rMax нужен, чтобы вычислить
// размеры документа и уточнить координаты точек.
func genSvg(points []SvgCoords, rMax int) string {
	buf := new(bytes.Buffer)
	padding := 5
	size := rMax*2 + padding*2
	buf.WriteString(fmt.Sprintf(`<svg version="1.1" xmlns="http://www.w3.org/2000/svg" width="%dpx" height="%dpx" viewBox="0 0 %d %d" style="background: white">`, size, size, size, size))
	buf.WriteByte('\n')
	// Белый фон вручную. Inkscape, например, игнорирует `background: white` в <svg>.
	buf.WriteString(fmt.Sprintf(`<rect x="0" y="0" width="%dpx" height="%dpx" fill="white" />`, size, size))
	buf.WriteByte('\n')
	buf.WriteString("<polyline points=\"")

	for i, p := range points {
		if i != 0 {
			buf.WriteByte(' ')
		}
		buf.WriteString(fmt.Sprintf("%v %v", p[0]+float64(padding), p[1]+float64(padding)))
	}

	buf.WriteString(`" fill="none" stroke="black" stroke-width="0.5" stroke-linecap="round" />`)

	if debug {
		buf.WriteString(fmt.Sprintf(`<rect x="%d" y="%d" width="1px" height="1px" fill="red" />`, size/2, size/2))
		buf.WriteString(fmt.Sprintf(`<circle cx="%d" cy="%d" r="%d" fill-opacity="0" stroke="red" stroke-width="0.1" />`, size/2, size/2, r0))
	}

	buf.WriteByte('\n')
	buf.WriteString("</svg>\n")

	return buf.String()
}

func polar2svg(polar PolarCoords, rMax int) SvgCoords {
	x := float64(polar.r)*math.Cos(polar.phi) + float64(rMax)
	y := -(float64(polar.r) * math.Sin(polar.phi)) + float64(rMax)
	return SvgCoords{x, y}
}

func main() {
	args := os.Args[1:]
	if len(args) != 1 {
		fmt.Fprintln(os.Stderr, "You must specify phrase as the only argument")
		os.Exit(1)
	}
	phrase := args[0]

	char2count := make(map[byte]int)
	polarPoints := make([]PolarCoords, 0, len(phrase))
	rMax := r0

	for i := 0; i < len(phrase); i++ {
		char := phrase[i]
		count := char2count[char]
		char2count[char]++
		r := r0 + count*inc
		if r > rMax {
			rMax = r
		}
		phi, ok := char2phi[char]
		if !ok {
			fmt.Fprintf(os.Stderr, "Unallowed character %q in phrase. Printable ASCII characters are allowed only. Skipped.\n", char)
			continue
		}
		pPoint := PolarCoords{r: r, phi: phi}
		polarPoints = append(polarPoints, pPoint)
	}

	points := make([]SvgCoords, 0, len(polarPoints))
	for _, pPoint := range polarPoints {
		points = append(points, polar2svg(pPoint, rMax))
	}

	svg := genSvg(points, rMax)

	fmt.Println(svg)
}
