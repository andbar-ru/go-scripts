package main

import (
	"bufio"
	"cmp"
	"context"
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"math/rand/v2"
	"net"
	"net/http"
	"net/netip"
	"os"
	"slices"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/net/proxy"
)

const (
	successPoints = 12
	userAgent     = "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/151.0.0.0 Safari/537.36" // Brave 1.93.134
)

// Флаги
var (
	help           bool
	dbPath         string
	filePath       string
	onlyDB         bool
	onlyFile       bool
	shuffleFile    bool
	timeoutSecs    int
	targetURL      string
	stopAfterNHits int
	minSpeedForHit int
)

// Ошибки
var (
	errCreatingSOCKS5Dialer     = errors.New("failed to create SOCKS5 dialer")
	errDialerIsNotContextDialer = errors.New("proxy dialer does not implement proxy.ContextDialer")
)

// Переменные, инициируемые позже.
var (
	db     *sql.DB
	file   *os.File
	client *http.Client
)

type Hit struct {
	address string
	speed   int
}

func init() {
	initFlags()

	if dbPath != "" {
		initDB(dbPath)
	}
	if filePath != "" {
		initFile(filePath)
	}

	client = &http.Client{
		Timeout: time.Duration(timeoutSecs) * time.Second,
	}
}

func helpAndExit(err error) {
	exitCode := 0
	if err != nil {
		fmt.Printf("Error: %v\n\n", err)
		exitCode = 1
	}
	fmt.Println("Usage: go run main.go [OPTION]...")
	fmt.Println("Available flags:")
	flag.PrintDefaults()
	os.Exit(exitCode)
}

func initFlags() {
	flag.BoolVar(&help, "help", false, "print help")
	flag.StringVar(&dbPath, "db", "", "path to the database")
	flag.StringVar(&filePath, "file", "", "path to the file containing a list of proxy servers, one ip:port per line")
	flag.BoolVar(&onlyDB, "only-db", false, "use only the database as the source. Mutually exclusive with -only-file.")
	flag.BoolVar(&onlyFile, "only-file", false, "use only the file as the source. Mutually exclusive with -only-db.")
	flag.BoolVar(&shuffleFile, "shuffle-file", false, "shuffle the proxies from the file source")
	flag.IntVar(&timeoutSecs, "timeout", 24, "timeout for HTTP requests, in seconds")
	flag.StringVar(&targetURL, "target", "https://www.youtube.com", "URL to request through each proxy")
	flag.IntVar(&stopAfterNHits, "stop-after-n-hits", 0, "number of hits to stop after. 0 means no limit. A hit is any success with a speed not less than -min-speed-for-hit.")
	flag.IntVar(&minSpeedForHit, "min-speed-for-hit", 0, "minimum speed, in bytes per second, at which the success counts as a hit")
	flag.Parse()

	if help {
		helpAndExit(nil)
	}
	if filePath == "" && dbPath == "" {
		helpAndExit(errors.New("neither -file nor -db is specified. At least one source is required."))
	}
	if onlyDB && dbPath == "" {
		helpAndExit(errors.New("-only-db requires -db to be specified"))
	}
	if onlyFile && filePath == "" {
		helpAndExit(errors.New("-only-file requires -file to be specified"))
	}
	if shuffleFile && filePath == "" {
		helpAndExit(errors.New("-shuffle-file requires -file to be specified"))
	}
	if onlyDB && onlyFile {
		helpAndExit(errors.New("-only-db and -only-file are mutually exclusive"))
	}
}

func initDB(dbPath string) {
	if dbPath == "" {
		log.Fatal("Empty dbPath")
	}

	var err error
	db, err = sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatalf("Failed to open database %s: %v", dbPath, err)
	}

	query := `CREATE TABLE IF NOT EXISTS proxies (
		address TEXT PRIMARY KEY,
		points INTEGER NOT NULL,
		speed INTEGER NOT NULL,
		updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	)`
	if _, err := db.Exec(query); err != nil {
		log.Fatalf("Query '%s' failed: %v", query, err)
	}
}

func initFile(filePath string) {
	if filePath == "" {
		log.Fatal("Empty filePath")
	}

	var err error
	file, err = os.Open(filePath)
	if err != nil {
		log.Fatalf("Failed to open file %s: %v", filePath, err)
	}
}

func getProxiesFromDB() ([]string, error) {
	query := `SELECT address FROM proxies WHERE points > 0 ORDER BY points DESC, speed DESC`

	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("query '%s' failed: %w", query, err)
	}
	defer rows.Close()

	var proxies []string

	for rows.Next() {
		var proxy string
		if err := rows.Scan(&proxy); err != nil {
			return nil, fmt.Errorf("query '%s' scan failed: %w", query, err)
		}
		proxies = append(proxies, proxy)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("query '%s' iteration failed: %w", query, err)
	}

	return proxies, nil
}

func getProxiesFromFile() ([]string, error) {
	var proxies []string
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()
		addrPort, err := netip.ParseAddrPort(line)
		if err != nil || !addrPort.IsValid() {
			continue
		}
		proxies = append(proxies, addrPort.String())
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	if shuffleFile {
		rand.Shuffle(len(proxies), func(i, j int) {
			proxies[i], proxies[j] = proxies[j], proxies[i]
		})
	}

	return proxies, nil
}

func updateInDB(address string, success bool, speed int) error {
	var query string
	var params []any
	if success {
		query = `INSERT INTO proxies (address, points, speed) VALUES (?, ?, ?)
		ON CONFLICT(address) DO UPDATE SET
			points = excluded.points,
			speed = excluded.speed,
			updated_at = CURRENT_TIMESTAMP`
		params = []any{address, successPoints, speed}
	} else {
		// Не обновляю записи о неудачных прокси, если их нет в базе. Вариант с points = 0 тоже
		// считается отсутствующим при сканировании. Это сделано намеренно, чтобы не держать в базе
		// прокси, которые никогда не были рабочими.
		query = `UPDATE proxies SET
			points = points - 1,
			updated_at = CURRENT_TIMESTAMP
		WHERE address = ? AND points > 0`
		params = []any{address}
	}

	_, err := db.Exec(query, params...)
	return err
}

func checkProxy(address string) (int, error) {
	dialer, err := proxy.SOCKS5("tcp", address, nil, proxy.Direct)
	if err != nil {
		return 0, fmt.Errorf("%w: %w", errCreatingSOCKS5Dialer, err)
	}

	contextDialer, ok := dialer.(proxy.ContextDialer)
	if !ok {
		return 0, errDialerIsNotContextDialer
	}

	transport := &http.Transport{
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			return contextDialer.DialContext(ctx, network, addr)
		},
		ForceAttemptHTTP2: true,
	}
	defer transport.CloseIdleConnections()

	// Менять транспорт у глобального клиента допустимо, только если прокси будут проверяться
	// последовательно. Если надумаю реализовать конкурентность, то понадобится делать собственный
	// клиент для каждого прокси.
	client.Transport = transport

	start := time.Now()

	request, err := http.NewRequest("GET", targetURL, nil)
	if err != nil {
		return 0, err
	}
	request.Header.Add("User-Agent", userAgent)
	response, err := client.Do(request)
	if err != nil {
		return 0, err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("status code %d", response.StatusCode)
	}

	nBytes, err := io.Copy(io.Discard, response.Body)
	if err != nil {
		return 0, err
	}
	elapsed := time.Since(start).Seconds()
	speed := int(float64(nBytes) / elapsed)

	return speed, nil
}

func checkProxies(proxies []string) ([]Hit, error) {
	var hits []Hit
	numProxies := len(proxies)

	for i, proxy := range proxies {
		fmt.Printf("%d/%d. %s:", i+1, numProxies, proxy)

		speed, err := checkProxy(proxy)
		if err != nil {
			if errors.Is(err, errCreatingSOCKS5Dialer) || errors.Is(err, errDialerIsNotContextDialer) {
				return nil, err
			}
			fmt.Printf(" FAIL: %v\n", err)
			if db != nil {
				if err := updateInDB(proxy, false, 0); err != nil {
					log.Printf("ERROR: failed to update in database: %v", err)
				}
			}
			continue
		}

		fmt.Printf(" SUCCESS, Speed: %d b/s\n", speed)
		if db != nil {
			if err := updateInDB(proxy, true, speed); err != nil {
				log.Printf("ERROR: failed to update in database: %v", err)
			}
		}
		if speed >= minSpeedForHit {
			hits = append(hits, Hit{
				address: proxy,
				speed:   speed,
			})
		}
		if stopAfterNHits > 0 && len(hits) >= stopAfterNHits {
			break
		}
	}

	slices.SortFunc(hits, func(a, b Hit) int {
		return cmp.Compare(b.speed, a.speed)
	})

	return hits, nil
}

func main() {
	if db != nil {
		defer db.Close()
	}
	if file != nil {
		defer file.Close()
	}

	var proxies []string
	uniqueProxies := make(map[string]bool)

	if db != nil && !onlyFile {
		dbProxies, err := getProxiesFromDB()
		if err != nil {
			log.Fatalf("Failed to get proxies from database: %v", err)
		}
		for _, proxy := range dbProxies {
			// В базе прокси гарантированно не повторяются, поэтому не проверяем.
			uniqueProxies[proxy] = true
			proxies = append(proxies, proxy)
		}
	}
	if file != nil && !onlyDB {
		fileProxies, err := getProxiesFromFile()
		if err != nil {
			log.Fatalf("Failed to get proxies from file: %v", err)
		}
		// В файле могут быть прокси, которые уже есть в базе, поэтому проверяем.
		for _, proxy := range fileProxies {
			if !uniqueProxies[proxy] {
				uniqueProxies[proxy] = true
				proxies = append(proxies, proxy)
			}
		}
	}

	fmt.Printf("Target URL: %s\n\n", targetURL)

	hits, err := checkProxies(proxies)
	if err != nil {
		log.Fatalf("Failed to check proxies: %v", err)
	}

	fmt.Println()

	if len(hits) > 0 {
		fmt.Println("Hit proxies:")
		for i, hit := range hits {
			fmt.Printf("%d. %s\t%d b/s\n", i+1, hit.address, hit.speed)
		}
	} else {
		fmt.Println("There are no hits :-(")
	}
}
