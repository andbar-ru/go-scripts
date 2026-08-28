package main

import "fmt"

const N = 8 // board size

type board [N]int

var count = 0 // total count of possible layouts

// Checks the position (n, c) is out of attack.
func isPlaceOK(a board, n, c int) bool {
	for i := 0; i < n-1; i++ { // for each already placed queen
		if a[i] == c || a[i]-i-1 == c-n || a[i]+i+1 == c+n { // the same column or diagonal?
			return false // position is under attack
		}
	}
	return true // position is out of attack
}

// Prints the board.
func printSolution(a board) {
	count++

	for i := range a {
		for j := 1; j <= N; j++ {
			if a[i] == j {
				fmt.Print("X ")
			} else {
				fmt.Print("- ")
			}
		}
		fmt.Println()
	}
	fmt.Println()
}

// Add to the board all queens from 'n' to 'N'
func addQueen(a board, n int) {
	if n > N { // all queens have been placed
		printSolution(a)
	} else { // try to place n-th queen
		for c := 1; c <= N; c++ {
			if isPlaceOK(a, n, c) {
				a[n-1] = c // place n-th queen in column 'c'
				addQueen(a, n+1)
			}
		}
	}
}

func main() {
	var a board
	addQueen(a, 1)

	fmt.Println(count)
}
