package main

import (
	"database/sql"
	"fmt"
	"math"
	"os"
	"path"

	_ "github.com/mattn/go-sqlite3"
)

var (
	database = path.Join(os.Getenv("HOME"), "Images/distrs/db.sqlite3")
)

type distr struct {
	name       string
	count      int
	lastUpdate int
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	// Check if database exists
	if _, err := os.Stat(database); os.IsNotExist(err) {
		panic(err)
	}

	// Open database
	db, err := sql.Open("sqlite3", database)
	check(err)
	defer db.Close()

	// Print coordinates.
	var latitude, longitude float64
	err = db.QueryRow("SELECT latitude, longitude FROM stats ORDER BY date DESC LIMIT 1").Scan(&latitude, &longitude)
	check(err)
	fmt.Printf("Coordinates: %.4f %.4f\n\n", latitude, longitude)

	// Query
	query := "SELECT name, count, last_update FROM distrs ORDER BY count DESC, last_update ASC"
	rows, err := db.Query(query)
	check(err)
	defer rows.Close()

	// Collect distrs
	var nameFieldSize, countFieldSize, newestLastUpdate, oldestLastUpdate int
	oldestLastUpdate = 1e8 // certainly greater than any date
	distrs := []distr{}
	for rows.Next() {
		var d distr
		err := rows.Scan(&d.name, &d.count, &d.lastUpdate)
		check(err)
		distrs = append(distrs, d)

		// Update name field size
		if len(d.name) > nameFieldSize {
			nameFieldSize = len(d.name)
		}
		// Update newest and oldest last_updates
		if d.lastUpdate > newestLastUpdate {
			newestLastUpdate = d.lastUpdate
		}
		if d.lastUpdate < oldestLastUpdate {
			oldestLastUpdate = d.lastUpdate
		}
	}
	countFieldSize = len(fmt.Sprint(distrs[0].count)) // number of digits

	if err := rows.Err(); err != nil {
		panic(err)
	}

	// Difference between adjucent distrs to define leaders
	diff := int(math.Ceil(float64(distrs[0].count) / 10))

	// Define leaders from the end
	lastLeaderIndex := -1 // no leaders by default
	for i := len(distrs) - 1; i >= 0; i-- {
		if distrs[i-1].count-distrs[i].count >= diff {
			lastLeaderIndex = i - 1
			break
		}
	}

	// Print table
	numberFieldSize := len(fmt.Sprint(len(distrs)))
	for i, distr := range distrs {
		boldness := 0
		color := 37
		if i <= lastLeaderIndex {
			boldness = 1
		}
		fmt.Printf("\x1b[%d;%dm", boldness, color)
		fmt.Printf("%*d. %-*s %*d %d", numberFieldSize, i+1, nameFieldSize, distr.name, countFieldSize, distr.count, distr.lastUpdate)
		if distr.lastUpdate == newestLastUpdate {
			fmt.Print(" newest")
		}
		if distr.lastUpdate == oldestLastUpdate {
			fmt.Print(" oldest")
		}
		fmt.Print("\x1b[0m")
		fmt.Println()
	}
}
