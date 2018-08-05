/*
Script downloads today's wallpaper from bingwallpaper.com, sets wallpaper and shows message with
wallpaper description. Information about downloaded wallpapers is saved into WP_FILE. If today's
wallpaper has been downloaded already, script does nothing. If there are missed dates, script
downloads wallpapers at that dates. WP_FILE's lines have the following format:
YYYYMMDD <wallpaper-file-name> <description>.
*/
package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

const (
	BASE_URL       = "https://bingwallpaper.com"
	RESOLUTION     = "1920x1080"
	LAYOUT         = "20060102"
	COPYRIGHT_TEXT = "- Bing™ Wallpaper Gallery"
)

var (
	IMG_DIR   = fmt.Sprintf("%s/Images/bing-wallpapers", os.Getenv("HOME"))
	WP_FILE   = fmt.Sprintf("%s/wallpapers", IMG_DIR)
	now       = time.Now()
	TODAY     = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	YESTERDAY = TODAY.AddDate(0, 0, -1)
	lastDate  time.Time
)

func check(err error) {
	if err != nil {
		log.Panic(err)
	}
}

// Download wallpaper at the specified date.
func downloadWallpaper(date time.Time) (string, string) {
	var filename, title string

	// Fetch the page with photo
	url := fmt.Sprintf("%s/%s.html", BASE_URL, date.Format(LAYOUT))
	res, err := http.Get(url)
	check(err)
	defer res.Body.Close()
	if res.StatusCode != 200 {
		log.Panicf("status code error: %d %s", res.StatusCode, res.Status)
	}

	// Parse the page and fetch the image metadata
	root, err := goquery.NewDocumentFromReader(res.Body)
	check(err)
	imgContainers := root.Find(".imgContainer")
	if imgContainers.Length() == 0 {
		photos := root.Find("#photos")
		if photos.Text() != "\n' title='Previous' hidefocus='true' target='_self' class='nph_btn_pphoto' id='photoPrev'>" {
			log.Panicf("photos have unexpected text: %s", photos.Text())
		}
		title = "There is no wallpaper on this date"
		return filename, title
	}

	imgContainer := imgContainers.First()
	img := imgContainer.Find("img").First()
	titleTag := root.Find("title").First()
	if titleTag.Length() == 0 {
		log.Panicf("Page on url %s hasn't title tag", url)
	}
	title = titleTag.Text()
	copyrightIndex := strings.Index(title, COPYRIGHT_TEXT)
	if copyrightIndex != -1 {
		title = strings.Trim(title[:copyrightIndex], " ")
	}
	src, ok := img.Attr("src")
	if !ok {
		log.Panic("img has not attribute 'src'")
	}
	src = strings.TrimLeft(src, "/")
	if !strings.HasPrefix(src, "https://") {
		src = "https://" + src
	}
	var re = regexp.MustCompile(`\d+x\d+(\.\w+$)`)
	src = re.ReplaceAllString(src, RESOLUTION+"$1") // change resolution
	lastSlashIndex := strings.LastIndex(src, "/")
	filename = src[lastSlashIndex+1:]
	filepath := fmt.Sprintf("%s/%s", IMG_DIR, filename)
	fmt.Println(filepath)

	// Download image
	output, err := os.Create(filepath)
	if err != nil {
		log.Panicf("Could not create file %s, err: %s", filepath, err)
	}
	defer output.Close()
	response, err := http.Get(src)
	if err != nil {
		log.Panicf("Could not download image from %s, err: %s", src, err)
	}
	defer response.Body.Close()
	_, err = io.Copy(output, response.Body)
	if err != nil {
		log.Panicf("Could not write image to file, err: %s", err)
	}

	return filename, title
}

// Set wallpaper and show message with title.
func setWallpaper(filename, title string) {
	filepath := fmt.Sprintf("%s/%s", IMG_DIR, filename)

	setWallpaperCmd := exec.Command("fbsetbg", "-f", filepath)
	err := setWallpaperCmd.Start()
	check(err)

	msgCmd := exec.Command("xmessage", "-buttons", "OK", "-title", "New wallpaper", "-center", title)
	err = msgCmd.Start()
	check(err)
}

// Save record about wallpaper into file.
func logWallpaper(date time.Time, filename, title string) {
	// Escape single quotes for sed.
	title = strings.Replace(title, "'", `\x27`, -1)
	line := fmt.Sprintf("%s %s %s\\n", date.Format(LAYOUT), filename, title)
	sedCmd := exec.Command("sed", "-i", fmt.Sprintf("1s;^;%s;", line), WP_FILE)
	err := sedCmd.Start()
	check(err)
}

func main() {
	// Create directory if not exists
	_, err := os.Stat(IMG_DIR)
	if os.IsNotExist(err) {
		err = os.Mkdir(IMG_DIR, 0755)
		check(err)
	}

	// If date of last downloaded image is today, exit
	_, err = os.Stat(WP_FILE)
	if os.IsNotExist(err) {
		f, err := os.Create(WP_FILE)
		check(err)
		_, err = f.WriteString("\n")
		check(err)
		f.Close()
	} else {
		f, err := os.Open(WP_FILE)
		check(err)
		lastDateBytes := make([]byte, 8) // YYYYMMDD
		_, err = f.Read(lastDateBytes)
		check(err)
		lastDate, err = time.Parse(LAYOUT, string(lastDateBytes))
		check(err)
		if lastDate == TODAY {
			os.Exit(0)
		}
	}

	var filenameYesterday, titleYesterday string

	// Download wallpapers at dates before today if not yet downloaded.
	if !lastDate.IsZero() && lastDate.Before(YESTERDAY) {
		curDate := lastDate.AddDate(0, 0, 1)
		var filename, title string
		for !curDate.After(YESTERDAY) {
			filename, title = downloadWallpaper(curDate)
			logWallpaper(curDate, filename, title)
			curDate = curDate.AddDate(0, 0, 1)
		}
		filenameYesterday = filename
		titleYesterday = title
	}

	// Download wallpaper at today
	filename, title := downloadWallpaper(TODAY)
	if filename != "" {
		setWallpaper(filename, title)
		logWallpaper(TODAY, filename, title)
	} else if filenameYesterday != "" {
		setWallpaper(filenameYesterday, titleYesterday)
	}
}
