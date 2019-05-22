package main

import (
	"database/sql"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	_ "github.com/mattn/go-sqlite3"
)

const (
	baseURL    = "https://distrowatch.com/"
	timeLayout = "20060102"
	userAgent  = "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/67.0.3396.87 Safari/537.36 OPR/54.0.2952.46" // Opera 54
)

var (
	now         = time.Now()
	today       = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	todayYYMMDD = today.Format(timeLayout)
	distrsDir   = path.Join(os.Getenv("HOME"), "Images/distrs")
	database    = path.Join(distrsDir, "db.sqlite3")
	client      = &http.Client{}
	distrCount  = 100
)

// Outcome stores outcome from baseURL.
type Outcome struct {
	distrName string
	distrURL  string
	mean      float64
	stdev     float64
}

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func checkResponse(response *http.Response) {
	if response.StatusCode != 200 {
		log.Fatalf("Status code error: %d %s", response.StatusCode, response.Status)
	}
}

func getResponse(url string) *http.Response {
	request, err := http.NewRequest("GET", url, nil)
	check(err)
	request.Header.Add("User-Agent", userAgent)
	response, err := client.Do(request)
	check(err)
	return response
}

func getMean(values []int) float64 {
	var sum int
	for _, value := range values {
		sum += value
	}
	return float64(sum) / float64(len(values))
}

func getStdev(values []int, mean float64) float64 {
	var sqrDiffSum float64
	for _, value := range values {
		val := float64(value)
		sqrDiffSum += (val - mean) * (val - mean)
	}
	return math.Sqrt(sqrDiffSum / float64(len(values)-1))
}

func getOutcome() Outcome {
	// Get main page
	response := getResponse(baseURL)
	defer response.Body.Close()
	checkResponse(response)

	// Parse the page and fetch first distribution, hits per day of which didn't change since yesterday.
	root, err := goquery.NewDocumentFromReader(response.Body)
	check(err)

	hpdTds := root.Find("td.phr3") // HPD: Hits Per Day (Column header)
	if hpdTds.Length() == 0 {
		log.Fatal("There is no tds with class phr3.")
	} else if hpdTds.Length() != distrCount {
		log.Printf("WARNING: number of tds with HPD is not %d, just %d", distrCount, hpdTds.Length())
		distrCount = hpdTds.Length()
	}

	// Find first td which has img with alt="=" and fill Outcome struct up.
	outcome := Outcome{}
	hpds := make([]int, 0, distrCount)
	var equalFound bool
	hpdTds.Each(func(index int, hpdTd *goquery.Selection) {
		hpd, err := strconv.Atoi(hpdTd.Text())
		check(err)
		hpds = append(hpds, hpd)

		if equalFound {
			return
		}

		img := hpdTd.ChildrenFiltered("img").First()
		// Every hpdTd must contain just one img.
		if img.Length() == 0 {
			log.Fatalf("td.phr3 with index %d has not an img.", index)
		}
		// Image must have the attribute 'alt'.
		alt, ok := img.Attr("alt")
		if !ok {
			log.Fatalf("img in td.phr3 with index %d has not attribute 'alt'", index)
		}
		if alt == "=" {
			equalFound = true

			distributionTd := hpdTd.Prev()
			if !distributionTd.HasClass("phr2") {
				log.Fatalf("td.phr3 with index %d has previous sibling (distributionTd) with class name != 'phr2'.")
			}
			a := distributionTd.ChildrenFiltered("a").First()
			if a.Length() == 0 {
				log.Fatalf("td.phr3 with index %d has not an 'a' in previous sibling.", index)
			}
			outcome.distrName = a.Text()
			url, ok := a.Attr("href")
			if !ok {
				log.Fatalf("a in td.phr2 with index %d has not attribute 'href'", index)
			}
			if !strings.HasPrefix(url, "http") {
				url = baseURL + url
			}
			outcome.distrURL = url
		}
	})

	if outcome.distrName == "" {
		log.Fatal("Could not find distribution with img.alt == '='.")
	}
	outcome.mean = getMean(hpds)
	outcome.stdev = getStdev(hpds, outcome.mean)

	return outcome
}

// Update or insert count of distribution name in database.
func updateDb(db *sql.DB, outcome Outcome) {
	tx, err := db.Begin()
	check(err)
	_, err = tx.Exec("INSERT OR IGNORE INTO distrs (name, count, last_update) VALUES (?, 0, ?)", outcome.distrName, todayYYMMDD)
	check(err)
	_, err = tx.Exec("UPDATE distrs SET count = count + 1, last_update = ? WHERE name = ?", todayYYMMDD, outcome.distrName)
	check(err)

	latitude := math.Mod(outcome.stdev, 180)
	if latitude > 90 {
		latitude = latitude - 180
	}
	longitude := math.Mod(outcome.mean, 360)
	if longitude > 180.0 {
		longitude = longitude - 360.0
	}
	_, err = tx.Exec("INSERT INTO stats (date, mean, stdev, latitude, longitude) VALUES (?, ?, ?, ?, ?)", todayYYMMDD, fmt.Sprintf("%.4f", outcome.mean), fmt.Sprintf("%.4f", outcome.stdev), fmt.Sprintf("%.4f", latitude), fmt.Sprintf("%.4f", longitude))
	check(err)

	err = tx.Commit()
	check(err)
}

func downloadScreenshot(distrURL string) string {
	// Get distr page
	response := getResponse(distrURL)
	defer response.Body.Close()
	checkResponse(response)

	// Parse the page and fetch full url of screenshot.
	root, err := goquery.NewDocumentFromReader(response.Body)
	check(err)

	a := root.Find("td.TablesTitle > a").First()
	if a.Length() == 0 {
		log.Fatalf("Could not find screenshot on page %s", distrURL)
	}
	url, ok := a.Attr("href")
	if !ok {
		log.Fatalf("Screenshot a has not attribute 'href' on page %s", distrURL)
	}
	if !strings.HasPrefix(url, "http") {
		url = baseURL + url
	}
	base := path.Base(url)
	screenshotPath := path.Join(distrsDir, base)

	// Download screenshot
	output, err := os.Create(screenshotPath)
	if err != nil {
		log.Fatalf("Could not create file %s, err: %s", screenshotPath, err)
	}
	defer output.Close()

	response = getResponse(url)
	defer response.Body.Close()
	checkResponse(response)

	_, err = io.Copy(output, response.Body)
	if err != nil {
		log.Fatalf("Could not write image %s to file %s, err: %s", url, screenshotPath, err)
	}

	return screenshotPath
}

func main() {
	// Create directories if they are not exist
	_, err := os.Stat(distrsDir)
	if os.IsNotExist(err) {
		err = os.Mkdir(distrsDir, 0755)
		check(err)
	}

	// Check database existance
	if _, err := os.Stat(database); os.IsNotExist(err) {
		check(err)
	}

	// Open database
	db, err := sql.Open("sqlite3", database)
	check(err)
	defer db.Close()

	// If date of last_update is today, exit.
	var lastUpdate string
	err = db.QueryRow("SELECT last_update FROM distrs ORDER BY last_update DESC LIMIT 1").Scan(&lastUpdate)
	check(err)
	if lastUpdate == todayYYMMDD {
		fmt.Println("Database is already updated today.")
		os.Exit(0)
	}

	outcome := getOutcome()
	updateDb(db, outcome)
	_ = downloadScreenshot(outcome.distrURL)
}
