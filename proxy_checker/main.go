package main

import (
	"bufio"
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
)

var (
	help           bool
	dbPath         string
	filePath       string
	stopAfterNHits int
	timeoutSecs    int
	targetURL      string
	minSpeedForHit int
	shuffle        bool

	errCreatingSOCKS5Dialer     = errors.New("failed to create SOCKS5 dialer")
	errDialerIsNotContextDialer = errors.New("proxy dialer does not implement proxy.ContextDialer")

	db     *sql.DB
	client *http.Client
)

type Proxy struct {
	address string
	success bool
	speed   int
}

func init() {
	initFlags()

	if dbPath != "" {
		initDB(dbPath)
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
	flag.StringVar(&filePath, "file", "", "path to the file with the list of proxy servers")
	flag.IntVar(&stopAfterNHits, "stop-after-n-hits", 0, "the number of hits to stop after. Hit is any success with speed not less than -min-speed-for-hit.")
	flag.IntVar(&timeoutSecs, "timeout", 24, "timeout for HTTP requests")
	flag.StringVar(&targetURL, "target", "https://www.youtube.com", "target URL to make requests to")
	flag.IntVar(&minSpeedForHit, "min-speed-for-hit", 0, "the minimum speed at which the success is a hit")
	flag.BoolVar(&shuffle, "shuffle", false, "shuffle proxies if source is a file")
	flag.Parse()

	if help {
		helpAndExit(nil)
	}
	if filePath == "" && dbPath == "" {
		helpAndExit(errors.New("neither --file nor --db are specified. It's not clear what to do."))
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
	}
	defer transport.CloseIdleConnections()

	client.Transport = transport

	start := time.Now()

	resp, err := client.Get(targetURL)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("status code %d", resp.StatusCode)
	}

	nBytes, err := io.Copy(io.Discard, resp.Body)
	if err != nil {
		return 0, err
	}
	elapsed := time.Since(start).Seconds()
	speed := int(float64(nBytes) / elapsed)

	return speed, nil
}

func checkProxies(addresses []string) ([]Proxy, error) {
	fmt.Printf("target url is %s\n\n", targetURL)

	var hitCount int
	var hitProxies []Proxy

	for i, address := range addresses {
		fmt.Printf("%d. %s:", i+1, address)

		speed, err := checkProxy(address)
		if err != nil {
			if errors.Is(err, errCreatingSOCKS5Dialer) || errors.Is(err, errDialerIsNotContextDialer) {
				return nil, err
			}
			fmt.Printf(" FAIL: %v\n", err)
			if db != nil {
				if err := updateInDB(address, false, 0); err != nil {
					log.Printf("ERROR: failed to update in database: %v", err)
				}
			}
			continue
		}

		fmt.Printf(" SUCCESS, Speed: %d b/s\n", speed)
		if db != nil {
			if err := updateInDB(address, true, speed); err != nil {
				log.Printf("ERROR: failed to update in database: %v", err)
			}
		}
		if speed >= minSpeedForHit {
			hitCount++
			hitProxies = append(hitProxies, Proxy{
				address: address,
				success: true,
				speed:   speed,
			})
		}
		if stopAfterNHits > 0 && hitCount >= stopAfterNHits {
			break
		}
	}

	slices.SortFunc(hitProxies, func(a, b Proxy) int {
		return b.speed - a.speed
	})

	return hitProxies, nil
}

func handleFile() ([]Proxy, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var addresses []string
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()
		addrPort, err := netip.ParseAddrPort(line)
		if err != nil || !addrPort.IsValid() {
			continue
		}
		addresses = append(addresses, addrPort.String())
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	file.Close()

	if shuffle {
		rand.Shuffle(len(addresses), func(i, j int) {
			addresses[i], addresses[j] = addresses[j], addresses[i]
		})
	}

	return checkProxies(addresses)
}

func handleDB() ([]Proxy, error) {
	query := `SELECT address FROM proxies WHERE points > 0 ORDER BY points DESC, speed DESC`

	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("query '%s' failed: %w", query, err)
	}
	defer rows.Close()

	var addresses []string

	for rows.Next() {
		var address string
		if err := rows.Scan(&address); err != nil {
			return nil, fmt.Errorf("query '%s' scan failed: %w", query, err)
		}
		addresses = append(addresses, address)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("query '%s' iteration failed: %w", query, err)
	}

	// Надо закрыть, так как checkProxies будет работать с той же таблицей.
	rows.Close()

	return checkProxies(addresses)
}

func main() {
	if db != nil {
		defer db.Close()
	}

	var err error
	var proxies []Proxy

	switch {
	case filePath != "":
		proxies, err = handleFile()
	case db != nil:
		proxies, err = handleDB()
	default:
		log.Fatal("unexpected case")
	}

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println()
	if len(proxies) > 0 {
		fmt.Println("Hit proxies:")
		for i, prx := range proxies {
			fmt.Printf("%d. %s\t%d b/s\n", i+1, prx.address, prx.speed)
		}
	} else {
		fmt.Println("There are no hit proxies")
	}
}
