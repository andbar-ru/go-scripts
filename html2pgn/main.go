package main

import (
	"fmt"
	"os"
	"strings"

	"golang.org/x/net/html"
)

const (
	headingClass         = "Heading_20_3"
	movesClass           = "Standard"
	predefinedMovesClass = "T1"
	noveltyClass         = "T2"
)

var (
	inHeading         = false
	inMoves           = false
	inPredefinedMoves = false
	inNovelty         = false
	round             int
	currentResult     string
)

var (
	commonTags = `[Event "Stockfish vs Stockfish games"]
[Site "localhost"]
[White "Stockfish 13"]
[Black "Stockfish 13"]
[TimeControl "1800+0"]
`
	// Date, Round, Result, ECO tags are dymamic.

	pgn string
)

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

func getClass(token html.Token) string {
	for _, attr := range token.Attr {
		if attr.Key == "class" {
			return attr.Val
		}
	}
	return ""
}

func fillTags(heading string) {
	pgn += commonTags
	fields := strings.Fields(heading)
	if len(fields) < 4 {
		return
	}
	round++
	pgn += fmt.Sprintf("[Round \"%d\"]\n", round)
	pgn += fmt.Sprintf("[ECO \"%s\"]\n", fields[1])

	result := fields[2]
	result = strings.Replace(result, ":", "-", 1)
	result = strings.ReplaceAll(result, "½", "1/2")
	pgn += fmt.Sprintf("[Result \"%s\"]\n", result)
	currentResult = result

	date := fields[3]
	date = strings.Replace(date, "(", "", 1)
	date = strings.Replace(date, ")", "", 1)
	date = strings.ReplaceAll(date, "-", ".")
	pgn += fmt.Sprintf("[Date \"%s\"]\n", date)

	pgn += "\n"
}

func main() {
	if len(os.Args) < 2 {
		panic("you must give file path as the first argument")
	}
	filePath := os.Args[1]
	file, err := os.Open(filePath)
	checkErr(err)
	tokenizer := html.NewTokenizer(file)
L:
	for {
		tokenType := tokenizer.Next()
		switch tokenType {
		case html.ErrorToken:
			break L
		case html.StartTagToken:
			switch class := getClass(tokenizer.Token()); class {
			case headingClass:
				inHeading = true
			case movesClass:
				inMoves = true
			case predefinedMovesClass:
				inPredefinedMoves = true
			case noveltyClass:
				inNovelty = true
			}
		case html.EndTagToken:
			tagNameRaw, _ := tokenizer.TagName()
			tagName := string(tagNameRaw)
			switch tagName {
			case "p":
				switch {
				case inHeading:
					inHeading = false
				case inMoves:
					inMoves = false
					pgn += " " + currentResult + "\n\n"
				}
			case "span":
				if inPredefinedMoves {
					inPredefinedMoves = false
					pgn += " {end of predefined moves}"
				}
				if inNovelty {
					inNovelty = false
					pgn += " {novelty}"
				}
			}
		case html.TextToken:
			switch {
			case inHeading:
				heading := string(tokenizer.Text())
				fillTags(heading)
			case inMoves:
				pgn += string(tokenizer.Text())
			}
		}
	}

	fmt.Println(pgn)
}
