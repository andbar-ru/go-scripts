package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"
)

const (
	CHESSMATCHESDIR = "CHESSMATCHESDIR"
)

var (
	chessmatchesDir = os.Getenv(CHESSMATCHESDIR)
	tagRgx          = regexp.MustCompile(`^\[([[:alpha:]]+) "([^"]*)"\]$`)
	tags            = map[string]string{
		"Date":   "",
		"White":  "",
		"Black":  "",
		"Result": "",
		"ECO":    "",
	}
)

func main() {
	if len(os.Args) < 2 {
		log.Panic("You must specify path to pgn file as the first argument.")
	}
	pgnFilePath := os.Args[1]
	pgnFile, err := os.Open(pgnFilePath)
	if err != nil {
		log.Panic(err)
	}
	defer pgnFile.Close()

	if chessmatchesDir == "" {
		log.Panicf("env var %s is not defined\n", CHESSMATCHESDIR)
	}
	fi, err := os.Stat(chessmatchesDir)
	if err != nil {
		log.Panic(err)
	}
	if !fi.Mode().IsDir() {
		log.Panicf("%s=%q is not directory\n", CHESSMATCHESDIR, chessmatchesDir)
	}

	nTags := 0

	scanner := bufio.NewScanner(pgnFile)
	for scanner.Scan() {
		match := tagRgx.FindStringSubmatch(scanner.Text())
		if match == nil {
			break
		}
		if len(match) != 3 {
			log.Panicf("%q: number of items in tag is less than 2\n", match[0])
		}
		key, value := match[1], match[2]
		if v, ok := tags[key]; ok {
			if v == "" {
				nTags++
			}
			tags[key] = value
		}
	}

	if nTags != len(tags) {
		log.Panicf("Expected %d tags, got %d\n", len(tags), nTags)
	}

	white := strings.ReplaceAll(tags["White"], " ", "")
	black := strings.ReplaceAll(tags["Black"], " ", "")

	var gamesFilePath string
	create := false

	whiteBlackPath := filepath.Join(chessmatchesDir, fmt.Sprintf("%s_vs_%s", white, black))
	blackWhitePath := filepath.Join(chessmatchesDir, fmt.Sprintf("%s_vs_%s", black, white))
	if fi, err = os.Stat(whiteBlackPath); err == nil {
		gamesFilePath = whiteBlackPath
	} else if fi, err = os.Stat(blackWhitePath); err == nil {
		gamesFilePath = blackWhitePath
	} else {
		c := strings.Compare(white, black)
		if c < 0 {
			gamesFilePath = whiteBlackPath
		} else {
			gamesFilePath = blackWhitePath
		}
		create = true
	}

	flag := os.O_RDWR
	if create {
		flag |= os.O_CREATE | os.O_EXCL
	}
	gamesFile, err := os.OpenFile(gamesFilePath, flag, 0644)
	if err != nil {
		log.Panic(err)
	}
	defer gamesFile.Close()

	var players [2]string
	scanner = bufio.NewScanner(gamesFile)
	bytesRead := 0
	var totalDoubled [2]int

	// First line
	var firstLine string
	if create {
		if gamesFilePath == whiteBlackPath {
			players[0], players[1] = white, black
		} else if gamesFilePath == blackWhitePath {
			players[0], players[1] = black, white
		}
		firstLine = fmt.Sprintf("%s %s %s\n", strings.Repeat(" ", 17), players[0], players[1])
		_, err = gamesFile.WriteString(firstLine)
		if err != nil {
			log.Panic(err)
		}
	} else {
		if !scanner.Scan() {
			log.Panicf("file %s is empty but not created\n", gamesFilePath)
		}
		firstLine = scanner.Text()
		if firstLine == fmt.Sprintf("%s %s %s", strings.Repeat(" ", 17), white, black) {
			players[0], players[1] = white, black
		} else if firstLine == fmt.Sprintf("%s %s %s", strings.Repeat(" ", 17), black, white) {
			players[0], players[1] = black, white
		} else {
			log.Panicf("%q: incorrect first line\n", firstLine)
		}
	}
	if players[0] == "" || players[1] == "" {
		log.Panicln("firstPlayer and/or secondPlayer are undefined")
	}
	bytesRead += len(firstLine) + 1

	// Past game lines
	if !create {
		gameRgx := regexp.MustCompile(`^\d{4}-\d{2}-\d{2} [A-E]\d{2} (?:WB|BW)` + strings.Repeat(" ", utf8.RuneCountInString(players[0])) + `([½10])` + strings.Repeat(" ", utf8.RuneCountInString(players[1])) + `([½01])$`)

		for scanner.Scan() {
			text := scanner.Text()

			if text == "" {
				break
			}
			bytesRead += len(text) + 1

			match := gameRgx.FindStringSubmatch(text)
			if len(match) != 3 {
				log.Panicf("%q: game line is in incorrect format\n", text)
			}
			var resultDoubled [2]int
			for i, m := range match[1:] {
				switch m {
				case "½":
					resultDoubled[i] = 1
				case "1":
					resultDoubled[i] = 2
				}
			}
			if resultDoubled[0]+resultDoubled[1] != 2 {
				log.Panicf("%q: incorrect result\n", text)
			}
			totalDoubled[0] += resultDoubled[0]
			totalDoubled[1] += resultDoubled[1]
		}

		err := gamesFile.Truncate(int64(bytesRead))
		if err != nil {
			log.Panic(err)
		}
		_, err = gamesFile.Seek(0, os.SEEK_END)
		if err != nil {
			log.Panic(err)
		}
	}

	// New game line
	t, err := time.Parse("2006.01.02", tags["Date"])
	if err != nil {
		log.Panic(err)
	}
	date := t.Format(time.DateOnly)
	eco := tags["ECO"]
	order := "WB"
	if players[0] == black {
		order = "BW"
	}
	result := strings.Split(tags["Result"], "-")
	if len(result) != 2 {
		log.Panicf("Incorrect result: %q\n", tags["Result"])
	}
	if order == "BW" {
		result[0], result[1] = result[1], result[0]
	}
	for i, v := range result {
		if v == "1/2" {
			result[i] = "½"
		}
	}
	var resultDoubled [2]int
	for i, r := range result {
		switch r {
		case "½":
			resultDoubled[i] = 1
		case "1":
			resultDoubled[i] = 2
		}
	}
	if resultDoubled[0]+resultDoubled[1] != 2 {
		log.Panicf("%q: incorrect result\n", tags["Result"])
	}
	totalDoubled[0] += resultDoubled[0]
	totalDoubled[1] += resultDoubled[1]
	gameLine := fmt.Sprintf("%s %s %s%*s%*s\n", date, eco, order, utf8.RuneCountInString(players[0])+1, result[0], utf8.RuneCountInString(players[1])+1, result[1])
	_, err = gamesFile.WriteString(gameLine)
	if err != nil {
		log.Panic(err)
	}

	// Total line
	var totals [2]string
	for i, total := range totalDoubled {
		var halfStr string
		if total%2 == 1 {
			halfStr = "½"
		}
		var integerStr string
		integer := total / 2
		if integer > 0 || halfStr == "" {
			integerStr = fmt.Sprintf("%d", integer)
		}
		totals[i] = fmt.Sprintf("%s%s", integerStr, halfStr)
	}
	totalLine := fmt.Sprintf("\nTotal%*s%*s\n", 12+utf8.RuneCountInString(players[0])+1, totals[0], utf8.RuneCountInString(players[1])+1, totals[1])
	_, err = gamesFile.WriteString(totalLine)
	if err != nil {
		log.Panic(err)
	}

	fmt.Println("SUCCESS")
}
