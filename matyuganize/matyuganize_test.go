package matyuganize

import (
	"os"
	"testing"
)

var src2want = map[string]string{
	"Не дрейфить, не наезжать, неплохой, нанавижу, ненадолго, не такой как все, НАТО, нельзя, Ниневия — столица Ассирии, броневик": "Не дрейфить, не на езжать, не плохой, на на вижу, не на долго, не такой как все, НА ТО, не льзя, Ни не вия — столица Ассирии, броневик",
	"накидка":  "на кидка",
	"наклейка": "на клейка",
	"бла бла нестыковка тут":         "бла бла не стыковка тут",
	"здесь висит картина, а там нет": "здесь висит картина, а там не т",
	"нисколько, тяни, тян":           "ни сколько, тяни, тян",
	"книга":                          "книга",
}

func TestMatyuganize(t *testing.T) {
	for src, want := range src2want {
		res := Matyuganize(src)
		if res != want {
			t.Errorf("Matyuganize(%q) = %q, want %q", src, res, want)
		}
	}
}

func TestMatyuganize1(t *testing.T) {
	for src, want := range src2want {
		res := Matyuganize1(src)
		if res != want {
			t.Errorf("Matyuganize(%q) = %q, want %q", src, res, want)
		}
	}
}

func TestWarAndPiece(t *testing.T) {
	file, err := os.ReadFile("./WarAndPieceBook1.txt")
	if err != nil {
		panic(err)
	}
	s := string(file)
	len0 := len(s)
	matyuganized := Matyuganize(s)
	matyuganized1 := Matyuganize1(s)
	if len(matyuganized) <= len0 {
		t.Errorf("Function \"Matyuganize\" dosn't work")
	}
	if len(matyuganized1) <= len0 {
		t.Errorf("Function \"Matyuganize1\" dosn't work")
	}
	if len(matyuganized) != len(matyuganized1) {
		t.Errorf("Matyuganize and Matyuganize1 return different results")
		/*
			err = os.WriteFile("WarAndPieceMatyuganized.txt", []byte(matyuganized), 0666)
			if err != nil {
				panic(err)
			}
			err = os.WriteFile("WarAndPieceMatyuganized1.txt", []byte(matyuganized1), 0666)
			if err != nil {
				panic(err)
			}
		*/
	}
}
