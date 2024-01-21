package main

import (
	"fmt"
	"math/rand"
	"strconv"
	"time"
)

const N = 15

var canned = []string{"Tuna and mussels", "Tuna and salmon", "Tuna and crab", "Tuna", "Tuna and shrimps"}

func main() {
	today, err := strconv.Atoi(time.Now().Format("20060102"))
	if err != nil {
		panic(err)
	}
	r := rand.New(rand.NewSource(int64(today)))

	chosen := make(map[string]int)
	for _, food := range canned {
		chosen[food] = 0
	}
	for i := 0; i < N; i++ {
		food := canned[r.Intn(len(canned))]
		chosen[food]++
	}

	for food, n := range chosen {
		fmt.Printf("%s: %d\n", food, n)
	}
}
