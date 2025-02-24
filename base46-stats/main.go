package main

import (
	"bufio"
	"cmp"
	"errors"
	"fmt"
	"log"
	"math"
	"os"
	"path"
	"regexp"
	"slices"
	"strconv"
	"strings"
)

var (
	base46ThemesDir  = path.Join(os.Getenv("HOME"), ".local/share/nvim/lazy/base46/lua/base46/themes")
	colorRgx         = regexp.MustCompile(`^#[0-9a-f]{6}$`)
	colorCodeRgx     = regexp.MustCompile(`^"#[0-9a-fA-F]{6}"`)
	colorNameRgx     = regexp.MustCompile(`^M\.base_30\.\w+`)
	syntaxColorNames = map[string]struct{}{
		"base08": {},
		"base09": {},
		"base0A": {},
		"base0B": {},
		"base0C": {},
		"base0D": {},
		"base0E": {},
		"base0F": {},
	}
)

// rgb возвращает значения rgb в диапазоне от 0 до 1 для цвета, заданного в hex-формате.
func rgb(color string) ([3]float64, error) {
	var rgb [3]float64

	if !colorRgx.MatchString(color) {
		return rgb, fmt.Errorf("invalid hex color: %q", color)
	}
	rr, _ := strconv.ParseInt(color[1:3], 16, 64)
	gg, _ := strconv.ParseInt(color[3:5], 16, 64)
	bb, _ := strconv.ParseInt(color[5:7], 16, 64)
	rgb[0] = float64(rr) / 255
	rgb[1] = float64(gg) / 255
	rgb[2] = float64(bb) / 255

	return rgb, nil
}

// luminance (perceived brightness) вычисляет относительную яркость цвета.
// Процесс вычисления подробно можно посмотреть на странице https://www.101computing.net/colour-luminance-and-contrast-ratio/.
func luminance(color string) (float64, error) {
	rgb, err := rgb(color)
	if err != nil {
		return 0.0, err
	}

	for i := range rgb {
		if rgb[i] <= 0.03928 {
			rgb[i] /= 12.92
		} else {
			rgb[i] = math.Pow((rgb[i]+0.055)/1.055, 2.4)
		}
	}

	l := 0.2126*rgb[0] + 0.7152*rgb[1] + 0.0722*rgb[2]

	return l, nil
}

// contrast вычисляет контраст двух цветов.
// Формулу взял здесь: https://www.101computing.net/colour-luminance-and-contrast-ratio/
func contrast(color1, color2 string) (float64, error) {
	l1, err := luminance(color1)
	if err != nil {
		return 0, err
	}
	l2, err := luminance(color2)
	if err != nil {
		return 0, err
	}

	if l1 == l2 {
		return 1, nil
	}
	if l1 > l2 {
		l1, l2 = l2, l1
	}

	c := (l2 + 0.05) / (l1 + 0.05)

	return c, nil
}

// hue вычисляет тон цвета и возвращает в градусах в диапазоне [0, 360).
// Формулу взял здесь: https://www.rapidtables.com/convert/color/rgb-to-hsl.html
func hue(color string) (float64, error) {
	rgb, err := rgb(color)
	if err != nil {
		return 0, err
	}

	var max, min = 0.0, 1.0
	for _, c := range rgb {
		if c > max {
			max = c
		}
		if c < min {
			min = c
		}
	}
	delta := max - min

	var hue float64

	switch {
	case delta == 0:
		hue = 0
	case max == rgb[0]:
		hue = (rgb[1] - rgb[2]) / delta
	case max == rgb[1]:
		hue = (rgb[2]-rgb[0])/delta + 2
	case max == rgb[2]:
		hue = (rgb[0]-rgb[1])/delta + 4
	default:
		return 0, fmt.Errorf("unexpected max = %v when rgb is %v", max, rgb)
	}

	hue *= 60
	if hue < 0 {
		hue += 360
	}

	return hue, nil
}

func avg(values []float64) float64 {
	var sum float64
	for _, v := range values {
		sum += v
	}
	avg := sum / float64(len(values))
	return avg
}

func std(values []float64) float64 {
	avg := avg(values)
	var sumSqr float64
	for _, v := range values {
		sumSqr += (v - avg) * (v - avg)
	}
	std := math.Sqrt(sumSqr / float64(len(values)))
	return std
}

type contrastStats struct {
	min float64
	max float64
	avg float64
	std float64
}

type theme struct {
	name       string
	type_      string
	baseColors struct {
		bg string
		fg string
	}
	syntaxColors []string
}

func (t theme) baseContrast() (float64, error) {
	c, err := contrast(t.baseColors.bg, t.baseColors.fg)
	if err != nil {
		return 0, err
	}

	return c, nil
}

func (t theme) syntaxContrasts() ([]float64, error) {
	contrasts := make([]float64, 0, len(t.syntaxColors))
	for _, color := range t.syntaxColors {
		c, err := contrast(color, t.baseColors.bg)
		if err != nil {
			return nil, err
		}
		contrasts = append(contrasts, c)
	}

	return contrasts, nil
}

func (t theme) syntaxContrastStats() (contrastStats, error) {
	var zeroStats contrastStats
	contrasts, err := t.syntaxContrasts()
	if err != nil {
		return zeroStats, err
	}

	var stats contrastStats

	var min, max float64 = math.Inf(1), 0
	for _, c := range contrasts {
		if c < min {
			min = c
		}
		if c > max {
			max = c
		}
	}

	stats.min = min
	stats.max = max
	stats.avg = avg(contrasts)
	stats.std = std(contrasts)

	return stats, nil
}

func (t theme) syntaxColorsHueIncStd() (float64, error) {
	// Убираем дубликаты из syntaxColors
	uniqSyntaxColors := make([]string, 0, len(t.syntaxColors))
	colorsMap := make(map[string]struct{}, len(t.syntaxColors))
	for _, c := range t.syntaxColors {
		if _, ok := colorsMap[c]; !ok {
			uniqSyntaxColors = append(uniqSyntaxColors, c)
		}
		colorsMap[c] = struct{}{}
	}

	hues := make([]float64, 0, len(uniqSyntaxColors))
	for _, c := range uniqSyntaxColors {
		h, err := hue(c)
		if err != nil {
			return 0, err
		}
		hues = append(hues, h)
	}

	slices.Sort(hues)

	hues = append(hues, hues[0]+360)

	incs := make([]float64, 0, len(hues))
	for i := 0; i < len(hues)-1; i++ {
		incs = append(incs, hues[i+1]-hues[i])
	}

	std := std(incs)

	return std, nil
}

func (t theme) score() (int, error) {
	score := 0

	baseContrast, err := t.baseContrast()
	if err != nil {
		return 0, err
	}
	if baseContrast < 4.5 {
		score += 0
	} else if baseContrast < 7 {
		score += 1
	} else if baseContrast < 9.5 {
		score += 2
	} else {
		score += 3
	}

	syntaxContrastStats, err := t.syntaxContrastStats()
	if err != nil {
		return 0, err
	}

	minSyntaxContrast := syntaxContrastStats.min
	if minSyntaxContrast < 4.5 {
		score += 0
	} else if minSyntaxContrast < 7 {
		score += 1
	} else if minSyntaxContrast < 9.5 {
		score += 2
	} else {
		score += 3
	}

	avgSyntaxContrast := syntaxContrastStats.avg
	if avgSyntaxContrast < 4.5 {
		score += 0
	} else if avgSyntaxContrast < 7 {
		score += 1
	} else if avgSyntaxContrast < 9.5 {
		score += 2
	} else {
		score += 3
	}

	syntaxContrastStd := syntaxContrastStats.std
	if syntaxContrastStd < 1 {
		score += 3
	} else if syntaxContrastStd < 2 {
		score += 2
	} else if syntaxContrastStd < 3 {
		score += 1
	} else {
		score += 0
	}

	hueIncStd, err := t.syntaxColorsHueIncStd()
	if err != nil {
		return 0, err
	}
	if hueIncStd < 20 {
		score += 3
	} else if hueIncStd < 40 {
		score += 2
	} else if hueIncStd < 60 {
		score += 1
	} else {
		score += 0
	}

	return score, nil
}

func parseThemeFile(filepath string) (theme, error) {
	var zeroTheme theme
	var parsedTheme theme

	file, err := os.Open(filepath)
	if err != nil {
		return zeroTheme, err
	}
	defer file.Close()

	parsedTheme.name = strings.TrimSuffix(path.Base(filepath), path.Ext(filepath))

	base30 := make(map[string]string)
	var mode string // base30, base16

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		text := strings.TrimSpace(strings.ReplaceAll(scanner.Text(), "'", "\""))
		switch {
		case strings.HasPrefix(text, "--"): // lua comment
			continue
		case text == "M.base_30 = {":
			mode = "base30"
		case text == "M.base_16 = {":
			mode = "base16"
		case text == "}":
			mode = ""
		case strings.HasPrefix(text, "M.type = "):
			parts := strings.SplitN(text, "=", 2)
			if len(parts) != 2 {
				return zeroTheme, fmt.Errorf("invalid type text: %q", text)
			}
			subParts := strings.SplitN(parts[1], "\"", 3)
			if len(subParts) != 3 {
				return zeroTheme, fmt.Errorf("invalid type text: %q", text)
			}
			type_ := subParts[1]
			switch type_ {
			case "light":
				parsedTheme.type_ = "light"
			case "dark":
				parsedTheme.type_ = "dark"
			default:
				return zeroTheme, errors.New("invalid type")
			}
		case mode == "base30":
			parts := strings.SplitN(text, "=", 2)
			if len(parts) != 2 {
				return zeroTheme, fmt.Errorf("invalid base30 entry: %q", text)
			}
			name := strings.TrimSpace(parts[0])
			subParts := strings.SplitN(parts[1], "\"", 3)
			if len(subParts) != 3 {
				return zeroTheme, fmt.Errorf("invalid base30 entry: %q", text)
			}
			color := strings.ToLower(subParts[1])
			if !colorRgx.MatchString(color) {
				return zeroTheme, fmt.Errorf("invalid color in base30 entry: %q", text)
			}
			base30[name] = color
		case mode == "base16":
			parts := strings.SplitN(text, "=", 2)
			if len(parts) != 2 {
				return zeroTheme, fmt.Errorf("invalid base16 entry: %q", text)
			}
			name := strings.TrimSpace(parts[0])
			colorPart := strings.TrimSpace(parts[1])
			var color string

			switch {
			case colorCodeRgx.MatchString(colorPart):
				subParts := strings.SplitN(colorPart, "\"", 3)
				if len(subParts) != 3 {
					return zeroTheme, fmt.Errorf("invalid color in base16 entry: %q", text)
				}
				color = strings.ToLower(subParts[1])
				if !colorRgx.MatchString(color) {
					return zeroTheme, fmt.Errorf("invalid color in base16 entry: %q", text)
				}
			case colorNameRgx.MatchString(colorPart):
				colorName := colorNameRgx.FindString(colorPart)
				if colorName == "" {
					return zeroTheme, fmt.Errorf("invalid color in base16 entry: %q", text)
				}
				subParts := strings.Split(colorName, ".")
				if len(subParts) != 3 {
					return zeroTheme, fmt.Errorf("invalid color in base16 entry: %q", text)
				}
				var ok bool
				color, ok = base30[subParts[2]]
				if !ok {
					return zeroTheme, fmt.Errorf("invalid color in base16 entry: %q", text)
				}
			default:
				return zeroTheme, fmt.Errorf("unexpected base16 entry: %q", text)
			}

			_, isSyntaxColor := syntaxColorNames[name]

			switch {
			case name == "base00":
				parsedTheme.baseColors.bg = color
			case name == "base05":
				parsedTheme.baseColors.fg = color
			case isSyntaxColor:
				parsedTheme.syntaxColors = append(parsedTheme.syntaxColors, color)
			default:
			}
		default:
		}
	}

	if err := scanner.Err(); err != nil {
		return zeroTheme, err
	}

	if parsedTheme.name == "" {
		return zeroTheme, errors.New("parsedTheme has no name")
	}
	if parsedTheme.type_ == "" {
		return zeroTheme, errors.New("parsedTheme has no type")
	}
	if parsedTheme.baseColors.bg == "" || parsedTheme.baseColors.fg == "" {
		return zeroTheme, errors.New("parsedTheme.baseColors is not complete")
	}
	if len(parsedTheme.syntaxColors) != 8 {
		return zeroTheme, fmt.Errorf("parsedTheme.syntaxColors size is not 8, just %d", len(parsedTheme.syntaxColors))
	}

	return parsedTheme, nil
}

func main() {
	type entry struct {
		name                string
		baseContrast        float64
		syntaxContrastStats contrastStats
		hueIncStd           float64
		score               int
	}

	lightThemes := make([]entry, 0)
	darkThemes := make([]entry, 0)

	files, err := os.ReadDir(base46ThemesDir)
	if err != nil {
		panic(err)
	}

	var maxLightThemeNameLen, maxDarkThemeNameLen int

	for _, file := range files {
		if file.IsDir() {
			continue
		}
		filepath := path.Join(base46ThemesDir, file.Name())
		parsedTheme, err := parseThemeFile(filepath)
		if err != nil {
			log.Panicf("%s: %v", filepath, err)
		}
		baseContrast, err := parsedTheme.baseContrast()
		if err != nil {
			panic(err)
		}
		syntaxContrastStats, err := parsedTheme.syntaxContrastStats()
		if err != nil {
			panic(err)
		}
		hueIncStd, err := parsedTheme.syntaxColorsHueIncStd()
		if err != nil {
			panic(err)
		}
		score, err := parsedTheme.score()
		if err != nil {
			panic(err)
		}

		name := parsedTheme.name

		if parsedTheme.type_ == "light" {
			if len(name) > maxLightThemeNameLen {
				maxLightThemeNameLen = len(name)
			}
			lightThemes = append(lightThemes, entry{name, baseContrast, syntaxContrastStats, hueIncStd, score})
		} else if parsedTheme.type_ == "dark" {
			if len(name) > maxDarkThemeNameLen {
				maxDarkThemeNameLen = len(name)
			}
			darkThemes = append(darkThemes, entry{name, baseContrast, syntaxContrastStats, hueIncStd, score})
		} else {
			log.Panicf("unexpected theme type: %q", parsedTheme.type_)
		}
	}

	slices.SortFunc(lightThemes, func(a, b entry) int {
		return cmp.Compare(a.score, b.score)
	})
	slices.SortFunc(darkThemes, func(a, b entry) int {
		return cmp.Compare(a.score, b.score)
	})

	fmt.Println("--------------------------------------------------")
	fmt.Println("Light themes sorted by score asc:")
	fmt.Println("name | base contrast | min syntax contrast | avg syntax contrast | syntax contrast std | hue inc std | score")
	fmt.Println("--------------------------------------------------")
	for _, entry := range lightThemes {
		fmt.Printf("%-*s: %5.2f %8.2f %8.2f %8.2f %8.2f %10d\n",
			maxLightThemeNameLen,
			entry.name,
			entry.baseContrast,
			entry.syntaxContrastStats.min,
			entry.syntaxContrastStats.avg,
			entry.syntaxContrastStats.std,
			entry.hueIncStd,
			entry.score,
		)
	}

	fmt.Println()
	fmt.Println("--------------------------------------------------")
	fmt.Println("Dark themes sorted by score asc:")
	fmt.Println("name | base contrast | min syntax contrast | avg syntax contrast | syntax contrast std | hue inc std | score")
	fmt.Println("--------------------------------------------------")
	for _, entry := range darkThemes {
		fmt.Printf("%-*s: %5.2f %8.2f %8.2f %8.2f %8.2f %10d\n",
			maxDarkThemeNameLen,
			entry.name,
			entry.baseContrast,
			entry.syntaxContrastStats.min,
			entry.syntaxContrastStats.avg,
			entry.syntaxContrastStats.std,
			entry.hueIncStd,
			entry.score,
		)
	}
}
