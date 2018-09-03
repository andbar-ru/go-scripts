/*
Скрипт генерирует пароль по фразе.
Настоятельно рекомендуется перед фразой вводить мастер-пароль, например:
genpass '<мастер-пароль><фраза>'
*/
package main

import (
	"crypto/sha512"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"time"
)

const (
	lDesc = "включить в набор строчные буквы"
	uDesc = "включить в набор прописные буквы"
	dDesc = "включить в набор цифры"
	sDesc = "включить в набор спецсимволы"
	cDesc = "количество символов, из которого будет состоять пароль, по умолчанию 12, максимум 64"
	nDesc = "количество генерируемых паролей, если не заданы фразы, иначе определяется по числу фраз"
	hDesc = "справка по программе"
)

var (
	// Символы в группах в порядке увеличения кодов ascii.
	lower   = []byte("abcdefghijklmnopqrstuvwxyz")
	upper   = []byte("ABCDEFGHIJKLMNOPQRSTUVWXYZ")
	digits  = []byte("0123456789")
	symbols = []byte(" !\"#$%&'()*+,-./:;<=>?@[\\]^_`{|}~")

	charactersCapacity = len(lower) + len(upper) + len(digits) + len(symbols)
)

func printHelp() {
	fmt.Println("Использование: genpass [ключи] [[мастер-пароль]фраза]...")
	fmt.Println("Генерировать пароль(и) по фразе(ам) или случайным образом.")
	fmt.Println("Перед фразами рекомендуется добавлять мастер-пароль.")
	fmt.Println("Ключи. Если никакие ключи не указаны, то используются -l -u -d -s -c 12:")
	fmt.Printf("  -l  %s\n", lDesc)
	fmt.Printf("  -u  %s\n", uDesc)
	fmt.Printf("  -d  %s\n", dDesc)
	fmt.Printf("  -s  %s\n", sDesc)
	fmt.Printf("  -c <число>  %s\n", cDesc)
	fmt.Println()
	fmt.Printf("  -n <число>  %s\n", nDesc)
	fmt.Printf("  -h  %s\n", hDesc)

	os.Exit(0)
}

func main() {
	if len(os.Args) == 1 {
		printHelp()
	}

	// Объявления ключей
	l := flag.Bool("l", false, lDesc)
	u := flag.Bool("u", false, uDesc)
	d := flag.Bool("d", false, dDesc)
	s := flag.Bool("s", false, sDesc)
	c := flag.Int("c", 12, cDesc)
	n := flag.Int("n", 1, nDesc)
	h := flag.Bool("h", false, "Справка")

	flag.Parse()

	if *h {
		printHelp()
	}

	// Проверка ключей
	if *c > 64 || *c < 1 {
		fmt.Fprintf(os.Stderr, "Количество символов (-c=%d) неправильное, должно быть от 1 до 64\n", *c)
		os.Exit(1)
	}
	if *n < 1 {
		fmt.Fprintf(os.Stderr, "Количество паролей (-n=%d) неправильное, должно быть больше 0\n", *n)
		os.Exit(1)
	}

	// Если не указано, какие символы использовать при генерации пароля, использовать все возможные.
	if (*l || *u || *d || *s) == false {
		*l, *u, *d, *s = true, true, true, true
	}

	phrases := flag.Args()

	// Количество паролей по числу фраз
	if len(phrases) > 0 {
		*n = len(phrases)
	}

	// Набор символов для генерации пароля
	var characters = make([]byte, 0, charactersCapacity)
	if *l {
		characters = append(characters, lower...)
	}
	if *u {
		characters = append(characters, upper...)
	}
	if *d {
		characters = append(characters, digits...)
	}
	if *s {
		characters = append(characters, symbols...)
	}
	charactersLen := len(characters)

	// Единый контейнер для паролей
	password := make([]byte, *c, *c)

	if len(phrases) == 0 {
		// Если не заданы фразы, то генерируются рандомные пароли.
		rand.Seed(time.Now().UnixNano())
		for i := 0; i < *n; i++ {
			for i := range password {
				password[i] = characters[rand.Intn(charactersLen)]
			}
			fmt.Printf("%s\n", password)
		}
	} else {
		// Если фразы заданы, то для каждой фразы по определённому алгоритму вычисляется пароль.
		for _, phrase := range phrases {
			sum := sha512.Sum512([]byte(phrase))
			for i := range password {
				password[i] = characters[int(sum[i])%charactersLen]
			}
			fmt.Printf("%s\n", password)
		}
	}
}
