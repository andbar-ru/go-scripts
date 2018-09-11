package main

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

const (
	N            = 1000000
	maxBarLength = 80
)

func main() {
	results := make(map[int]int)
	rand.Seed(time.Now().UnixNano())
	for i := 0; i < N; i++ {
		n := 1
		for {
			if r := rand.Intn(10); r == 0 {
				results[n]++
				break
			}
			n++
		}
	}
	printResults(results)
}

func printResults(results map[int]int) {
	// Figure out max key and max value
	var maxKey, maxValue int
	for key, value := range results {
		if key > maxKey {
			maxKey = key
		}
		if value > maxValue {
			maxValue = value
		}
	}

	maxKeyDigits := len(strconv.Itoa(maxKey))
	k := maxValue / maxBarLength

	for n := 1; n <= maxKey; n++ {
		fmt.Printf("%*d|", maxKeyDigits, n)
		bar := strings.Repeat("=", results[n]/k)
		fmt.Println(bar)
	}
}
