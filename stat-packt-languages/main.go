package main

import (
	"fmt"
	"net/http"
	"slices"
	"strconv"

	"github.com/PuerkitoBio/goquery"
)

const (
	baseUrl   = "https://github.com/orgs/PacktPublishing/repositories?type=all"
	userAgent = "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/119.0.0.0 Safari/537.36"
)

var (
	client         = &http.Client{}
	languageCounts = make(map[string]int)
	curPage        = 0
)

type LangCount struct {
	language string
	count    int
}

func setHeaders(request *http.Request) {
	request.Header.Set("authority", "github.com")
	request.Header.Set("accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8")
	request.Header.Set("accept-language", "en-US,en;q=0.6")
	request.Header.Set("cache-control", "no-cache")
	request.Header.Set("cookie", "Insert value from browser session")
	request.Header.Add("pragma", "no-cache")
	request.Header.Set("sec-ch-ua", `"Brave";v="119", "Chromium";v="119", "Not?A_Brand";v="24"`)
	request.Header.Set("sec-ch-ua-mobile", "?0")
	request.Header.Set("sec-ch-ua-platform", `"Linux"`)
	request.Header.Set("sec-fetch-dest", "document")
	request.Header.Set("sec-fetch-mode", "navigate")
	request.Header.Set("sec-fetch-site", "none")
	request.Header.Set("sec-fetch-user", "?1")
	request.Header.Set("sec-gpc", "1")
	request.Header.Set("upgrade-insecure-requests", "1")
	request.Header.Set("user-agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/119.0.0.0 Safari/537.36")
}

func handleUrl(url string) string {
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		panic(err)
	}
	setHeaders(request)
	response, err := client.Do(request)
	if err != nil {
		panic(err)
	}
	defer response.Body.Close()

	for response.StatusCode != 200 {
		panic(fmt.Sprintf("%s: status code %d (%s)", url, response.StatusCode, response.Status))
	}

	root, err := goquery.NewDocumentFromReader(response.Body)
	if err != nil {
		panic(err)
	}
	rows := root.Find(".Box-row")
	if rows.Length() == 0 {
		panic(fmt.Sprintf("%s: no elements of class 'Box-row'", url))
	}
	rows.Each(func(i int, row *goquery.Selection) {
		elem := row.Find(`[itemprop="programmingLanguage"]`)
		if elem.Length() == 0 {
			languageCounts[""]++
			return
		}
		lang := elem.Text()
		languageCounts[lang]++
	})

	aNextPage := root.Find(".next_page")
	if aNextPage.Length() == 0 {
		panic(fmt.Sprintf("%s: no elements of class 'next_page'", url))
	}
	aNextPage = aNextPage.First()
	if aNextPage.HasClass("disabled") {
		// That's the last page
		return ""
	}
	href, _ := aNextPage.Attr("href")
	if href == "" {
		panic(fmt.Sprintf("%s: element of class 'next_page' has no href attribute", url))
	}

	url = "https://github.com" + href
	return url
}

func main() {
	nextUrl := baseUrl
	for nextUrl != "" {
		curPage++
		if curPage%10 == 0 {
			fmt.Print(curPage)
		} else {
			fmt.Print(".")
		}
		nextUrl = handleUrl(nextUrl)
	}
	fmt.Println()
	fmt.Println()

	langCounts := make([]LangCount, 0, len(languageCounts))
	maxLangLength := 0
	total := 0
	for lang, count := range languageCounts {
		if lang == "" {
			lang = "no language"
		}
		langCounts = append(langCounts, LangCount{lang, count})
		if len(lang) > maxLangLength {
			maxLangLength = len(lang)
		}
		total += count
	}
	slices.SortFunc(langCounts, func(a, b LangCount) int {
		return b.count - a.count // reverse
	})
	totalLength := len(strconv.Itoa(total))

	for _, lc := range langCounts {
		fmt.Printf("%-*s %*d\n", maxLangLength+1, lc.language, totalLength, lc.count)
	}
	fmt.Println()
	fmt.Printf("%-*s %*d\n", maxLangLength+1, "Total", totalLength, total)
}
