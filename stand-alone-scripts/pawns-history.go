/*
Objectives:
* Parse pgn files given in command line arguments. +
* Determine which moves are pawn moves. +
* Determine which moves capture a pawn.
* Statistics:
  - average fraction of pawn moves of all moves; +
  - correlation between fraction of pawn moves and number of moves in a game; +
  - average chances to survive for all pawns and for each pawn individually;
  - average chance to promote for all pawns and for each pawn individually;
  - chances to survive if pawn moves first;
  - chances to survive if pawn moves last or doesn't move;
  - balance of kills and deaths for all pawns and for each pawn individually;
  - average number of moves for each pawn;
  - how many moves for one death for each pawn;
  - correlation between captures and survival rate;

For iccf games check that final position is right comparing with position on www.iccf.com.
*/
package main

import (
	"bufio"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

const (
	white = iota
	black
)

const (
	noPromotion = iota
	knight
	bishop
	rook
	queen
)

var (
	iccfRegex          = regexp.MustCompile(`game(\d+)\.pgn$`)
	commentsRegex      = regexp.MustCompile(`{.*?}`)
	variationsRegex    = regexp.MustCompile(`\(.*?\)`)
	tagsRegex          = regexp.MustCompile(`\[.*?\]`)
	moveRegex          = regexp.MustCompile(`(\d+\.+)?(.*)`)
	moveNumRegex       = regexp.MustCompile(`\d+\.`)
	dotWithSpacesRegex = regexp.MustCompile(`\.\s+`)
	isPawnPlyRegex     = regexp.MustCompile(`^[a-h]`)
	isCapturePlyRegex  = regexp.MustCompile(`x`)
	pawnPlyRegex       = regexp.MustCompile(`^([a-h]x)?([a-h][1-8])(=[NBRQ])?(?:[+#])?$`)
	capturePlyRegex    = regexp.MustCompile(`^[a-hNBRQK](?:[a-h1-8])?x([a-h][1-8])(?:=[NBRQ])?(?:[+#])?$`)
	squareRegex        = regexp.MustCompile(`[a-h][1-8]`)
	byteFileOrRankMap  = map[byte]uint8{
		'a': 1,
		'b': 2,
		'c': 3,
		'd': 4,
		'e': 5,
		'f': 6,
		'g': 7,
		'h': 8,
		'1': 1,
		'2': 2,
		'3': 3,
		'4': 4,
		'5': 5,
		'6': 6,
		'7': 7,
		'8': 8,
	}
	bytePieceMap = map[byte]uint8{
		'N': knight,
		'B': bishop,
		'R': rook,
		'Q': queen,
	}
	fileStringMap = map[uint8]string{
		1: "a",
		2: "b",
		3: "c",
		4: "d",
		5: "e",
		6: "f",
		7: "g",
		8: "h",
	}
	colorStringMap = map[uint8]string{
		0: "white",
		1: "black",
	}

	pawns [2][8]*Pawn
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
	color          uint8
	initSquare     Square
	promotionCount int
	moveCount      int
	captureCount   int // pawn capture other pieces
	capturedCount  int // pawn captured by other pieces
	// in current game
	square    Square // empty if the pawn has been captured in current game
	promotion uint8
}

type Move [2]string

type Game struct {
	moves []Move
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

// findPawnByPly finds and returns the pawn that made move ply or panics.
func findPawnByPly(ply string, pawns [8]*Pawn) (foundPawn *Pawn) {
	color := pawns[0].color
	plyParts := pawnPlyRegex.FindStringSubmatch(ply)

	squareStr := plyParts[2]
	if squareStr == "" {
		panic(fmt.Sprintf("Could not fetch square from ply %s", ply))
	}
	file := byteFileOrRankMap[squareStr[0]]
	rank := byteFileOrRankMap[squareStr[1]]
	if file == 0 || rank == 0 {
		panic(fmt.Sprintf("Could not make square from ply %s", ply))
	}

	captureStr := plyParts[1]
	// There is no capture, pawn must be on the square file.
	if captureStr == "" {
		candidates := make([]*Pawn, 0)
		for _, pawn := range pawns {
			if pawn.promotion == noPromotion && pawn.square != (Square{}) && pawn.square.file == file {
				candidates = append(candidates, pawn)
			}
		}
		if len(candidates) == 0 {
			panic(fmt.Sprintf("Could not find pawn for move %s, no candidates", ply))
		}
		for _, pawn := range candidates {
			if (color == white && pawn.square.rank == rank-1) || (color == black && pawn.square.rank == rank+1) {
				foundPawn = pawn
				break
			}
		}
		// First move may be two squares forward.
		if foundPawn == nil {
			if color == white && rank == 4 {
				for _, pawn := range candidates {
					if pawn.square.rank == 2 {
						foundPawn = pawn
						break
					}
				}
			} else if color == black && rank == 5 {
				for _, pawn := range candidates {
					if pawn.square.rank == 7 {
						foundPawn = pawn
						break
					}
				}
			}
		}
		if foundPawn == nil {
			panic(fmt.Sprintf("Could not find pawn for move %s", ply))
		}
	} else {
		if color == white {
			rank = rank - 1
		} else if color == black {
			rank = rank + 1
		}
		square := Square{
			file: byteFileOrRankMap[captureStr[0]],
			rank: rank,
		}
		for _, pawn := range pawns {
			if pawn.promotion == noPromotion && pawn.square != (Square{}) && pawn.square == square {
				foundPawn = pawn
				break
			}
		}
		if foundPawn == nil {
			panic(fmt.Sprintf("Could not find pawn for move %s", ply))
		}
	}

	return
}

// validateFinalPosition compares pawns final position with final position of the game on url.
// If they are not the same, returns error.
// It is suitable only for ICCF games.
func validateFinalPosition(url string) error {
	var survivedPawns [2][]*Pawn
	for color := range pawns {
		for _, pawn := range pawns[color] {
			if pawn.promotion == noPromotion && pawn.square != (Square{}) {
				survivedPawns[color] = append(survivedPawns[color], pawn)
			}
		}
	}

	response, err := http.Get(url)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode != 200 {
		return fmt.Errorf("Status code error: %d %s", response.StatusCode, response.Status)
	}

	doc, err := goquery.NewDocumentFromReader(response.Body)
	if err != nil {
		return err
	}

	var squaresWithPawn [2]*goquery.Selection
	squaresWithPawn[white] = doc.Find(".cb-wpw")
	squaresWithPawn[white] = squaresWithPawn[white].Add(".cb-wpb")
	squaresWithPawn[black] = doc.Find(".cb-bpw")
	squaresWithPawn[black] = squaresWithPawn[black].Add(".cb-bpb")

	for color := range survivedPawns {
		if len(survivedPawns[color]) != squaresWithPawn[color].Length() {
			return fmt.Errorf("Number of surviving %s pawns does not match: %d != %d", colorStringMap[uint8(color)], len(survivedPawns[color]), squaresWithPawn[color].Length())
		}
	}

	return nil
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

func (s Square) String() string {
	return fmt.Sprintf("%s%d", fileStringMap[s.file], s.rank)
}

func (p Pawn) String() string {
	return fmt.Sprintf("%s pawn{initSquare: %s, square: %s}", colorStringMap[p.color], p.initSquare, p.square)
}

// setUp sets up the pawn on the game start.
func (p *Pawn) setUp() {
	p.square = p.initSquare
	p.promotion = noPromotion
}

// move applies ply to the pawn.
func (p *Pawn) move(ply string) {
	plyParts := pawnPlyRegex.FindStringSubmatch(ply)
	p.moveCount++

	if plyParts[1] != "" {
		p.captureCount++
	}

	promotionStr := plyParts[3]
	if promotionStr != "" {
		p.promotionCount++
		p.promotion = bytePieceMap[promotionStr[1]]
	}

	squareStr := plyParts[2]
	if squareStr == "" {
		panic(fmt.Sprintf("Could not fetch square from ply %s", ply))
	}
	file := byteFileOrRankMap[squareStr[0]]
	rank := byteFileOrRankMap[squareStr[1]]
	if file == 0 || rank == 0 {
		panic(fmt.Sprintf("Could not make square from ply %s", ply))
	}
	square := Square{file: file, rank: rank}
	p.square = square
}

// setUp sets up the pawns.
func (g *Game) setUp() {
	for color := range pawns {
		for _, pawn := range pawns[color] {
			pawn.setUp()
		}
	}
}

// play tracks game moves and changes pawns' properties.
func (g *Game) play() {
	var promoted [2]bool // At least one pawn has been promoted
	// var whitePromotions, blackPromotions map[int]bool // What pieces pawns have been promoted

	for _, move := range g.moves {
		for color, ply := range move {
			if !promoted[color] {
				if isPawnPlyRegex.MatchString(ply) {
					pawn := findPawnByPly(ply, pawns[color])
					pawn.move(ply)
					if pawn.promotion != noPromotion {
						promoted[color] = true
					}
				}
				if isCapturePlyRegex.MatchString(ply) {
					plyParts := capturePlyRegex.FindStringSubmatch(ply)
					if len(plyParts) < 2 {
						fmt.Println(ply)
					}
					squareStr := plyParts[1]
					if squareStr == "" {
						panic(fmt.Sprintf("Could not fetch square from ply %s", ply))
					}
					fmt.Printf("%s: %s %s\n", colorStringMap[uint8(color)], ply, squareStr)
				}
			} else {
				fmt.Printf("%s: Promoted!!!\n", colorStringMap[uint8(color)])
				// Consider moves by promotion pieces
			}
		}
	}
}

func (g *Game) analyse() {
	stats.games++
	var gamePlies, gamePawnPlies int

	for _, move := range g.moves {
		for _, ply := range move {
			if ply != "" { // That might be in the last move of a game
				stats.allPlies++
				gamePlies++

				if isPawnPlyRegex.MatchString(ply) {
					stats.pawnPlies++
					gamePawnPlies++
				}
			}
		}
	}
	stats.gamePliesList = append(stats.gamePliesList, gamePlies)
	stats.gamePawnPliesList = append(stats.gamePawnPliesList, gamePawnPlies)
}

func init() {
	for color := range pawns {
		var rank uint8
		if color == white {
			rank = 2
		} else {
			rank = 7
		}
		for i := range pawns[color] {
			square := Square{file: uint8(i) + 1, rank: rank}
			pawns[color][i] = &Pawn{
				color:      uint8(color),
				initSquare: square,
				square:     square,
			}
		}
	}
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
			// if iccfRegex.MatchString(filepath) {
			// 	gameId := iccfRegex.FindStringSubmatch(filepath)[1]
			// 	err = validateFinalPosition("https://www.iccf.com/game?id=" + gameId)
			// 	if err != nil {
			// 		panic(err)
			// 	}
			// }
			game.analyse()
		}
	}

	fmt.Println(stats)
}
