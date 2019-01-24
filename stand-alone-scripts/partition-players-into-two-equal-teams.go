/*
Есть множество игроков с разной силой игры. Изначально рейтинг у всех одинаковый. Рейтинг означает просто процент результатов проведённых игр. В разные дни из множества отбираются 12 игроков, делятся на 2 команды по 6 человек и играют несколько волейбольных игр. В конце игрового дня происходит перерасчёт рейтинга игроков в зависимости от результата команды, в которой он играл. В следующие игровые дни игроки переделиваются на команды так, чтобы разница в среднем рейтинге игроков в командах была наименьшей и в конце снова перерасчёт рейтинга. Так они играют много дней. Допускаем, что реальная сила игроков относительно друг друга не меняется. Задача проверить, будет ли рейтинг игроков в таких условиях сильно коррелировать с их реальной силой. Рейтинг игроков находится в диапазоне от 0 до 1. Реальная сила игроков для удобства сравнения тоже будет в пределах от 0 до 1.
*/
package main

import (
	"fmt"
	"math"
	"math/rand"
)

func getPercent(dividend, divider int) float64 {
	if divider == 0 {
		return 0
	}
	return float64(divider*100) / float64(dividend)
}

type Player struct {
	name     string
	gender   rune
	strength float64
	games    [2]int // всего и выигранных
}

func (p *Player) getSuccessRate() float64 {
	if p.games[0] == 0 {
		return 0.0
	}
	return float64(p.games[1]) / float64(p.games[0])
}

func (p Player) String() string {
	return fmt.Sprintf("%s: сила %.1f, кол-во игр %d, кол-во побед %d, успешность %.2f\n", p.name, p.strength, p.games[0], p.games[1], p.getSuccessRate())
}

type Stats struct {
	games int
	days  int
	// Игры команд из пар по признакам: всего и выигранных:
	// по силе:
	moreStrongTeamGames  [2]int
	lessStrongTeamGames  [2]int
	equalStrongTeamGames [2]int
	// по успешности -- доле выигранных игр:
	moreSuccessfulTeamGames  [2]int
	lessSuccessfulTeamGames  [2]int
	equalSuccessfulTeamGames [2]int
}

func (s Stats) String() string {
	str := fmt.Sprintf("Количество игр: %d\n", s.games)
	str += fmt.Sprintf("Количество дней: %d\n", s.days)

	str += fmt.Sprintln("Статистика команд по признакам:")
	str += fmt.Sprintf("  более сильные: %d из %d (%.1f %%)\n", s.moreStrongTeamGames[1], s.moreStrongTeamGames[0], getPercent(s.moreStrongTeamGames[0], s.moreStrongTeamGames[1]))
	str += fmt.Sprintf("  менее сильные: %d из %d (%.1f %%)\n", s.lessStrongTeamGames[1], s.lessStrongTeamGames[0], getPercent(s.lessStrongTeamGames[0], s.lessStrongTeamGames[1]))
	str += fmt.Sprintf("  равные по силе: %d из %d (%.1f %%)\n", s.equalStrongTeamGames[1], s.equalStrongTeamGames[0], getPercent(s.equalStrongTeamGames[0], s.equalStrongTeamGames[1]))
	str += "\n"
	str += fmt.Sprintf("  более успешные: %d из %d (%.1f %%)\n", s.moreSuccessfulTeamGames[1], s.moreSuccessfulTeamGames[0], getPercent(s.moreSuccessfulTeamGames[0], s.moreSuccessfulTeamGames[1]))
	str += fmt.Sprintf("  менее успешные: %d из %d (%.1f %%)\n", s.lessSuccessfulTeamGames[1], s.lessSuccessfulTeamGames[0], getPercent(s.lessSuccessfulTeamGames[0], s.lessSuccessfulTeamGames[1]))
	str += fmt.Sprintf("  равные по успеху: %d из %d (%.1f %%)\n", s.equalSuccessfulTeamGames[1], s.equalSuccessfulTeamGames[0], getPercent(s.equalSuccessfulTeamGames[0], s.equalSuccessfulTeamGames[1]))

	return str
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
	stats      = Stats{}
)

// getPlayersSuccessRateSum возвращает сумму рейтингов списка игроков.
func getPlayersSuccessRateSum(players []*Player) float64 {
	var sum float64
	for _, player := range players {
		sum += player.getSuccessRate()
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
	successSum1  float64
	strengthSum1 float64
	team2        []*Player
	successSum2  float64
	strengthSum2 float64
}

func (tp TeamPair) String() string {
	s := "Команда 1\n"
	for _, player := range tp.team1 {
		s += fmt.Sprint(player)
	}
	s += "\n"
	s += "Команда 2\n"
	for _, player := range tp.team2 {
		s += fmt.Sprint(player)
	}
	return s
}

// ratingSumDiff возвращает разницу сумм рейтингов игроков команд.
func (tp *TeamPair) ratingSumDiff() float64 {
	return math.Abs(tp.successSum1 - tp.successSum2)
}

// play играет игру и записывает результаты.
func (tp *TeamPair) play() {
	// Более и менее сильные и более и менее успешные команды.
	var moreStrongTeam, lessStrongTeam, moreSuccessfulTeam, lessSuccessfulTeam []*Player
	if tp.strengthSum1 != tp.strengthSum2 {
		if tp.strengthSum1 > tp.strengthSum2 {
			moreStrongTeam = tp.team1
			lessStrongTeam = tp.team2
		} else {
			moreStrongTeam = tp.team2
			lessStrongTeam = tp.team1
		}
	}
	if tp.successSum1 != tp.successSum2 {
		if tp.successSum1 > tp.successSum2 {
			moreSuccessfulTeam = tp.team1
			lessSuccessfulTeam = tp.team2
		} else {
			moreSuccessfulTeam = tp.team2
			lessSuccessfulTeam = tp.team1
		}
	}

	// Шансы первой команды победить в отдельно взятой игре.
	threshold := tp.strengthSum1 / (tp.strengthSum1 + tp.strengthSum2)
	team1Won := rand.Float64() < threshold

	/* Обновляем статистику. */
	// Независимую от результата игры.
	stats.games++
	// По силе.
	if moreStrongTeam == nil && lessStrongTeam == nil {
		stats.equalStrongTeamGames[0] += 2 // игры у каждой команды
		stats.equalStrongTeamGames[1] += 1 // победа только у одной
	} else {
		stats.moreStrongTeamGames[0]++
		stats.lessStrongTeamGames[0]++
	}
	// По успеху.
	if moreSuccessfulTeam == nil && lessSuccessfulTeam == nil {
		stats.equalSuccessfulTeamGames[0] += 2 // игры у каждой команды
		stats.equalSuccessfulTeamGames[1] += 1 // победа только у одной
	} else {
		stats.moreSuccessfulTeamGames[0]++
		stats.lessSuccessfulTeamGames[0]++
	}
	// Обновляем счётчик сыгранных игр обеих команд.
	for _, player := range tp.team1 {
		player.games[0]++
	}
	for _, player := range tp.team2 {
		player.games[0]++
	}

	// В зависимости от результата игры.
	if team1Won {
		if moreStrongTeam != nil && lessStrongTeam != nil {
			if moreStrongTeam[0] == tp.team1[0] {
				stats.moreStrongTeamGames[1]++
			} else {
				stats.lessStrongTeamGames[1]++
			}
		}
		for _, player := range tp.team1 {
			player.games[1]++
		}
	} else {
		if moreStrongTeam != nil && lessStrongTeam != nil {
			if moreStrongTeam[0] == tp.team2[0] {
				stats.moreStrongTeamGames[1]++
			} else {
				stats.lessStrongTeamGames[1]++
			}
		}
		for _, player := range tp.team2 {
			player.games[1]++
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
			tp.successSum2 += getPlayersSuccessRateSum(dayPlayers[i:])
			tp.strengthSum2 += getPlayersStrengthSum(dayPlayers[i:])
		} else {
			tp.team1 = append(tp.team1, dayPlayers[i:]...)
			tp.successSum1 += getPlayersSuccessRateSum(dayPlayers[i:])
			tp.strengthSum1 += getPlayersStrengthSum(dayPlayers[i:])
		}
		return tp
	}

	// Каждого из dayPlayers мы можем сунуть либо в одну либо в другую команду. Надо найти лучший вариант из двух.
	return getMoreEqualTeamPair(
		findEqualTeamPair(
			TeamPair{
				append(tp.team1, dayPlayers[i]),
				tp.successSum1 + dayPlayers[i].getSuccessRate(),
				tp.strengthSum1 + dayPlayers[i].strength,
				append([]*Player{}, tp.team2...),
				tp.successSum2,
				tp.strengthSum2,
			},
			i+1,
		),
		findEqualTeamPair(
			TeamPair{
				append([]*Player{}, tp.team1...),
				tp.successSum1,
				tp.strengthSum1,
				append(tp.team2, dayPlayers[i]),
				tp.successSum2 + dayPlayers[i].getSuccessRate(),
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

	// Разделить игроков на 2 равные максимально близкие по среднему рейтингу команды.
	teamPair := findEqualTeamPair(
		TeamPair{
			team1:        []*Player{},
			successSum1:  0,
			strengthSum1: 0,
			team2:        []*Player{},
			successSum2:  0,
			strengthSum2: 0,
		},
		0,
	)
	fmt.Println(teamPair)

	teamPair.play()
	fmt.Println(teamPair)

	fmt.Println(stats)
}
