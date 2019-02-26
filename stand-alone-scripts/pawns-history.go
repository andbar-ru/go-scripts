/*
Objectives:
* Parse pgn files given in command line arguments. +
* Determine which moves are pawn moves. +
* Determine which moves capture a pawn. +
* Statistics:
  - average fraction of pawn moves of all moves; +
  - correlation between fraction of pawn moves and number of moves in a game; +
  - average chances of survival for all pawns, for pawns of each color and for each pawn individually in a game; +
  - average chances to promote for all pawns, for pawns of each color and for each pawn individually in a game; +
  - chances of survival if pawn moves;
  - chances of survival if pawn doesn't move;
  - chances of survival if pawn moves first; +
  - chances of survival if pawn moves last;
  - balance of kills and deaths for all pawns, for pawns of each color and for each pawn individually;
  - average number of moves for each pawn in a game;
  - how many moves for one death for each pawn;
  - correlation between captures and survival rate;

For iccf games validate that final position comparing it with position on www.iccf.com. +
*/
package main

import (
	"bufio"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"path"
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
	pawn = iota
	knight
	bishop
	rook
	queen
	king
)

var (
	iccfRegex          = regexp.MustCompile(`game(\d+)\.pgn$`)
	commentsRegex      = regexp.MustCompile(`{.*?}`)
	variationsRegex    = regexp.MustCompile(`\(.*?\)`)
	tagsRegex          = regexp.MustCompile(`\[.*?\]`)
	moveRegex          = regexp.MustCompile(`(\d+\.+)?(.*)`)
	moveNumRegex       = regexp.MustCompile(`\d+\.`)
	dotWithSpacesRegex = regexp.MustCompile(`\.\s+`)
	plyRegex           = regexp.MustCompile(`^([NBRQK])?([a-h1-8])?(x)?([a-h][1-8]|O-O(?:-O)?)(=[NBRQ])?(?:[+#])?$`)
	isPawnPlyRegex     = regexp.MustCompile(`^[a-h]`)
	isEnPassantRegex   = regexp.MustCompile(`^[a-h]x[a-h][36](?:[+#])?$`)
	isCaptureRegex     = regexp.MustCompile(`x`)
	squareRegex        = regexp.MustCompile(`^[a-h][1-8]$`)
	squareClassRegex   = regexp.MustCompile(`^cb-([bw][pnbrqk])?[bw]$`)

	byteFileOrRankMap = map[byte]uint8{
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
		white: "white",
		black: "black",
	}
	pieceSymbolMap = map[uint8]string{
		pawn:   "p",
		knight: "n",
		bishop: "b",
		rook:   "r",
		queen:  "q",
		king:   "k",
	}
	pieceStringMap = map[uint8]string{
		pawn:   "pawn",
		knight: "knight",
		bishop: "bishop",
		rook:   "rook",
		queen:  "queen",
		king:   "king",
	}

	board     = &Board{}
	allPieces [2][16]*Piece // pieces in order: ppppppppnnbbrrqk
	stats     = &Stats{}
)

type Stats struct {
	games                                    int
	allPlies                                 int
	pawnPlies                                int
	gamePliesList                            []int
	gamePawnPliesList                        []int
	captureCount                             int
	pawnMovedFirstAndSurvivedCount           [2]int
	pawnMovedLastOrDidntMoveAndSurvivedCount [2]int
}

type PgnParser struct {
	reader          *bufio.Reader
	prevLineIsEmpty bool
	err             error
}

type Square struct {
	file  uint8
	rank  uint8
	piece *Piece
}

type Board struct {
	squares          [8][8]*Square
	pawnsMovedInGame [2]map[*Piece]bool
}

type Piece struct {
	color          uint8
	initType       uint8 // Pawns might have different initType and type
	curType        uint8 // due to promotion.
	initSquare     *Square
	curSquare      *Square // nil if piece has been captured
	moveCount      int
	captureCount   int
	capturedCount  int // always 0 for kings
	promotionCount int // only for pawns
	// game stats
	movedFirst           bool
	movedLastOrDidntMove bool
}

type Move [2]string

type Game struct {
	moves []Move
}

func check(err error) {
	if err != nil {
		fmt.Println(board)
		panic(err)
	}
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
	check(err)
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

func pickPiece(pieces []*Piece, plyParts []string) (*Piece, error) {
	if len(pieces) == 0 {
		return nil, fmt.Errorf("There are no pieces to pick from for ply %s", plyParts[0])
	}
	if len(pieces) == 1 {
		if plyParts[2] != "" {
			return nil, fmt.Errorf("Must be more than one piece to pick from for ply %s", plyParts[0])
		}
		return pieces[0], nil
	}
	// len(pieces) > 1
	if plyParts[2] == "" {
		// One piece can move, others do not, because of check to own king.
		piecesAble := make([]*Piece, 0, 1)
		destSquare, err := board.getSquareString(plyParts[4])
		check(err)
		for _, piece := range pieces {
			srcSquare := piece.curSquare
			board.movePieceOnSquare(piece, destSquare, true) // test move
			result := board.kingInCheck(piece.color)
			if !result {
				piecesAble = append(piecesAble, piece)
			}
			board.movePieceOnSquare(piece, srcSquare, true) // roll back move
		}
		if len(piecesAble) == 0 {
			return nil, fmt.Errorf("There are no pieces to pick from. Ply %s", plyParts[0])
		} else if len(piecesAble) > 1 {
			return nil, fmt.Errorf("There are multiple pieces to pick from but there are no extra info about source square. Ply %s", plyParts[0])
		}
		return piecesAble[0], nil
	}
	source := plyParts[2][0]
	var foundPiece *Piece
	if source >= 97 && source <= 104 { // a-h
		file := byteFileOrRankMap[source]
		for _, piece := range pieces {
			if piece.curSquare.file == file {
				if foundPiece != nil {
					return nil, fmt.Errorf("Multiple pieces are suitable for move %s", plyParts[0])
				}
				foundPiece = piece
			}
		}
	} else if source >= 49 && source <= 56 { // 1-8
		rank := byteFileOrRankMap[source]
		for _, piece := range pieces {
			if piece.curSquare.rank == rank {
				if foundPiece != nil {
					return nil, fmt.Errorf("Multiple pieces are suitable for move %s", plyParts[0])
				}
				foundPiece = piece
			}
		}
	}
	if foundPiece == nil {
		return nil, fmt.Errorf("Could not find piece for move %s", plyParts[0])
	}
	return foundPiece, nil
}

// validateFinalPosition compares final position with final position of this game on iccf.com.
// If they are not the same, returns error.
// It is suitable only for ICCF games.
func validateFinalPosition(url string) error {
	response, err := http.Get(url)
	check(err)
	defer response.Body.Close()

	if response.StatusCode != 200 {
		return fmt.Errorf("Status code error: %d %s", response.StatusCode, response.Status)
	}

	doc, err := goquery.NewDocumentFromReader(response.Body)
	check(err)

	chessboard := doc.Find(".chessboard").First()
	squares := chessboard.ChildrenFiltered("[name]")
	if squares.Length() != 64 {
		return fmt.Errorf("Number of squares is not 64 (%d)", squares.Length())
	}
	var errSquares []string

	squares.Each(func(i int, span *goquery.Selection) {
		name, _ := span.Attr("name")
		if !squareRegex.MatchString(name) {
			panic(fmt.Sprintf("Unexpected square name %q", name))
		}
		class, _ := span.Attr("class")
		if !squareClassRegex.MatchString(class) {
			panic(fmt.Sprintf("Unexpected square class %q", class))
		}
		file := byteFileOrRankMap[name[0]]
		rank := byteFileOrRankMap[name[1]]
		square, err := board.getSquare(file, rank)
		check(err)

		// Assemble the expected class
		var squareColor, squarePiece, squarePieceColor string
		if file&1 == rank&1 {
			squareColor = "b"
		} else {
			squareColor = "w"
		}
		if piece := square.piece; piece != nil {
			squarePiece = pieceSymbolMap[piece.curType]
			squarePieceColor = string(colorStringMap[piece.color][0])
		}
		expectedClass := fmt.Sprintf("cb-%s%s%s", squarePieceColor, squarePiece, squareColor)

		if class != expectedClass {
			errSquares = append(errSquares, name)
		}
	})

	if len(errSquares) > 0 {
		return fmt.Errorf("Mismatched squares: %s", strings.Join(errSquares, ","))
	}

	return nil
}

func validateStats() {
	// Total captureCount is equal to sum of capturedCount and sum of captureCount of all pieces.
	var capturedSum, captureSum int
	for color := range allPieces {
		for _, piece := range allPieces[color] {
			capturedSum += piece.capturedCount
			captureSum += piece.captureCount
		}
	}
	if stats.captureCount != capturedSum {
		panic(fmt.Sprintf("Total captureCount is not equal to sum of capturedCount of all pieces: %d != %d", stats.captureCount, capturedSum))
	}
	if stats.captureCount != captureSum {
		panic(fmt.Sprintf("Total captureCount is not equal to sum of captureCount of all pieces: %d != %d", stats.captureCount, captureSum))
	}

	// Pawn moved first
	if stats.pawnMovedFirstAndSurvivedCount[0] != stats.games*2 {
		panic(fmt.Sprintf("Pawn moved first count is not double of games: %d != %d * 2", stats.pawnMovedFirstAndSurvivedCount[0], stats.games))
	}
	// Pawn moved last or didn't move
	if stats.pawnMovedLastOrDidntMoveAndSurvivedCount[0] < stats.games*2 {
		panic(fmt.Sprintf("Pawn moved last or didn't move count must not be less than double of games: %d < %d * 2", stats.pawnMovedLastOrDidntMoveAndSurvivedCount[0], stats.games))
	}
}

func (s *Stats) String() string {
	var output string

	output += fmt.Sprintf("Games: %d\n", s.games)
	output += fmt.Sprintf("All plies: %d\n", s.allPlies)
	output += fmt.Sprintf("Pawn plies: %d (%.1f %%)\n", s.pawnPlies, getPercent(float64(s.pawnPlies), float64(s.allPlies)))
	output += fmt.Sprintf("Correlation between fraction of pawn moves and number of moves in a game: %.4f\n", s.getCorrelationBetweenPawnFractionAndGamePlies())

	pawnSurvivalTotal, pawnSurvivalColor, pawnSurvivals := s.getPawnSurvivalStats()
	output += fmt.Sprintf("Total pawn chances of survival: %.2f %%, white: %.2f %%, black: %.2f %%\n", pawnSurvivalTotal*100.0, pawnSurvivalColor[0]*100.0, pawnSurvivalColor[1]*100.0)
	output += "Individually:\n"
	for color := range allPieces {
		for _, piece := range allPieces[color][:8] {
			output += fmt.Sprintf("  %s: %.2f %%\n", piece.initSquare, pawnSurvivals[piece]*100.0)
		}
	}

	pawnPromotionTotal, pawnPromotionColor, pawnPromotions := s.getPawnPromotionStats()
	output += fmt.Sprintf("Total pawn chances of promotion: %.2f %%, white: %.2f %%, black: %.2f %%\n", pawnPromotionTotal*100.0, pawnPromotionColor[0]*100.0, pawnPromotionColor[1]*100.0)
	output += "Individually:\n"
	for color := range allPieces {
		for _, piece := range allPieces[color][:8] {
			output += fmt.Sprintf("  %s: %.2f %%\n", piece.initSquare, pawnPromotions[piece]*100.0)
		}
	}

	output += fmt.Sprintf("Chances of survival if pawn moves first: %d of %d = %.2f %%\n", s.pawnMovedFirstAndSurvivedCount[1], s.pawnMovedFirstAndSurvivedCount[0], float64(s.pawnMovedFirstAndSurvivedCount[1])/float64(s.pawnMovedFirstAndSurvivedCount[0])*100.0)
	output += fmt.Sprintf("Chances of survival if pawn moves last or doesn't move: %d of %d = %.2f %%\n", s.pawnMovedLastOrDidntMoveAndSurvivedCount[1], s.pawnMovedLastOrDidntMoveAndSurvivedCount[0], float64(s.pawnMovedLastOrDidntMoveAndSurvivedCount[1])/float64(s.pawnMovedLastOrDidntMoveAndSurvivedCount[0])*100.0)

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
	check(err)
	return correlation
}

func (s *Stats) getPawnSurvivalStats() (pawnSurvivalTotal float64, pawnSurvivalColor [2]float64, pawnSurvivals map[*Piece]float64) {
	var pawnSurvivalTotalSum int
	var pawnSurvivalColorSums [2]int
	pawnSurvivals = make(map[*Piece]float64)

	for color := range allPieces {
		for _, piece := range allPieces[color][:8] {
			pawnSurvivalTotalSum += s.games - piece.capturedCount
			pawnSurvivalColorSums[color] += s.games - piece.capturedCount
			pawnSurvivals[piece] = float64(s.games-piece.capturedCount) / float64(s.games)
		}
	}
	pawnSurvivalTotal = float64(pawnSurvivalTotalSum) / float64(s.games*16)
	for color, sum := range pawnSurvivalColorSums {
		pawnSurvivalColor[color] = float64(sum) / float64(s.games*8)
	}
	return
}

func (s *Stats) getPawnPromotionStats() (pawnPromotionTotal float64, pawnPromotionColor [2]float64, pawnPromotions map[*Piece]float64) {
	var pawnPromotionTotalSum int
	var pawnPromotionColorSums [2]int
	pawnPromotions = make(map[*Piece]float64)

	for color := range allPieces {
		for _, piece := range allPieces[color][:8] {
			pawnPromotionTotalSum += piece.promotionCount
			pawnPromotionColorSums[color] += piece.promotionCount
			pawnPromotions[piece] = float64(piece.promotionCount) / float64(s.games)
		}
	}
	pawnPromotionTotal = float64(pawnPromotionTotalSum) / float64(s.games*16)
	for color, sum := range pawnPromotionColorSums {
		pawnPromotionColor[color] = float64(sum) / float64(s.games*8)
	}
	return
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

func (b Board) String() string {
	var output string
	output += "┌───┬───┬───┬───┬───┬───┬───┬───┐\n"

	for rank := 8; rank >= 1; rank-- {
		for file := range b.squares {
			square, err := b.getSquare(uint8(file+1), uint8(rank))
			check(err)
			output += "│ "
			if square.piece != nil {
				output += square.piece.symbol()
			} else {
				output += " "
			}
			output += " "
		}
		output += "│\n"
		if rank != 1 {
			output += "├───┼───┼───┼───┼───┼───┼───┼───┤\n"
		} else {
			output += "└───┴───┴───┴───┴───┴───┴───┴───┘\n"
		}
	}

	return output
}

func (b *Board) getSquare(file, rank uint8) (*Square, error) {
	if file == 0 || file > 8 {
		return nil, fmt.Errorf("Invalid file %d", file)
	}
	if rank == 0 || rank > 8 {
		return nil, fmt.Errorf("Invalid rank %d", rank)
	}
	return b.squares[file-1][rank-1], nil
}

func (b *Board) getSquareString(str string) (*Square, error) {
	file := byteFileOrRankMap[str[0]]
	rank := byteFileOrRankMap[str[1]]
	return b.getSquare(file, rank)
}

func (b *Board) movePieceOnSquare(piece *Piece, square *Square, test bool) {
	if piece.curSquare != nil {
		piece.curSquare.piece = nil
	}
	if square.piece != nil {
		panic(fmt.Sprintf("Square %s is occupied by %s", square, square.piece))
	}
	square.piece = piece
	piece.curSquare = square
	if !test {
		piece.moveCount++
	}
}

// capture removes captured piece from the board.
func (b *Board) capture(plyParts []string, color uint8) {
	square, err := b.getSquareString(plyParts[4])
	check(err)

	// Search piece of other color on the square.
	piece := square.piece
	if piece == nil { // en passant?
		if !isEnPassantRegex.MatchString(plyParts[0]) {
			panic(fmt.Sprintf("Capture of unknown piece: %s", plyParts[0]))
		}
		if color == white && square.rank == 6 {
			square, _ = b.getSquare(square.file, 5)
			piece = square.piece
			if piece == nil || piece.curType != pawn || piece.color == color {
				panic(fmt.Sprintf("En passant of unknown pawn: %s", plyParts[0]))
			}
		}
		if color == black && square.rank == 3 {
			square, _ = b.getSquare(square.file, 4)
			piece = square.piece
			if piece == nil || piece.curType != pawn || piece.color == color {
				panic(fmt.Sprintf("En passant of unknown pawn: %s", plyParts[0]))
			}
		}
	}
	if piece == nil || piece.color == color {
		panic(fmt.Sprintf("Capture of unknown piece: %s", plyParts[0]))
	}
	// Remove the piece.
	square.piece = nil
	piece.curSquare = nil
	piece.capturedCount++
}

func (b *Board) movePawn(plyParts []string, color uint8) {
	square, err := b.getSquareString(plyParts[4])
	check(err)

	// Search piece
	var piece *Piece
	var step int
	if color == white {
		step = -1
	} else {
		step = 1
	}
	if plyParts[3] == "" {
		srcSquare, err := b.getSquare(square.file, uint8(int(square.rank)+step))
		check(err)
		piece = srcSquare.piece
		if piece == nil && ((color == white && square.rank == 4) || (color == black && square.rank == 5)) {
			srcSquare, err = b.getSquare(square.file, uint8(int(square.rank)+2*step))
			check(err)
			piece = srcSquare.piece
		}
	} else {
		srcFile := byteFileOrRankMap[plyParts[2][0]]
		srcSquare, err := b.getSquare(srcFile, uint8(int(square.rank)+step))
		check(err)
		piece = srcSquare.piece
	}
	if piece == nil || piece.curType != pawn || piece.color != color {
		panic("Unknown pawn move")
	}

	// special case - pawn promotion
	if (color == white && square.rank == 8) || (color == black && square.rank == 1) {
		promotion := plyParts[5]
		if promotion == "" {
			panic("Pawn has reached back rank without promotion")
		}
		pieceType := bytePieceMap[promotion[1]]
		piece.curType = pieceType
		piece.promotionCount++
	}

	if plyParts[3] == "x" {
		piece.captureCount++
	}
	b.movePieceOnSquare(piece, square, false)

	// More pawn specific stats
	if len(b.pawnsMovedInGame[color]) == 0 {
		piece.movedFirst = true
	}
	if len(b.pawnsMovedInGame[color]) < 8 {
		piece.movedLastOrDidntMove = false
	}
	b.pawnsMovedInGame[color][piece] = true
}

func (b *Board) moveKnight(plyParts []string, color uint8) {
	square, err := b.getSquareString(plyParts[4])
	check(err)

	// Collect candidate pieces
	pieces := make([]*Piece, 0, 2)
	steps := [2]int{1, 2}
	factors := [2]int{-1, 1}
	for i := range steps {
		for _, f1 := range factors {
			for _, f2 := range factors {
				file := uint8(int(square.file) + steps[i]*f1)
				rank := uint8(int(square.rank) + steps[i^1]*f2)
				srcSquare, err := b.getSquare(file, rank)
				if err == nil {
					piece := srcSquare.piece
					if piece != nil && piece.color == color && piece.curType == knight {
						pieces = append(pieces, piece)
					}
				}
			}
		}
	}

	piece, err := pickPiece(pieces, plyParts)
	check(err)

	if plyParts[3] == "x" {
		piece.captureCount++
	}
	b.movePieceOnSquare(piece, square, false)
}

func (b *Board) moveBishop(plyParts []string, color uint8) {
	square, err := b.getSquareString(plyParts[4])
	check(err)

	// Collect candidate pieces
	pieces := make([]*Piece, 0, 1)
	factors := [2]int{-1, 1}
	for _, f1 := range factors {
		for _, f2 := range factors {
			step := 1
			for {
				file := uint8(int(square.file) + step*f1)
				rank := uint8(int(square.rank) + step*f2)
				srcSquare, err := b.getSquare(file, rank)
				if err != nil {
					break
				}
				piece := srcSquare.piece
				if piece != nil {
					if piece.color == color && piece.curType == bishop {
						pieces = append(pieces, piece)
					}
					break
				}
				step++
			}
		}
	}

	piece, err := pickPiece(pieces, plyParts)
	check(err)

	if plyParts[3] == "x" {
		piece.captureCount++
	}
	b.movePieceOnSquare(piece, square, false)
}

func (b *Board) moveRook(plyParts []string, color uint8) {
	square, err := b.getSquareString(plyParts[4])
	check(err)

	// Collect candidate pieces
	pieces := make([]*Piece, 0, 2)
	factors := [2]int{-1, 1}
	for _, f1 := range factors {
		// along the rank
		step := 1
		for {
			file := uint8(int(square.file) + step*f1)
			srcSquare, err := b.getSquare(file, square.rank)
			if err != nil {
				break
			}
			piece := srcSquare.piece
			if piece != nil {
				if piece.color == color && piece.curType == rook {
					pieces = append(pieces, piece)
				}
				break
			}
			step++
		}

		// along the file
		step = 1
		for {
			rank := uint8(int(square.rank) + step*f1)
			srcSquare, err := b.getSquare(square.file, rank)
			if err != nil {
				break
			}
			piece := srcSquare.piece
			if piece != nil {
				if piece.color == color && piece.curType == rook {
					pieces = append(pieces, piece)
				}
				break
			}
			step++
		}
	}

	piece, err := pickPiece(pieces, plyParts)
	check(err)

	if plyParts[3] == "x" {
		piece.captureCount++
	}
	b.movePieceOnSquare(piece, square, false)
}

func (b *Board) moveQueen(plyParts []string, color uint8) {
	square, err := b.getSquareString(plyParts[4])
	check(err)

	// Collect candidate pieces
	pieces := make([]*Piece, 0, 1)
	factors := [3]int{-1, 0, 1}
	for _, f1 := range factors {
		for _, f2 := range factors {
			if f1 == 0 && f2 == 0 {
				continue
			}
			step := 1
			for {
				file := uint8(int(square.file) + step*f1)
				rank := uint8(int(square.rank) + step*f2)
				srcSquare, err := b.getSquare(file, rank)
				if err != nil {
					break
				}
				piece := srcSquare.piece
				if piece != nil {
					if piece.color == color && piece.curType == queen {
						pieces = append(pieces, piece)
					}
					break
				}
				step++
			}
		}
	}

	piece, err := pickPiece(pieces, plyParts)
	check(err)

	if plyParts[3] == "x" {
		piece.captureCount++
	}
	b.movePieceOnSquare(piece, square, false)
}

func (b *Board) moveKing(plyParts []string, color uint8) {
	// Castling
	if plyParts[4] == "O-O" || plyParts[4] == "O-O-O" {
		var kingFile, rookFile, rank, kingDestFile, rookDestFile uint8
		kingFile = 5
		if color == white {
			rank = 1
		} else {
			rank = 8
		}
		if plyParts[4] == "O-O" {
			rookFile = 8
			kingDestFile = 7
			rookDestFile = 6
		} else {
			rookFile = 1
			kingDestFile = 3
			rookDestFile = 4
		}

		kingSquare, _ := b.getSquare(kingFile, rank)
		kingPiece := kingSquare.piece
		if kingPiece == nil || kingPiece.color != color || kingPiece.curType != king {
			panic(fmt.Sprintf("%s: could not find king", plyParts[0]))
		}
		rookSquare, _ := b.getSquare(rookFile, rank)
		rookPiece := rookSquare.piece
		if rookPiece == nil || rookPiece.color != color || rookPiece.curType != rook {
			panic(fmt.Sprintf("%s: could not find rook", plyParts[0]))
		}
		kingDestSquare, _ := b.getSquare(kingDestFile, rank)
		rookDestSquare, _ := b.getSquare(rookDestFile, rank)

		b.movePieceOnSquare(kingPiece, kingDestSquare, false)
		b.movePieceOnSquare(rookPiece, rookDestSquare, false)
	} else {
		square, err := b.getSquareString(plyParts[4])
		check(err)

		// Collect candidate pieces
		pieces := make([]*Piece, 0, 1)
		steps := [3]int{-1, 0, 1}
		for _, fileStep := range steps {
			for _, rankStep := range steps {
				if fileStep == 0 && rankStep == 0 {
					continue
				}
				file := uint8(int(square.file) + fileStep)
				rank := uint8(int(square.rank) + rankStep)
				srcSquare, err := b.getSquare(file, rank)
				if err != nil {
					continue
				}
				piece := srcSquare.piece
				if piece != nil && piece.color == color && piece.curType == king {
					pieces = append(pieces, piece)
				}
			}
		}

		piece, err := pickPiece(pieces, plyParts)
		check(err)

		if plyParts[3] == "x" {
			piece.captureCount++
		}
		b.movePieceOnSquare(piece, square, false)
	}
}

// move makes move and modifies board and pieces.
func (b *Board) move(ply string, color uint8) {
	plyParts := plyRegex.FindStringSubmatch(ply)
	if len(plyParts) == 0 {
		panic(fmt.Sprintf("Unexpected ply: %s", ply))
	}

	// First captures
	if plyParts[3] == "x" {
		b.capture(plyParts, color)
	}

	switch {
	case plyParts[1] == "" && ply[0] != 'O':
		b.movePawn(plyParts, color)
	case plyParts[1] == "N":
		b.moveKnight(plyParts, color)
	case plyParts[1] == "B":
		b.moveBishop(plyParts, color)
	case plyParts[1] == "R":
		b.moveRook(plyParts, color)
	case plyParts[1] == "Q":
		b.moveQueen(plyParts, color)
	case plyParts[1] == "K" || ply[0] == 'O':
		b.moveKing(plyParts, color)
	default:
	}
}

// kingInCheck determines if the king of given color is in check by opponents' bishop, rook or queen.
func (b *Board) kingInCheck(color uint8) bool {
	piece := allPieces[color][15]
	if piece.curType != king {
		panic(fmt.Sprintf("Piece is not a king, just %s", pieceStringMap[piece.curType]))
	}

	square := piece.curSquare
	factors := [3]int{-1, 0, 1}
	for _, f1 := range factors {
		for _, f2 := range factors {
			step := 1
			for {
				file := uint8(int(square.file) + step*f1)
				rank := uint8(int(square.rank) + step*f2)
				srcSquare, err := b.getSquare(file, rank)
				if err != nil {
					break
				}
				piece := srcSquare.piece
				if piece != nil {
					if piece.color == color^1 {
						if f1 != 0 && f2 != 0 { // diagonal
							if piece.curType == bishop || piece.curType == queen {
								return true
							}
						} else { // file or rank
							if piece.curType == rook || piece.curType == queen {
								return true
							}
						}
					}
					break
				}
				step++
			}
		}
	}
	return false
}

// setUp sets up all pieces before game starts.
func (b *Board) setUp() {
	// Clear the board
	for file := range b.squares {
		for _, square := range b.squares[file] {
			square.piece = nil
		}
	}
	for color := range b.pawnsMovedInGame {
		b.pawnsMovedInGame[color] = make(map[*Piece]bool)
	}

	for color := range allPieces {
		for _, piece := range allPieces[color] {
			// for pawns
			piece.curType = piece.initType
			piece.movedFirst = false
			piece.movedLastOrDidntMove = true

			piece.curSquare = piece.initSquare
			piece.curSquare.piece = piece
		}
	}
}

// symbol returns one symbol representing the piece.
func (p *Piece) symbol() string {
	symbol := pieceSymbolMap[p.curType]
	if p.color == white {
		symbol = strings.ToUpper(symbol)
	}
	return symbol
}

// String implements interface Stringer.
func (p *Piece) String() string {
	return fmt.Sprintf("%s %s{initSquare: %s, curSquare: %s}", colorStringMap[p.color], pieceStringMap[p.curType], p.initSquare, p.curSquare)
}

// play tracks game moves and gathers statistics.
func (g *Game) play() {
	stats.games++
	var gamePlies, gamePawnPlies int

	for _, move := range g.moves {
		for color, ply := range move {
			if ply != "" { // black ply in the last move
				stats.allPlies++
				gamePlies++

				board.move(ply, uint8(color))

				if isPawnPlyRegex.MatchString(ply) {
					stats.pawnPlies++
					gamePawnPlies++
				}
				if isCaptureRegex.MatchString(ply) {
					stats.captureCount++
				}
			}
		}
	}
	stats.gamePliesList = append(stats.gamePliesList, gamePlies)
	stats.gamePawnPliesList = append(stats.gamePawnPliesList, gamePawnPlies)

	var movedFirst, movedLastOrDidntMove bool
	for color := range allPieces {
		for _, piece := range allPieces[color][:8] {
			if piece.initType != pawn {
				panic(fmt.Sprintf("Piece is not a pawn, just %s", pieceStringMap[piece.initType]))
			}
			if piece.movedFirst {
				stats.pawnMovedFirstAndSurvivedCount[0]++
				if piece.curSquare != nil {
					stats.pawnMovedFirstAndSurvivedCount[1]++
				}
				movedFirst = true
			}
			if piece.movedLastOrDidntMove {
				stats.pawnMovedLastOrDidntMoveAndSurvivedCount[0]++
				if piece.curSquare != nil {
					stats.pawnMovedLastOrDidntMoveAndSurvivedCount[1]++
				}
				movedLastOrDidntMove = true
			}
		}
	}
	if movedFirst == false {
		panic(fmt.Sprintf("This game has not first pawn moves: %s", g.moves))
	}
	if movedLastOrDidntMove == false {
		panic(fmt.Sprintf("This game has not last pawn moves or pawns without moves: %s", g.moves))
	}
}

func init() {
	// Init board
	for file := 1; file <= 8; file++ {
		for rank := 1; rank <= 8; rank++ {
			board.squares[file-1][rank-1] = &Square{
				file: uint8(file),
				rank: uint8(rank),
			}
		}
	}
	// Init pieces
	var index, color int

	var initPiece = func(pieceType, file, rank uint8) {
		square, err := board.getSquare(file, rank)
		check(err)
		piece := &Piece{
			color:      uint8(color),
			initType:   pieceType,
			curType:    pieceType,
			initSquare: square,
		}
		allPieces[color][index] = piece
		index++
	}

	for color = range allPieces {
		var rank uint8
		index = 0

		// Pawns
		if color == white {
			rank = 2
		} else {
			rank = 7
		}
		for file := uint8(1); file <= 8; file++ {
			initPiece(pawn, file, rank)
		}

		// Pieces
		if color == white {
			rank = 1
		} else {
			rank = 8
		}
		for _, file := range [2]uint8{2, 7} {
			initPiece(knight, file, rank)
		}
		for _, file := range [2]uint8{3, 6} {
			initPiece(bishop, file, rank)
		}
		for _, file := range [2]uint8{1, 8} {
			initPiece(rook, file, rank)
		}
		initPiece(queen, 4, rank)
		initPiece(king, 5, rank)
	}
}

func main() {
	arg := os.Args[1]
	fileinfo, err := os.Stat(arg)
	check(err)
	mode := fileinfo.Mode()

	filepaths := make([]string, 0)

	if mode.IsRegular() && path.Ext(fileinfo.Name()) == ".pgn" {
		filepaths = append(filepaths, arg)
	} else if mode.IsDir() {
		dir, err := os.Open(arg)
		check(err)
		defer dir.Close()

		fileinfos, err := dir.Readdir(0)
		check(err)
		for _, fileinfo := range fileinfos {
			if fileinfo.Mode().IsRegular() && path.Ext(fileinfo.Name()) == ".pgn" {
				filepaths = append(filepaths, path.Join(arg, fileinfo.Name()))
			}
		}
	} else {
		fmt.Println("Argument is of unknown mode")
	}

	if len(filepaths) == 0 {
		fmt.Println("There are nothing to process")
		return
	}

	for _, filepath := range filepaths {
		f, err := os.Open(filepath)
		check(err)
		defer f.Close()

		parser := newPgnParser(f)
		for parser.hasNextGame() {
			game, err := parser.nextGame()
			check(err)
			if game != nil {
				board.setUp()
				game.play()
				if iccfRegex.MatchString(filepath) {
					gameId := iccfRegex.FindStringSubmatch(filepath)[1]
					err = validateFinalPosition("https://www.iccf.com/game?id=" + gameId)
					if err != nil {
						fmt.Println(board)
						panic(err)
					}
				}
			}
		}
	}

	validateStats()
	fmt.Println(stats)
}
