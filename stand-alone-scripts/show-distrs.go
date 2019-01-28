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
	DATABASE = path.Join(os.Getenv("HOME"), "Images/distrs/db.sqlite3")
)

type distr struct {
	name        string
	count       int
	last_update int
}

func main() {
	// Check if database exists
	if _, err := os.Stat(DATABASE); os.IsNotExist(err) {
		panic(err)
	}

	// Open database
	db, err := sql.Open("sqlite3", DATABASE)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	// Query
	query := "SELECT name, count, last_update FROM distrs ORDER BY count DESC, last_update ASC"
	rows, err := db.Query(query)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	// Collect distrs
	var nameFieldSize, countFieldSize, newestLastUpdate, oldestLastUpdate int
	oldestLastUpdate = 1e8 // certainly greater than any date
	distrs := []distr{}
	for rows.Next() {
		var d distr
		err := rows.Scan(&d.name, &d.count, &d.last_update)
		if err != nil {
			panic(err)
		}
		distrs = append(distrs, d)

		// Update name field size
		if len(d.name) > nameFieldSize {
			nameFieldSize = len(d.name)
		}
		// Update newest and oldest last_updates
		if d.last_update > newestLastUpdate {
			newestLastUpdate = d.last_update
		}
		if d.last_update < oldestLastUpdate {
			oldestLastUpdate = d.last_update
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
		fmt.Printf("%*d. %-*s %*d %d", numberFieldSize, i+1, nameFieldSize, distr.name, countFieldSize, distr.count, distr.last_update)
		if distr.last_update == newestLastUpdate {
			fmt.Print(" newest")
		}
		if distr.last_update == oldestLastUpdate {
			fmt.Print(" oldest")
		}
		fmt.Print("\x1b[0m")
		fmt.Println()
	}
}
