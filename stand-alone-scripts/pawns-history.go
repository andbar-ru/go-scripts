/*
Objectives:
* Parse pgn files given in command line arguments. +
* Determine which moves are pawn moves. +
* Determine which moves capture a pawn.
* Statistics:
  - estimate average fraction of pawn moves of all moves; +
  - estimate correlation between fraction of pawn moves and number of moves in a game; +
  - average chances to survive for all pawns and for each pawn individually;
  - average chance to promote for all pawns and for each pawn individually;
  - chances to survive if pawn moves first;
  - chances to survive if pawn moves last or doesn't move;
  - balance of kills and deaths for all pawns and for each pawn individually;
  - average number of moves for each pawn;
  - how many moves for one death for each pawn;
*/
package main

import (
	"bufio"
	"fmt"
	"io"
	"math"
	"os"
	"regexp"
	"strconv"
	"strings"
)

const (
	noPromotion = iota
	knight
	bishop
	rook
	queen
)

var (
	commentsRegex      = regexp.MustCompile(`{.*?}`)
	variationsRegex    = regexp.MustCompile(`\(.*?\)`)
	tagsRegex          = regexp.MustCompile(`\[.*?\]`)
	moveRegex          = regexp.MustCompile(`(\d+\.+)?(.*)`)
	moveNumRegex       = regexp.MustCompile(`\d+\.`)
	dotWithSpacesRegex = regexp.MustCompile(`\.\s+`)
	isPawnPlyRegex     = regexp.MustCompile(`^[a-h]`)

	stats = &Stats{}
)

type Stats struct {
	games             int
	allPlies          int
	pawnPlies         int
	gamePliesList     []int
	gamePawnPliesList []int
}

type PgnParser struct {
	reader          *bufio.Reader
	prevLineIsEmpty bool
	err             error
}

type Square struct {
	file uint8
	rank uint8
}

type Pawn struct {
	initSquare Square
	square     Square
	promotion  uint8
	moves      int
	captured   bool
	captures   int
}

type Move [2]string

type Game struct {
	moves []Move
	pawns map[Square]Pawn
}

func getPercent(fraction, total float64) float64 {
	if total == 0 {
		return 0
	}
	return (fraction * 100) / total
}

func getCorrelation(set1, set2 []float64) (float64, error) {
	if len(set1) != len(set2) {
		return 0, fmt.Errorf("Set sizes are not equal: %d != %d", len(set1), len(set2))
	}

	var sum1, sum2, mean1, mean2, cov, s1, s2 float64

	for i := range set1 {
		sum1 += set1[i]
		sum2 += set2[i]
	}

	mean1 = sum1 / float64(len(set1))
	mean2 = sum2 / float64(len(set2))

	for i := range set1 {
		cov += (set1[i] - mean1) * (set2[i] - mean2)
		s1 += math.Pow((set1[i] - mean1), 2)
		s2 += math.Pow((set2[i] - mean2), 2)
	}

	return cov / math.Sqrt(s1*s2), nil
}

func newPgnParser(source io.Reader) (parser *PgnParser) {
	parser = &PgnParser{reader: bufio.NewReader(source)}
	return
}

func moveList(moveText string) ([]Move, error) {
	// remove comments
	text := commentsRegex.ReplaceAllString(moveText, "")
	// remove variations
	text = variationsRegex.ReplaceAllString(text, "")
	// remove tag pairs
	text = tagsRegex.ReplaceAllString(text, "")
	// remove line breaks
	text = strings.Replace(text, "\n", " ", -1)
	// remove spaces after dot
	text = dotWithSpacesRegex.ReplaceAllString(text, ".")

	list := strings.Split(text, " ")
	moves := []Move{}
	for _, ply := range list {
		ply = strings.TrimSpace(ply) // ply - полуход
		switch ply {
		case "*", "1/2-1/2", "1-0", "0-1": // We don't care about outcome
		case "":
		default:
			results := moveRegex.FindStringSubmatch(ply)
			if len(results) == 3 {
				if results[1] != "" {
					moves = append(moves, Move{})
					moves[len(moves)-1][0] = results[2]
				} else {
					moves[len(moves)-1][1] = results[2]
				}
			} else {
				return nil, fmt.Errorf("Unexepected number of results: %d", len(results))
			}
		}
	}

	if err := validateMoves(moves, text); err != nil {
		return nil, err
	}

	return moves, nil
}

func validateMoves(moves []Move, moveText string) error {
	// Assert that the last move number matches the number of moves.
	moveNumsStrs := moveNumRegex.FindAllString(moveText, -1)
	lastMoveNumStr := moveNumsStrs[len(moveNumsStrs)-1]
	lastMoveNumStr = lastMoveNumStr[:len(lastMoveNumStr)-1]
	lastMoveNum, err := strconv.Atoi(lastMoveNumStr)
	if err != nil {
		return err
	}
	if len(moves) != lastMoveNum {
		return fmt.Errorf("Number of moves (%d) does not match the lat move number (%d)", len(moves), lastMoveNum)
	}
	for i, move := range moves {
		if i != len(moves)-1 {
			if move[0] == "" || move[1] == "" {
				return fmt.Errorf("Not last (%d) move contains empty ply: %v", i+1, moves)
			}
		} else {
			if move[0] == "" {
				return fmt.Errorf("Last move is empty: %v", moves)
			}
		}
	}

	return nil
}

func isPawnPly(ply string) bool {
	return isPawnPlyRegex.MatchString(ply)
}

func analyseGame(game *Game) {
	stats.games++
	var gamePlies, gamePawnPlies int

	for _, move := range game.moves {
		for _, ply := range move {
			if ply != "" { // That might be in the last move of a game
				stats.allPlies++
				gamePlies++

				if isPawnPly(ply) {
					stats.pawnPlies++
					gamePawnPlies++
				}
			}
		}
	}
	stats.gamePliesList = append(stats.gamePliesList, gamePlies)
	stats.gamePawnPliesList = append(stats.gamePawnPliesList, gamePawnPlies)
}

func (s *Stats) String() string {
	var output string

	output += fmt.Sprintf("Games: %d\n", s.games)
	output += fmt.Sprintf("All plies: %d\n", s.allPlies)
	output += fmt.Sprintf("Pawn plies: %d (%.1f %%)\n", s.pawnPlies, getPercent(float64(s.pawnPlies), float64(s.allPlies)))
	output += fmt.Sprintf("Correlation between fraction of pawn moves and number of moves in a game: %.4f\n", s.getCorrelationBetweenPawnFractionAndGamePlies())

	return output
}

func (s *Stats) getCorrelationBetweenPawnFractionAndGamePlies() float64 {
	plies := make([]float64, len(s.gamePliesList))
	for i, v := range s.gamePliesList {
		plies[i] = float64(v)
	}
	fractions := make([]float64, len(s.gamePawnPliesList))
	for i, v := range s.gamePawnPliesList {
		fractions[i] = float64(v) / plies[i]
	}
	correlation, err := getCorrelation(plies, fractions)
	if err != nil {
		panic(err)
	}
	return correlation
}

func (parser *PgnParser) hasNextGame() bool {
	return parser.err == nil
}

func (parser *PgnParser) nextGame() (game *Game, err error) {
	var moveText string
	var paragraphCount int

	for paragraphCount < 2 && parser.err == nil {
		line, err := parser.reader.ReadString('\n')
		if err != nil {
			parser.err = err
			if err != io.EOF {
				return nil, err
			}
		}

		if strings.TrimSpace(line) == "" {
			if !parser.prevLineIsEmpty {
				paragraphCount++
			}
			parser.prevLineIsEmpty = true
		} else {
			parser.prevLineIsEmpty = false
			if paragraphCount == 1 {
				moveText += line
			}
		}
	}

	if moveText != "" {
		moves, err := moveList(moveText)
		if err != nil {
			parser.err = err
		} else {
			game = &Game{moves: moves}
		}
	}
	if err == io.EOF {
		err = nil
	}
	return
}

// setUp creates and sets up the pawns.
func (g *Game) setUp() {
	g.pawns = make(map[Square]Pawn)
	// Set up pawns
	for _, rank := range []uint8{2, 7} {
		for file := uint8(1); file <= 8; file++ {
			square := Square{file: file, rank: rank}
			pawn := Pawn{
				initSquare: square,
				square:     square,
				promotion:  noPromotion,
				moves:      0,
				captured:   false,
				captures:   0,
			}
			g.pawns[square] = pawn
		}
	}
}

// play tracks pawn moves and changes its properties.
func (g *Game) play() {

}

func main() {
	filepath := os.Args[1]
	f, err := os.Open(filepath)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	parser := newPgnParser(f)
	for parser.hasNextGame() {
		game, err := parser.nextGame()
		if err != nil {
			panic(err)
		}
		if game != nil {
			game.setUp()
			game.play()
			analyseGame(game)
		}
	}

	fmt.Println(stats)
}
