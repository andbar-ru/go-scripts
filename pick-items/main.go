package main

import (
	"flag"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"time"
)

const (
	aDesc = "abstention urn model, is default"
	cDesc = "confirmation urn model"
	oDesc = "opposition urn model"
	nDesc = "number of items to pick"
	hDesc = "print help"
)

const (
	Undefined = iota
	Abstention
	Confirmation
	Opposition
)

func printHelp(exitCode int) {
	fmt.Println("Usage: pick-items [flags] [item list to pick among]...")
	fmt.Println("Pick randomly n items out of the item list.")
	fmt.Println("You must specify the number of items to pick and item list.")
	fmt.Println("You can specify the urn model to select elements. Possible choices: abstention, confirmation, opposition. Default is abstention. All models are replacement models.")
	fmt.Println("Flags:")
	fmt.Printf("  -a  %s\n", aDesc)
	fmt.Printf("  -c  %s\n", cDesc)
	fmt.Printf("  -o  %s\n", oDesc)
	fmt.Printf("  -n <number>  %s\n", nDesc)
	fmt.Printf("  -h  %s\n", hDesc)
	os.Exit(exitCode)
}

func b2i(b *bool) int {
	if *b {
		return 1
	}
	return 0
}

func main() {
	if len(os.Args) < 2 {
		printHelp(0)
	}

	abstention := flag.Bool("a", false, aDesc)
	confirmation := flag.Bool("c", false, cDesc)
	opposition := flag.Bool("o", false, oDesc)
	n := flag.Int("n", 0, nDesc)
	help := flag.Bool("h", false, hDesc)

	flag.Parse()

	if *help {
		printHelp(0)
	}
	if *n < 2 {
		fmt.Fprintf(os.Stderr, "You must specify the number of items to pick. This number must be greater than 0.\n\n")
		printHelp(1)
	}
	urnModelsSum := b2i(abstention) + b2i(confirmation) + b2i(opposition)
	if urnModelsSum > 1 {
		fmt.Fprintf(os.Stderr, "You can specify only one url model.\n\n")
		printHelp(1)
	}
	var urnModel int
	switch {
	case *abstention:
		urnModel = Abstention
	case *confirmation:
		urnModel = Confirmation
	case *opposition:
		urnModel = Opposition
	default:
		urnModel = Abstention
	}
	origItems := flag.Args()
	if len(origItems) < 2 {
		fmt.Fprintf(os.Stderr, "You must specify the item list. Its size must be greater than 1.\n\n")
		printHelp(1)
	}

	items := make([]string, len(origItems))
	copy(items, origItems)

	today, err := strconv.Atoi(time.Now().Format("20060102"))
	if err != nil {
		panic(err)
	}
	r := rand.New(rand.NewSource(int64(today)))

	pickedItems := make(map[string]int)

	for _, item := range items {
		pickedItems[item] = 0
	}
	for i := 0; i < *n; i++ {
		item := items[r.Intn(len(items))]
		pickedItems[item]++
		switch urnModel {
		case Abstention:
		case Confirmation:
			items = append(items, item)
		case Opposition:
			for _, it := range origItems {
				if it != item {
					items = append(items, it)
				}
			}
		}
	}

	for item, n := range pickedItems {
		fmt.Printf("%s: %d\n", item, n)
	}
}
