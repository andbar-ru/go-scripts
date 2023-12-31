package main

import (
	"fmt"
	"math/rand"
)

const DEBUG = false
const N = 1 << 20

var (
	successCount int
	failCount    int
)

func filter(slice []int, f func(int) bool) []int {
	newSlice := make([]int, 0, len(slice))
	for _, v := range slice {
		if f(v) {
			newSlice = append(newSlice, v)
		}
	}
	return newSlice
}

func debug(format string, args ...interface{}) {
	if DEBUG {
		fmt.Printf(format, args...)
	}
}

func noSimulation() {
	for i := 0; i < N; i++ {
		successIndex := rand.Intn(3)
		answerIndex := rand.Intn(3)
		debug("%d %d\n", successIndex, answerIndex)
		if answerIndex == successIndex {
			successCount++
		} else {
			failCount++
		}
	}
}

func answerNotChanged() {
	for i := 0; i < N; i++ {
		var doors [3]bool
		allIndexes := []int{0, 1, 2}
		successIndex := rand.Intn(len(doors))
		doors[successIndex] = true
		debug("doors: %v\n", doors)
		answerIndex := rand.Intn(len(doors))
		debug("First answer: %d\n", answerIndex)
		otherIndexes := filter(allIndexes, func(v int) bool { return v != answerIndex })
		otherFailIndexes := filter(otherIndexes, func(v int) bool { return v != successIndex })
		randFailIndex := otherFailIndexes[rand.Intn(len(otherFailIndexes))]
		otherIndexesExcludingOneFail := filter(otherIndexes, func(v int) bool { return v != randFailIndex })
		secondChooseIndexes := make([]int, 0, 2)
		secondChooseIndexes = append(secondChooseIndexes, answerIndex)
		secondChooseIndexes = append(secondChooseIndexes, otherIndexesExcludingOneFail...)
		debug("Second choose: %v\n", secondChooseIndexes)
		answer2Index := answerIndex
		debug("Second answer: %d\n", answer2Index)
		if answer2Index == successIndex {
			debug("Success\n")
			successCount++
		} else {
			debug("Fail\n")
			failCount++
		}
	}
}

func answerChanged() {
	for i := 0; i < N; i++ {
		var doors [3]bool
		allIndexes := []int{0, 1, 2}
		successIndex := rand.Intn(len(doors))
		doors[successIndex] = true
		debug("doors: %v\n", doors)
		answerIndex := rand.Intn(len(doors))
		debug("First answer: %d\n", answerIndex)
		otherIndexes := filter(allIndexes, func(v int) bool { return v != answerIndex })
		otherFailIndexes := filter(otherIndexes, func(v int) bool { return v != successIndex })
		randFailIndex := otherFailIndexes[rand.Intn(len(otherFailIndexes))]
		otherIndexesExcludingOneFail := filter(otherIndexes, func(v int) bool { return v != randFailIndex })
		secondChooseIndexes := make([]int, 0, 2)
		secondChooseIndexes = append(secondChooseIndexes, answerIndex)
		secondChooseIndexes = append(secondChooseIndexes, otherIndexesExcludingOneFail...)
		debug("Second choose: %v\n", secondChooseIndexes)
		var answer2Index int
		for _, i := range secondChooseIndexes {
			if i != answerIndex {
				answer2Index = i
				break
			}
		}
		debug("Second answer: %d\n", answer2Index)
		if answer2Index == successIndex {
			debug("Success\n")
			successCount++
		} else {
			debug("Fail\n")
			failCount++
		}
	}
}

func main() {
	answerChanged()

	fmt.Println()
	fmt.Printf("Success: %d\n", successCount)
	fmt.Printf("Fail: %d\n", failCount)
	fmt.Printf("Ratio: %.4f\n", float64(successCount)/float64(N))
}
