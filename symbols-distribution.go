/*
Изучаем распределение символов в паролях, выдаваемых утилитой genpass.
Частота появление для символов неодинакова, потому что символов 95, а байт - 256.
В связи с этим ожидаем, что частота появления символов "$%&'()*+,-./:;<=>?@[\\]^_`{|}~" 2*n,
а остальных символов - 3*n, где n = N/4.
*/
package main

import (
	"fmt"
	"log"
	"os/exec"
	"strconv"
	"sync"
)

const (
	N = 1000000 // количество вызовов genpass
)

var (
	symbols     = []byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789 !\"#$%&'()*+,-./:;<=>?@[\\]^_`{|}~")
	processChan = make(chan bool, 100) // Максимум 100 одновременно запущенных процессов
)

type symbolsCounter struct {
	v   map[byte]int
	mux sync.Mutex
}

func (c *symbolsCounter) Inc(key byte) {
	c.mux.Lock()
	c.v[key]++
	c.mux.Unlock()
}
func (c *symbolsCounter) Value(key byte) int {
	c.mux.Lock()
	defer c.mux.Unlock()
	return c.v[key]
}

func main() {
	sc := &symbolsCounter{v: make(map[byte]int)}

	var wg sync.WaitGroup
	wg.Add(N)

	for i := 0; i < N; i++ {
		go func(phrase string) {
			processChan <- true
			defer wg.Done()

			out, err := exec.Command("genpass", "-c=64", phrase).Output()
			if err != nil {
				log.Fatal(err)
			}
			for _, b := range out {
				sc.Inc(b)
			}

			<-processChan
		}(strconv.Itoa(i))
	}

	wg.Wait()

	for _, s := range symbols {
		fmt.Printf("%s: %d\n", string(s), sc.Value(s))
	}
}

// Результаты
// a: 750038
// b: 750679
// c: 749068
// d: 749256
// e: 749457
// f: 748437
// g: 749483
// h: 751417
// i: 751782
// j: 747786
// k: 749580
// l: 749411
// m: 750459
// n: 749799
// o: 751090
// p: 750759
// q: 750747
// r: 751597
// s: 751754
// t: 749269
// u: 748060
// v: 751497
// w: 750285
// x: 749717
// y: 749425
// z: 750811
// A: 750444
// B: 750748
// C: 749538
// D: 749281
// E: 750993
// F: 750375
// G: 752533
// H: 748805
// I: 748421
// J: 750338
// K: 749776
// L: 751803
// M: 750115
// N: 749250
// O: 749985
// P: 751004
// Q: 750848
// R: 749363
// S: 752112
// T: 748549
// U: 748838
// V: 750303
// W: 750842
// X: 749976
// Y: 749732
// Z: 749046
// 0: 749514
// 1: 749236
// 2: 749763
// 3: 750135
// 4: 750598
// 5: 750197
// 6: 749561
// 7: 749874
// 8: 750748
// 9: 751126
//  : 749916
// !: 750287
// ": 751312
// #: 749052
// $: 498788
// %: 500212
// &: 499071
// ': 499879
// (: 500291
// ): 501164
// *: 498788
// +: 499869
// ,: 499205
// -: 499977
// .: 499324
// /: 499679
// :: 500386
// ;: 499865
// <: 499736
// =: 499793
// >: 499902
// ?: 499501
// @: 499884
// [: 500958
// \: 499554
// ]: 499465
// ^: 500117
// _: 500181
// `: 499464
// {: 499586
// |: 499692
// }: 500901
// ~: 498768
