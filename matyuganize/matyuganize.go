package matyuganize

import (
	"strings"
	"unicode"
)

var prefixes = map[[2]rune]struct{}{
	{'н', 'а'}: {},
	{'н', 'е'}: {},
	{'н', 'и'}: {},
}

// Matyuganize возвращает матюганизированную версию строки s.
// Термин "матюганизация" введён в обиход Андреем Барташевичем в 2023 году и
// означает раздельное написание "на", "не" и "ни" независимо от грамматических правил.
func Matyuganize(s string) string {
	matyuganized := make([]rune, 0, len(s))

	wordBeginning := true
	var prefix [2]rune

	clearPrefix := func() {
		for i := range prefix {
			prefix[i] = 0
		}
	}

	processRune := func(r rune) {
		defer func() {
			matyuganized = append(matyuganized, r)
		}()

		if !unicode.IsLetter(r) {
			if !wordBeginning {
				wordBeginning = true
			}
			if prefix[0] != 0 {
				clearPrefix()
			}
			return
		}

		if !wordBeginning {
			return
		}

		letter := unicode.ToLower(r)

		if letter == 'н' && prefix[0] == 0 {
			prefix[0] = letter
		} else if (letter == 'а' || letter == 'е' || letter == 'и') && prefix[0] == 'н' && prefix[1] == 0 {
			prefix[1] = letter
		} else if _, ok := prefixes[prefix]; ok {
			matyuganized = append(matyuganized, ' ')
			clearPrefix()
			if letter == 'н' {
				prefix[0] = letter
			}
		} else {
			clearPrefix()
			wordBeginning = false
		}
	}

	for _, r := range s {
		processRune(r)
	}

	return string(matyuganized)
}

// Версия функции Matyuganize от друга Максима Атюганова.
func Matyuganize1(s string) string {
	var resultBuilder strings.Builder
	resultBuilder.Grow(len(s) + len(s)/4)

	runeReader := strings.NewReader(s)
	prev, _, err := runeReader.ReadRune()
	if err != nil { // eof
		return resultBuilder.String()
	}
	resultBuilder.WriteRune(prev)
	cur, _, err := runeReader.ReadRune()
	if err != nil { // eof
		return resultBuilder.String()
	}
	resultBuilder.WriteRune(cur)

	next, _, err := runeReader.ReadRune()
	for err == nil { // not eof, next is valid
		if prev == 'н' &&
			(cur == 'а' || cur == 'е' || cur == 'и') &&
			next != ' ' && !unicode.IsPunct(next) {
			resultBuilder.WriteRune(' ')
		}
		resultBuilder.WriteRune(next)
		prev = cur
		cur = next
		next, _, err = runeReader.ReadRune()
	}

	return resultBuilder.String()
}
