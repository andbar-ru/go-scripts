package distrowatch

import (
	"database/sql"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	_ "github.com/mattn/go-sqlite3"
)

const (
	BASE_URL    = "https://distrowatch.com/"
	DB_TABLE    = "distrs"
	TIME_LAYOUT = "20060102"
	USER_AGENT  = "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/67.0.3396.87 Safari/537.36 OPR/54.0.2952.46" // Opera 54
)

var (
	now          = time.Now()
	TODAY        = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	TODAY_YYMMDD = TODAY.Format(TIME_LAYOUT)
	DISTRS_DIR   = path.Join(os.Getenv("HOME"), "Images/distrs")
	DATABASE     = path.Join(DISTRS_DIR, "db.sqlite3")
	client       = &http.Client{}
)

type Distr struct {
	name string
	url  string
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
	request.Header.Add("User-Agent", USER_AGENT)
	response, err := client.Do(request)
	check(err)
	return response
}

func getDistr() Distr {
	// Get main page
	response := getResponse(BASE_URL)
	defer response.Body.Close()
	checkResponse(response)

	// Parse the page and fetch first distribution, hits per day of which didn't change since yesterday.
	root, err := goquery.NewDocumentFromReader(response.Body)
	check(err)

	hpdTds := root.Find("td.phr3") // HPD: Hits Per Day (Column header)
	if hpdTds.Length() == 0 {
		log.Fatal("There is no tds with class phr3.")
	} else if hpdTds.Length() != 100 {
		log.Printf("WARNING: number of tds with HPD is not 100, just %d", hpdTds.Length())
	}

	// Find first td which has img with alt="=" and fill Distr struct up.
	distr := Distr{}
	hpdTds.EachWithBreak(func(index int, hpdTd *goquery.Selection) bool {
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
			distributionTd := hpdTd.Prev()
			if !distributionTd.HasClass("phr2") {
				log.Fatalf("td.phr3 with index %d has previous sibling (distributionTd) with class name != 'phr2'.")
			}
			a := distributionTd.ChildrenFiltered("a").First()
			if a.Length() == 0 {
				log.Fatalf("td.phr3 with index %d has not an 'a' in previous sibling.", index)
			}
			distr.name = a.Text()
			url, ok := a.Attr("href")
			if !ok {
				log.Fatalf("a in td.phr2 with index %d has not attribute 'href'", index)
			}
			if !strings.HasPrefix(url, "http") {
				url = BASE_URL + url
			}
			distr.url = url
			return false // break EachWithBreak
		}
		return true
	})

	if distr.name == "" {
		log.Fatal("Could not find distribution with img.alt == '='.")
	}

	return distr
}

// Update or insert count of distribution distr.name in database.
func updateDb(db *sql.DB, distr Distr) {
	tx, err := db.Begin()
	check(err)
	_, err = tx.Exec("INSERT OR IGNORE INTO distrs (name, count, last_update) VALUES (?, 0, ?)", distr.name, TODAY_YYMMDD)
	check(err)
	_, err = tx.Exec("UPDATE distrs SET count = count + 1, last_update = ? WHERE name = ?", TODAY_YYMMDD, distr.name)
	check(err)
	err = tx.Commit()
	check(err)
}

func downloadScreenshot(distr Distr) string {
	// Get distr page
	response := getResponse(distr.url)
	defer response.Body.Close()
	checkResponse(response)

	// Parse the page and fetch full url of screenshot.
	root, err := goquery.NewDocumentFromReader(response.Body)
	check(err)

	a := root.Find("td.TablesTitle > a").First()
	if a.Length() == 0 {
		log.Fatalf("Could not find screenshot on page %s", distr.url)
	}
	url, ok := a.Attr("href")
	if !ok {
		log.Fatalf("Screenshot a has not attribute 'href' on page %s", distr.url)
	}
	if !strings.HasPrefix(url, "http") {
		url = BASE_URL + url
	}
	base := path.Base(url)
	screenshotPath := path.Join(DISTRS_DIR, base)

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
	_, err := os.Stat(DISTRS_DIR)
	if os.IsNotExist(err) {
		err = os.Mkdir(DISTRS_DIR, 0755)
		check(err)
	}

	// Check database existance
	if _, err := os.Stat(DATABASE); os.IsNotExist(err) {
		check(err)
	}

	// Open database
	db, err := sql.Open("sqlite3", DATABASE)
	check(err)
	defer db.Close()

	// If date of last_update is today, exit.
	var lastUpdate string
	err = db.QueryRow("SELECT last_update FROM distrs ORDER BY last_update DESC LIMIT 1").Scan(&lastUpdate)
	check(err)
	if lastUpdate == TODAY_YYMMDD {
		fmt.Println("Database is already updated today.")
		os.Exit(0)
	}

	distr := getDistr()
	updateDb(db, distr)
	_ = downloadScreenshot(distr)
}
