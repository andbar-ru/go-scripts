package main

import (
	"fmt"
	"log"
	"math/rand/v2"
	"os"
	"regexp"
	"strings"
)

var (
	nonAlnumRgx = regexp.MustCompile(`[[:^alnum:]]`)
	b2s         = map[byte]string{
		'0': "0",
		'1': "1",
		'2': "2",
		'3': "3",
		'4': "4",
		'5': "5",
		'6': "6",
		'7': "7",
		'8': "8",
		'9': "9",
		'a': "1",
		'b': "2",
		'c': "3",
		'd': "4",
		'e': "5",
		'f': "6",
		'g': "7",
		'h': "8",
		'i': "9",
		'j': "a",
		'k': "b",
		'l': "c",
		'm': "d",
		'n': "e",
		'o': "f",
		'p': "10",
		'q': "11",
		'r': "12",
		's': "13",
		't': "14",
		'u': "15",
		'v': "16",
		'w': "17",
		'x': "18",
		'y': "19",
		'z': "1a",
	}
	byte16 = []byte{'8', '9', 'a', 'b'}
	rnd    = rand.New(rand.NewPCG(0, 0))
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Argument not specified")
	}

	word := nonAlnumRgx.ReplaceAllString(strings.ToLower(os.Args[1]), "")

	var uuid [32]byte
	uuid[12] = '4'
	uuid[16] = byte16[rnd.IntN(len(byte16))]

	i := -1

OuterLoop:
	for j := 0; j < len(word); j++ {
		b := word[j]
		s := b2s[b]
		for k := 0; k < len(s); k++ {
			i++
			if i >= len(uuid) {
				break OuterLoop
			}
			if uuid[i] != 0 { // байты с индексами 12 и 16 уже заполнены.
				i++
			}
			uuid[i] = s[k]
		}
	}

	for i < len(uuid) {
		i++
		if i >= len(uuid) {
			break
		}
		if uuid[i] != 0 { // байты с индексами 12 и 16 уже заполнены.
			i++
		}
		uuid[i] = '0'
	}

	fmt.Printf("%s-%s-%s-%s-%s\n", uuid[:8], uuid[8:12], uuid[12:16], uuid[16:20], uuid[20:])
}
