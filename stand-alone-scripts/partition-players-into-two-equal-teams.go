/*
Есть множество игроков с разной силой игры. Изначально рейтинг у всех одинаковый. Рейтинг означает просто процент результатов проведённых игр. В разные дни из множества отбираются 12 игроков, делятся на 2 команды по 6 человек и играют несколько волейбольных игр. В конце игрового дня происходит перерасчёт рейтинга игроков в зависимости от результата команды, в которой он играл. В следующие игровые дни игроки переделиваются на команды так, чтобы разница в среднем рейтинге игроков в командах была наименьшей и в конце снова перерасчёт рейтинга. Так они играют много дней. Допускаем, что реальная сила игроков относительно друг друга не меняется. Задача проверить, будет ли рейтинг игроков в таких условиях сильно коррелировать с их реальной силой. Рейтинг игроков находится в диапазоне от 0 до 1. Реальная сила игроков для удобства сравнения тоже будет в пределах от 0 до 1.
*/
package main

import (
	"fmt"
	"math"
	"math/rand"
)

type Player struct {
	name     string
	gender   rune
	strength float64
	games    int
	wins     int
}

func (p *Player) getRating() float64 {
	if p.games == 0 {
		return 0.0
	}
	return float64(p.wins) / float64(p.games)
}

func (p *Player) String() string {
	return fmt.Sprintf("%s: сила %.1f, кол-во игр %d, кол-во побед %d, результативность %.2f\n", p.name, p.strength, p.games, p.wins, p.getRating())
}

var (
	// Все игроки
	allPlayers = []*Player{
		// Мужчины
		&Player{name: "Авдей", gender: 'M', strength: 0.5},
		&Player{name: "Богдан", gender: 'M', strength: 0.6},
		&Player{name: "Вадим", gender: 'M', strength: 0.5},
		&Player{name: "Гавриил", gender: 'M', strength: 0.7},
		&Player{name: "Даниил", gender: 'M', strength: 0.6},
		&Player{name: "Евгений", gender: 'M', strength: 0.7},
		&Player{name: "Ждан", gender: 'M', strength: 0.4},
		&Player{name: "Зиновий", gender: 'M', strength: 0.7},
		&Player{name: "Иакинф", gender: 'M', strength: 0.6},
		&Player{name: "Касьян", gender: 'M', strength: 0.6},
		&Player{name: "Лаврентий", gender: 'M', strength: 0.9},
		&Player{name: "Магистриан", gender: 'M', strength: 1.0},
		&Player{name: "Назар", gender: 'M', strength: 0.5},
		&Player{name: "Олег", gender: 'M', strength: 0.4},
		&Player{name: "Павел", gender: 'M', strength: 0.5},
		&Player{name: "Разумник", gender: 'M', strength: 0.8},
		&Player{name: "Савва", gender: 'M', strength: 0.5},
		&Player{name: "Тарас", gender: 'M', strength: 0.5},
		&Player{name: "Фаддей", gender: 'M', strength: 0.6},
		&Player{name: "Харитон", gender: 'M', strength: 0.7},
		&Player{name: "Эдуард", gender: 'M', strength: 0.6},
		&Player{name: "Юлиан", gender: 'M', strength: 0.5},
		&Player{name: "Яков", gender: 'M', strength: 0.4},
		// Женщины
		&Player{name: "Агафья", gender: 'F', strength: 0.4},
		&Player{name: "Валентина", gender: 'F', strength: 0.1},
		&Player{name: "Галина", gender: 'F', strength: 0.4},
		&Player{name: "Дана", gender: 'F', strength: 0.6},
		&Player{name: "Евгения", gender: 'F', strength: 0.3},
		&Player{name: "Жанна", gender: 'F', strength: 0.5},
		&Player{name: "Зинаида", gender: 'F', strength: 0.3},
		&Player{name: "Инга", gender: 'F', strength: 0.6},
		&Player{name: "Карина", gender: 'F', strength: 0.4},
		&Player{name: "Лада", gender: 'F', strength: 0.6},
		&Player{name: "Маргарита", gender: 'F', strength: 0.1},
		&Player{name: "Надежда", gender: 'F', strength: 0.3},
		&Player{name: "Оксана", gender: 'F', strength: 0.4},
		&Player{name: "Пелагея", gender: 'F', strength: 0.3},
		&Player{name: "Рада", gender: 'F', strength: 0.6},
		&Player{name: "Светлана", gender: 'F', strength: 0.2},
		&Player{name: "Таисия", gender: 'F', strength: 0.4},
		&Player{name: "Ульяна", gender: 'F', strength: 0.4},
		&Player{name: "Фаина", gender: 'F', strength: 0.5},
		&Player{name: "Целестина", gender: 'F', strength: 0.1},
		&Player{name: "Юлия", gender: 'F', strength: 0.6},
		&Player{name: "Яна", gender: 'F', strength: 0.7},
	}
	dayPlayers [12]*Player // игроки в текущий день
)

// getPlayersRatingSum возвращает сумму рейтингов списка игроков.
func getPlayersRatingSum(players []*Player) float64 {
	var sum float64
	for _, player := range players {
		sum += player.getRating()
	}
	return sum
}

// getPlayersStrengthSum возвращает сумму сил списка игроков.
func getPlayersStrengthSum(players []*Player) float64 {
	var sum float64
	for _, player := range players {
		sum += player.strength
	}
	return sum
}

type TeamPair struct {
	team1        []*Player
	ratingSum1   float64
	strengthSum1 float64
	team2        []*Player
	ratingSum2   float64
	strengthSum2 float64
}

func (tp *TeamPair) String() string {
	s := "Команда 1\n"
	for _, player := range tp.team1 {
		s += player.String()
	}
	s += "\n"
	s += "Команда 2\n"
	for _, player := range tp.team2 {
		s += player.String()
	}
	return s
}

// ratingSumDiff возвращает разницу сумм рейтингов игроков команд.
func (tp *TeamPair) ratingSumDiff() float64 {
	return math.Abs(tp.ratingSum1 - tp.ratingSum2)
}

// play играет игру и записывает результаты.
func (tp *TeamPair) play() {
	// Шансы первой команды победить в отдельно взятой игре
	threshold := tp.strengthSum1 / (tp.strengthSum1 + tp.strengthSum2)
	fmt.Println(threshold)
	team1Won := rand.Float64() < threshold
	// Обновляем статистику игроков в зависимости от результата игры.
	for _, player := range tp.team1 {
		player.games++
		if team1Won {
			player.wins++
		}
	}
	for _, player := range tp.team2 {
		player.games++
		if !team1Won {
			player.wins++
		}
	}
}

// pickDayPlayers отбирает 12 игроков из allPlayers и помещает их в массив dayPlayers.
func pickDayPlayers() {
	indexes := rand.Perm(len(allPlayers))[:12]
	for i, index := range indexes {
		dayPlayers[i] = allPlayers[index]
	}
}

// getMoreEqualTeamPair возвращает ту пару команд, где разница между суммарными рейтингами команд меньше.
func getMoreEqualTeamPair(tp1, tp2 TeamPair) TeamPair {
	if tp1.ratingSumDiff() <= tp2.ratingSumDiff() {
		return tp1
	} else {
		return tp2
	}
}

// partitionDayPlayers делит игроков dayPlayers на 2 максимально близкие по среднему рейтингу команды.
func findEqualTeamPair(tp TeamPair, i int) TeamPair {
	// Если размер какой-нибудь команды достиг половины dayPlayers, оставшихся людей сунуть в другую группу и вернуть результат.
	if len(tp.team1) == len(dayPlayers)/2 || len(tp.team2) == len(dayPlayers)/2 {
		if len(tp.team1) == len(dayPlayers)/2 {
			tp.team2 = append(tp.team2, dayPlayers[i:]...)
			tp.ratingSum2 += getPlayersRatingSum(dayPlayers[i:])
			tp.strengthSum2 += getPlayersStrengthSum(dayPlayers[i:])
		} else {
			tp.team1 = append(tp.team1, dayPlayers[i:]...)
			tp.ratingSum1 += getPlayersRatingSum(dayPlayers[i:])
			tp.strengthSum1 += getPlayersStrengthSum(dayPlayers[i:])
		}
		return tp
	}

	// Каждого из dayPlayers мы можем сунуть либо в одну либо в другую команду. Надо найти лучший вариант из двух.
	return getMoreEqualTeamPair(
		findEqualTeamPair(
			TeamPair{
				append(tp.team1, dayPlayers[i]),
				tp.ratingSum1 + dayPlayers[i].getRating(),
				tp.strengthSum1 + dayPlayers[i].strength,
				append([]*Player{}, tp.team2...),
				tp.ratingSum2,
				tp.strengthSum2,
			},
			i+1,
		),
		findEqualTeamPair(
			TeamPair{
				append([]*Player{}, tp.team1...),
				tp.ratingSum1,
				tp.strengthSum1,
				append(tp.team2, dayPlayers[i]),
				tp.ratingSum2 + dayPlayers[i].getRating(),
				tp.strengthSum2 + dayPlayers[i].strength,
			},
			i+1,
		),
	)
}

func main() {
	// Инициировать генератор случайных чисел.
	rand.Seed(42)

	// Отобрать случайным образом из множества 12 игроков.
	pickDayPlayers()

	// Разделить игроков на 2 равные макисмально близкие по среднему рейтингу команды.
	teamPair := findEqualTeamPair(TeamPair{[]*Player{}, 0, 0, []*Player{}, 0, 0}, 0)
	fmt.Println(teamPair.String())

	teamPair.play()
	fmt.Println(teamPair.String())
}
