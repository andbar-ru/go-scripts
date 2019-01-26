/*
Есть множество игроков с разной силой игры. Изначально рейтинг у всех одинаковый. Рейтинг означает просто процент результатов проведённых игр. В разные дни из множества отбираются 12 игроков, делятся на 2 команды по 6 человек и играют несколько волейбольных игр. В конце игрового дня происходит перерасчёт рейтинга игроков в зависимости от результата команды, в которой он играл. В следующие игровые дни игроки переделиваются на команды так, чтобы разница в среднем рейтинге игроков в командах была наименьшей и в конце снова перерасчёт рейтинга. Так они играют много дней. Допускаем, что реальная сила игроков относительно друг друга не меняется. Задача проверить, будет ли рейтинг игроков в таких условиях сильно коррелировать с их реальной силой. Рейтинг игроков находится в диапазоне от 0 до 1. Реальная сила игроков для удобства сравнения тоже будет в пределах от 0 до 1.
*/
package main

import (
	"fmt"
	"math"
	"math/rand"
)

const (
	GAMES_A_DAY = 8
)

func getPercent(fraction, total float64) float64 {
	if total == 0 {
		return 0
	}
	return (fraction * 100) / total
}

type Player struct {
	name     string
	gender   rune
	strength float64
	games    [2]int // всего и выигранных
	days     [4]int // дневные итоги: всего, выигранных, проигранных и с равным результатом
}

func (p *Player) getSuccessRate() float64 {
	if p.games[0] == 0 {
		return 0.0
	}
	return float64(p.games[1]) / float64(p.games[0])
}

func (p Player) String() string {
	return fmt.Sprintf("%s: сила %.1f; игры: %d из %d, успешность %.2f; дни %d: +%d-%d=%d (%.1f %%)\n",
		p.name, p.strength, p.games[1], p.games[0], p.getSuccessRate(),
		p.days[0], p.days[1], p.days[2], p.days[3],
		getPercent(float64(p.days[1])+float64(p.days[3])/2, float64(p.days[0])))
}

type Stats struct {
	games int
	days  int
	// Игры команд из пар: всего и выигранных:
	// по силе:
	moreStrongTeamGames  [2]int
	lessStrongTeamGames  [2]int
	equalStrongTeamGames [2]int
	// по успешности -- доле выигранных игр:
	moreSuccessfulTeamGames  [2]int
	lessSuccessfulTeamGames  [2]int
	equalSuccessfulTeamGames [2]int
	// Дневные итоги команд из пар: всего, выигранных, проигранных и ничейных.
	// по силе:
	moreStrongTeamDays  [4]int
	lessStrongTeamDays  [4]int
	equalStrongTeamDays [4]int
	// по успешности -- доле выигранных игр:
	moreSuccessfulTeamDays  [4]int
	lessSuccessfulTeamDays  [4]int
	equalSuccessfulTeamDays [4]int
}

func (s Stats) String() string {
	str := fmt.Sprintf("Количество игр: %d\n", s.games)
	str += fmt.Sprintf("Количество дней: %d\n", s.days)

	var percent float64

	str += fmt.Sprintln("Статистика команд по играм:")
	percent = getPercent(float64(s.moreStrongTeamGames[1]), float64(s.moreStrongTeamGames[0]))
	str += fmt.Sprintf("  более сильные: %d из %d (%.1f %%)\n",
		s.moreStrongTeamGames[1], s.moreStrongTeamGames[0], percent)
	percent = getPercent(float64(s.lessStrongTeamGames[1]), float64(s.lessStrongTeamGames[0]))
	str += fmt.Sprintf("  менее сильные: %d из %d (%.1f %%)\n",
		s.lessStrongTeamGames[1], s.lessStrongTeamGames[0], percent)
	percent = getPercent(float64(s.equalStrongTeamGames[1]), float64(s.equalStrongTeamGames[0]))
	str += fmt.Sprintf("  равные по силе: %d из %d (%.1f %%)\n", s.equalStrongTeamGames[1], s.equalStrongTeamGames[0], percent)
	str += "\n"
	percent = getPercent(float64(s.moreSuccessfulTeamGames[1]), float64(s.moreSuccessfulTeamGames[0]))
	str += fmt.Sprintf("  более успешные: %d из %d (%.1f %%)\n", s.moreSuccessfulTeamGames[1], s.moreSuccessfulTeamGames[0], percent)
	percent = getPercent(float64(s.lessSuccessfulTeamGames[1]), float64(s.lessSuccessfulTeamGames[0]))
	str += fmt.Sprintf("  менее успешные: %d из %d (%.1f %%)\n", s.lessSuccessfulTeamGames[1], s.lessSuccessfulTeamGames[0], percent)
	percent = getPercent(float64(s.equalSuccessfulTeamGames[1]), float64(s.equalSuccessfulTeamGames[0]))
	str += fmt.Sprintf("  равные по успеху: %d из %d (%.1f %%)\n", s.equalSuccessfulTeamGames[1], s.equalSuccessfulTeamGames[0], percent)
	str += "\n"

	str += fmt.Sprintln("Статистика команд по игровым дням:")
	percent = getPercent(float64(s.moreStrongTeamDays[1])+float64(s.moreStrongTeamDays[3])/2, float64(s.moreStrongTeamDays[0]))
	str += fmt.Sprintf("  более сильные: %d: +%d-%d=%d (%.1f %%)\n", s.moreStrongTeamDays[0], s.moreStrongTeamDays[1], s.moreStrongTeamDays[2], s.moreStrongTeamDays[3], percent)
	percent = getPercent(float64(s.lessStrongTeamDays[1])+float64(s.lessStrongTeamDays[3])/2, float64(s.lessStrongTeamDays[0]))
	str += fmt.Sprintf("  менее сильные: %d: +%d-%d=%d (%.1f %%)\n", s.lessStrongTeamDays[0], s.lessStrongTeamDays[1], s.lessStrongTeamDays[2], s.lessStrongTeamDays[3], percent)
	percent = getPercent(float64(s.equalStrongTeamDays[1])+float64(s.equalStrongTeamDays[3])/2, float64(s.equalStrongTeamDays[0]))
	str += fmt.Sprintf("  равные по силе: %d: +%d-%d=%d (%.1f %%)\n", s.equalStrongTeamDays[0], s.equalStrongTeamDays[1], s.equalStrongTeamDays[2], s.equalStrongTeamDays[3], percent)
	str += "\n"
	percent = getPercent(float64(s.moreSuccessfulTeamDays[1])+float64(s.moreSuccessfulTeamDays[3])/2, float64(s.moreSuccessfulTeamDays[0]))
	str += fmt.Sprintf("  более успешные: %d: +%d-%d=%d (%.1f %%)\n", s.moreSuccessfulTeamDays[0], s.moreSuccessfulTeamDays[1], s.moreSuccessfulTeamDays[2], s.moreSuccessfulTeamDays[3], percent)
	percent = getPercent(float64(s.lessSuccessfulTeamDays[1])+float64(s.lessSuccessfulTeamDays[3])/2, float64(s.lessSuccessfulTeamDays[0]))
	str += fmt.Sprintf("  менее успешные: %d: +%d-%d=%d (%.1f %%)\n", s.lessSuccessfulTeamDays[0], s.lessSuccessfulTeamDays[1], s.lessSuccessfulTeamDays[2], s.lessSuccessfulTeamDays[3], percent)
	percent = getPercent(float64(s.equalSuccessfulTeamDays[1])+float64(s.equalSuccessfulTeamDays[3])/2, float64(s.equalSuccessfulTeamDays[0]))
	str += fmt.Sprintf("  равные по успеху: %d: +%d-%d=%d (%.1f %%)\n", s.equalSuccessfulTeamDays[0], s.equalSuccessfulTeamDays[1], s.equalSuccessfulTeamDays[2], s.equalSuccessfulTeamDays[3], percent)

	return str
}

// updateGames обновляет статистику по играм.
func (stats *Stats) updateGames(tp *TeamPair, team1Won bool) {
	/* Независимую от результата игры. */
	stats.games++
	// По силе.
	if moreStrongTeamPtr == nil && lessStrongTeamPtr == nil {
		stats.equalStrongTeamGames[0] += 2 // количество игр на 2 команды
		stats.equalStrongTeamGames[1] += 1 // победа только у одной
	} else if moreStrongTeamPtr != nil && lessStrongTeamPtr != nil {
		stats.moreStrongTeamGames[0]++
		stats.lessStrongTeamGames[0]++
	} else {
		panic("Unexpected situation")
	}
	// По успеху.
	if moreSuccessfulTeamPtr == nil && lessSuccessfulTeamPtr == nil {
		stats.equalSuccessfulTeamGames[0] += 2 // количество игр на 2 команды
		stats.equalSuccessfulTeamGames[1] += 1 // победа только у одной
	} else if moreSuccessfulTeamPtr != nil && lessSuccessfulTeamPtr != nil {
		stats.moreSuccessfulTeamGames[0]++
		stats.lessSuccessfulTeamGames[0]++
	} else {
		panic("Unexpected situation")
	}
	// Обновляем счётчик сыгранных игр обеих команд.
	for _, player := range tp.team1 {
		player.games[0]++
	}
	for _, player := range tp.team2 {
		player.games[0]++
	}

	/* Зависимую от результата игры. */
	if team1Won {
		if moreStrongTeamPtr != nil && lessStrongTeamPtr != nil {
			if moreStrongTeamPtr == &tp.team1 {
				stats.moreStrongTeamGames[1]++
			} else if lessStrongTeamPtr == &tp.team1 {
				stats.lessStrongTeamGames[1]++
			} else {
				panic("Unexpected situation")
			}
		}
		if moreSuccessfulTeamPtr != nil && lessSuccessfulTeamPtr != nil {
			if moreSuccessfulTeamPtr == &tp.team1 {
				stats.moreSuccessfulTeamGames[1]++
			} else if lessSuccessfulTeamPtr == &tp.team1 {
				stats.lessSuccessfulTeamGames[1]++
			} else {
				panic("Unexpected situation")
			}
		}
		for _, player := range tp.team1 {
			player.games[1]++
		}
	} else {
		if moreStrongTeamPtr != nil && lessStrongTeamPtr != nil {
			if moreStrongTeamPtr == &tp.team2 {
				stats.moreStrongTeamGames[1]++
			} else if lessStrongTeamPtr == &tp.team2 {
				stats.lessStrongTeamGames[1]++
			} else {
				panic("Unexpected situation")
			}
		}
		if moreSuccessfulTeamPtr != nil && lessSuccessfulTeamPtr != nil {
			if moreSuccessfulTeamPtr == &tp.team2 {
				stats.moreSuccessfulTeamGames[1]++
			} else if lessSuccessfulTeamPtr == &tp.team2 {
				stats.lessSuccessfulTeamGames[1]++
			} else {
				panic("Unexpected situation")
			}
		}
		for _, player := range tp.team2 {
			player.games[1]++
		}
	}
}

// updateDays обновляет статистику по дням.
func (stats *Stats) updateDays(tp *TeamPair, results map[bool]int) {
	/* Независимую от результатов. */
	stats.days++
	// По силе
	if moreStrongTeamPtr == nil && lessStrongTeamPtr == nil {
		stats.equalStrongTeamDays[0] += 2 // количество дней на 2 команды
	} else {
		stats.moreStrongTeamDays[0]++
		stats.lessStrongTeamDays[0]++
	}
	// По успеху.
	if moreSuccessfulTeamPtr == nil && lessSuccessfulTeamPtr == nil {
		stats.equalSuccessfulTeamDays[0] += 2 // количество дней на 2 команды
	} else {
		stats.moreSuccessfulTeamDays[0]++
		stats.lessSuccessfulTeamDays[0]++
	}
	// Обновляем счётчик сыгранных дней обеих команд.
	for _, player := range tp.team1 {
		player.days[0]++
	}
	for _, player := range tp.team2 {
		player.days[0]++
	}

	/* Зависимую от результатов. */
	if results[true] == results[false] {
		// по силе
		if moreStrongTeamPtr == nil && lessStrongTeamPtr == nil {
			stats.equalStrongTeamDays[3] += 2 // для каждой команды
		} else if moreStrongTeamPtr != nil && lessStrongTeamPtr != nil {
			stats.moreStrongTeamDays[3]++
			stats.lessStrongTeamDays[3]++
		} else {
			panic("Unexpected situation")
		}
		// по успешности
		if moreSuccessfulTeamPtr == nil && lessSuccessfulTeamPtr == nil {
			stats.equalSuccessfulTeamDays[3] += 2 // для каждой команды
		} else if moreSuccessfulTeamPtr != nil && lessSuccessfulTeamPtr != nil {
			stats.moreSuccessfulTeamDays[3]++
			stats.lessSuccessfulTeamDays[3]++
		} else {
			panic("Unexpected situation")
		}
		// счётчик сыгранных дней обеих команд
		for _, player := range tp.team1 {
			player.days[3]++
		}
		for _, player := range tp.team2 {
			player.days[3]++
		}
	} else if results[true] > results[false] {
		// по силе
		if moreStrongTeamPtr == nil && lessStrongTeamPtr == nil {
			// Для одной команды удачный день, для второй -- неудачный.
			stats.equalStrongTeamDays[1]++
			stats.equalStrongTeamDays[2]++
		} else if moreStrongTeamPtr != nil && lessStrongTeamPtr != nil {
			if moreStrongTeamPtr == &tp.team1 {
				stats.moreStrongTeamDays[1]++
				stats.lessStrongTeamDays[2]++
			} else if lessStrongTeamPtr == &tp.team1 {
				stats.lessStrongTeamDays[1]++
				stats.moreStrongTeamDays[2]++
			} else {
				panic("Unexpected situation")
			}
		} else {
			panic("Unexpected situation")
		}
		// по успешности
		if moreSuccessfulTeamPtr == nil && lessSuccessfulTeamPtr == nil {
			// Для одной команды удачный день, для второй -- неудачный.
			stats.equalSuccessfulTeamDays[1]++
			stats.equalSuccessfulTeamDays[2]++
		} else if moreSuccessfulTeamPtr != nil && lessSuccessfulTeamPtr != nil {
			if moreSuccessfulTeamPtr == &tp.team1 {
				stats.moreSuccessfulTeamDays[1]++
				stats.lessSuccessfulTeamDays[2]++
			} else if lessSuccessfulTeamPtr == &tp.team1 {
				stats.lessSuccessfulTeamDays[1]++
				stats.moreSuccessfulTeamDays[2]++
			} else {
				panic("Unexpected situation")
			}
		} else {
			panic("Unexpected situation")
		}
		// счётчик сыгранных дней обеих команд
		for _, player := range tp.team1 {
			player.days[1]++
		}
		for _, player := range tp.team2 {
			player.days[2]++
		}
	} else if results[true] < results[false] {
		// по силе
		if moreStrongTeamPtr == nil && lessStrongTeamPtr == nil {
			// Для одной команды удачный день, для второй -- неудачный.
			stats.equalStrongTeamDays[1]++
			stats.equalStrongTeamDays[2]++
		} else if moreStrongTeamPtr != nil && lessStrongTeamPtr != nil {
			if moreStrongTeamPtr == &tp.team1 {
				stats.moreStrongTeamDays[2]++
				stats.lessStrongTeamDays[1]++
			} else if lessStrongTeamPtr == &tp.team1 {
				stats.lessStrongTeamDays[2]++
				stats.moreStrongTeamDays[1]++
			} else {
				panic("Unexpected situation")
			}
		} else {
			panic("Unexpected situation")
		}
		// по успешности
		if moreSuccessfulTeamPtr == nil && lessSuccessfulTeamPtr == nil {
			// Для одной команды удачный день, для второй -- неудачный.
			stats.equalSuccessfulTeamDays[1]++
			stats.equalSuccessfulTeamDays[2]++
		} else if moreSuccessfulTeamPtr != nil && lessSuccessfulTeamPtr != nil {
			if moreSuccessfulTeamPtr == &tp.team1 {
				stats.moreSuccessfulTeamDays[2]++
				stats.lessSuccessfulTeamDays[1]++
			} else if lessSuccessfulTeamPtr == &tp.team1 {
				stats.lessSuccessfulTeamDays[2]++
				stats.moreSuccessfulTeamDays[1]++
			} else {
				panic("Unexpected situation")
			}
		} else {
			panic("Unexpected situation")
		}
		// счётчик сыгранных дней обеих команд
		for _, player := range tp.team1 {
			player.days[2]++
		}
		for _, player := range tp.team2 {
			player.days[1]++
		}
	} else {
		panic("Unexpected situation")
	}
}

// validate проверяет, что статистика согласована, иначе выводит ошибки.
func (stats *Stats) validate() {
	errs := make([]string, 0)

	if stats.games != stats.days*GAMES_A_DAY {
		errs = append(errs, "Количество игр не согласовано с количеством дней и количеством игр в день.")
	}

	// Проверяем статистику по играм.
	if stats.moreStrongTeamGames[0]+stats.lessStrongTeamGames[0]+stats.equalStrongTeamGames[0] != stats.games*2 {
		errs = append(errs, "Суммарное количество игр в категориях по силе не согласовано с общим количеством игр.")
	}

	if stats.moreSuccessfulTeamGames[0]+stats.lessSuccessfulTeamGames[0]+stats.equalSuccessfulTeamGames[0] != stats.games*2 {
		errs = append(errs, "Суммарное количество игр в категориях по успешности не согласовано с общим количеством игр.")
	}

	if stats.moreStrongTeamGames[1]+stats.lessStrongTeamGames[1]+stats.equalStrongTeamGames[1] != stats.games {
		errs = append(errs, "Суммарное количество выигранных игр в категориях по силе не согласовано с общим количеством игр.")
	}

	if stats.moreSuccessfulTeamGames[1]+stats.lessSuccessfulTeamGames[1]+stats.equalSuccessfulTeamGames[1] != stats.games {
		errs = append(errs, "Суммарное количество выигранных игр в категориях по успешности не согласовано с общим количеством игр.")
	}

	// Проверяем статистику по дням.
	if stats.moreStrongTeamDays[1]+stats.moreStrongTeamDays[2]+stats.moreStrongTeamDays[3] != stats.moreStrongTeamDays[0] {
		errs = append(errs, "Суммарное количество дней по результатам в более сильных командах не согласовано с количеством дней")
	}

	if stats.lessStrongTeamDays[1]+stats.lessStrongTeamDays[2]+stats.lessStrongTeamDays[3] != stats.lessStrongTeamDays[0] {
		errs = append(errs, "Суммарное количество дней по результатам в менее сильных командах не согласовано с количеством дней")
	}

	if stats.equalStrongTeamDays[1]+stats.equalStrongTeamDays[2]+stats.equalStrongTeamDays[3] != stats.equalStrongTeamDays[0] {
		errs = append(errs, "Суммарное количество дней по результатам в равных по силе командах не согласовано с количеством дней")
	}

	if stats.moreSuccessfulTeamDays[1]+stats.moreSuccessfulTeamDays[2]+stats.moreSuccessfulTeamDays[3] != stats.moreSuccessfulTeamDays[0] {
		errs = append(errs, "Суммарное количество дней по результатам в более успешных командах не согласовано с количеством дней")
	}

	if stats.lessSuccessfulTeamDays[1]+stats.lessSuccessfulTeamDays[2]+stats.lessSuccessfulTeamDays[3] != stats.lessSuccessfulTeamDays[0] {
		errs = append(errs, "Суммарное количество дней по результатам в менее успешных командах не согласовано с количеством дней")
	}

	if stats.equalSuccessfulTeamDays[1]+stats.equalSuccessfulTeamDays[2]+stats.equalSuccessfulTeamDays[3] != stats.equalSuccessfulTeamDays[0] {
		errs = append(errs, "Суммарное количество дней по результатам в равных по успешности командах не согласовано с количеством дней")
	}

	if stats.moreStrongTeamDays[0]+stats.lessStrongTeamDays[0]+stats.equalStrongTeamDays[0] != stats.days*2 {
		errs = append(errs, "Суммарное количество дней в категориях по силе не согласовано с общим количеством дней.")
	}

	if stats.moreSuccessfulTeamDays[0]+stats.lessSuccessfulTeamDays[0]+stats.equalSuccessfulTeamDays[0] != stats.days*2 {
		errs = append(errs, "Суммарное количество дней в категориях по успешности не согласовано с общим количеством дней.")
	}

	/* Валидация статистик игроков. */
	var games, wins, days, winDays, defeatDays, drawDays int

	for _, player := range allPlayers {
		if player.games[0] != player.days[0]*GAMES_A_DAY {
			errs = append(errs, fmt.Sprintf("Количество игр игрока %s не согласовано с его количеством дней.", player.name))
		}
		games += player.games[0]
		wins += player.games[1]
		days += player.days[0]
		winDays += player.days[1]
		defeatDays += player.days[2]
		drawDays += player.days[3]
	}

	if games != stats.games*12 {
		errs = append(errs, "Суммарное количество игр игроков не согласовано с общим количеством игр.")
	}

	if days != stats.days*12 {
		errs = append(errs, "Суммарное количество игровых дней игроков не согласовано с общим количеством игровых дней.")
	}

	if wins*2 != games {
		errs = append(errs, "Суммарное количество побед игроков не согласовано с общим количеством игр.")
	}

	if winDays+defeatDays+drawDays != days {
		errs = append(errs, "Суммарное количество дней по итогам не согласовано с общим количеством дней.")
	}

	if len(errs) > 0 {
		fmt.Println("Статистика не согласована! Ошибки:")
		for _, err := range errs {
			fmt.Println(err)
		}
	}
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
	stats = &Stats{}
	// Выясняется после сбора команды
	moreStrongTeamPtr     *[]*Player
	lessStrongTeamPtr     *[]*Player
	moreSuccessfulTeamPtr *[]*Player
	lessSuccessfulTeamPtr *[]*Player
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

// play играет игру, записывает результаты и возвращает, победила ли первая команда.
func (tp *TeamPair) play() bool {
	// Шансы первой команды победить в отдельно взятой игре.
	threshold := tp.strengthSum1 / (tp.strengthSum1 + tp.strengthSum2)
	team1Won := rand.Float64() < threshold

	stats.updateGames(tp, team1Won)

	return team1Won
}

// playDay играет несколько игр и записывает результаты.
func (tp *TeamPair) playDay() {
	results := make(map[bool]int)

	for i := 0; i < GAMES_A_DAY; i++ {
		team1Won := tp.play()
		results[team1Won]++
	}

	stats.updateDays(tp, results)
}

// getMoreEqualTeamPair возвращает ту пару команд, где разница между суммарными рейтингами команд меньше.
func getMoreEqualTeamPair(tp1, tp2 TeamPair) TeamPair {
	if tp1.ratingSumDiff() <= tp2.ratingSumDiff() {
		return tp1
	} else {
		return tp2
	}
}

// findEqualTeamPair делит игроков на 2 максимально близкие по суммарной успешности команды.
func findEqualTeamPair(players *[12]*Player, tp TeamPair, i int) TeamPair {
	// Если размер какой-нибудь команды достиг половины players, оставшихся людей сунуть в другую группу и вернуть результат.
	if len(tp.team1) == len(players)/2 || len(tp.team2) == len(players)/2 {
		if len(tp.team1) == len(players)/2 {
			tp.team2 = append(tp.team2, players[i:]...)
			tp.successSum2 += getPlayersSuccessRateSum(players[i:])
			tp.strengthSum2 += getPlayersStrengthSum(players[i:])
		} else {
			tp.team1 = append(tp.team1, players[i:]...)
			tp.successSum1 += getPlayersSuccessRateSum(players[i:])
			tp.strengthSum1 += getPlayersStrengthSum(players[i:])
		}
		return tp
	}

	// Каждого игрока мы можем сунуть либо в одну либо в другую команду. Надо найти лучший вариант из двух.
	return getMoreEqualTeamPair(
		findEqualTeamPair(
			players,
			TeamPair{
				append(tp.team1, players[i]),
				tp.successSum1 + players[i].getSuccessRate(),
				tp.strengthSum1 + players[i].strength,
				append([]*Player{}, tp.team2...),
				tp.successSum2,
				tp.strengthSum2,
			},
			i+1,
		),
		findEqualTeamPair(
			players,
			TeamPair{
				append([]*Player{}, tp.team1...),
				tp.successSum1,
				tp.strengthSum1,
				append(tp.team2, players[i]),
				tp.successSum2 + players[i].getSuccessRate(),
				tp.strengthSum2 + players[i].strength,
			},
			i+1,
		),
	)
}

// pickTeamPair отбирает игроков, составляет из них 2 команды и возвращает указатель на неё.
func pickTeamPair() *TeamPair {
	// Отобрать 12 игроков
	var players [12]*Player
	indexes := rand.Perm(len(allPlayers))[:12]
	for i, index := range indexes {
		players[i] = allPlayers[index]
	}

	// Разделить игроков на 2 равные максимально близкие по суммарной успешности команды.
	teamPair := findEqualTeamPair(
		&players,
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

	// Выяснить более/менее сильные/успешные команды
	moreStrongTeamPtr = nil
	lessStrongTeamPtr = nil
	moreSuccessfulTeamPtr = nil
	lessSuccessfulTeamPtr = nil

	if teamPair.strengthSum1 != teamPair.strengthSum2 {
		if teamPair.strengthSum1 > teamPair.strengthSum2 {
			moreStrongTeamPtr = &teamPair.team1
			lessStrongTeamPtr = &teamPair.team2
		} else {
			moreStrongTeamPtr = &teamPair.team2
			lessStrongTeamPtr = &teamPair.team1
		}
	}
	if teamPair.successSum1 != teamPair.successSum2 {
		if teamPair.successSum1 > teamPair.successSum2 {
			moreSuccessfulTeamPtr = &teamPair.team1
			lessSuccessfulTeamPtr = &teamPair.team2
		} else {
			moreSuccessfulTeamPtr = &teamPair.team2
			lessSuccessfulTeamPtr = &teamPair.team1
		}
	}

	return &teamPair
}

func getCorrelation(set1, set2 []float64) (float64, error) {
	if len(set1) != len(set2) {
		return 0, fmt.Errorf("Set sizes are not equal: %d != %d", len(set1), len(set2))
	}

	var sum1, sum2, mean1, mean2, cov, s1, s2 float64

	for i := range set1 {
		sum1 += set1[i]
		sum2 += set2[i]
	}

	mean1 = sum1 / float64(len(set1))
	mean2 = sum2 / float64(len(set2))

	for i := range set1 {
		cov += (set1[i] - mean1) * (set2[i] - mean2)
		s1 += math.Pow((set1[i] - mean1), 2)
		s2 += math.Pow((set2[i] - mean2), 2)
	}

	return cov / math.Sqrt(s1*s2), nil
}

func main() {
	// Инициировать генератор случайных чисел.
	rand.Seed(42)

	// Сыграть много дней
	for d := 0; d < 1000; d++ {
		teamPair := pickTeamPair()
		teamPair.playDay()
	}

	fmt.Println(stats)
	stats.validate()

	strengths := make([]float64, len(allPlayers))
	successRates := make([]float64, len(allPlayers))
	for i, player := range allPlayers {
		strengths[i] = player.strength
		successRates[i] = player.getSuccessRate()
	}
	correlation, err := getCorrelation(strengths, successRates)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Коэффициент корреляции между силой игрока и его успешностью %.3f.\n", correlation)
}
