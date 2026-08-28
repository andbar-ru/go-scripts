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
	"net"
	"net/http"
	"net/netip"
	"os"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/net/proxy"
)

const (
	successPoints = 12
)

var (
	help          bool
	dbPath        string
	filePath      string
	stopOnSuccess bool
	timeoutSecs   int
	targetURL     string

	errCreatingSOCKS5Dialer     = errors.New("failed to create SOCKS5 dialer")
	errDialerIsNotContextDialer = errors.New("proxy dialer does not implement proxy.ContextDialer")

	db     *sql.DB
	client *http.Client
)

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
	flag.BoolVar(&stopOnSuccess, "stop-on-success", true, "stop on first success")
	flag.IntVar(&timeoutSecs, "timeout", 24, "timeout for HTTP requests")
	flag.StringVar(&targetURL, "target", "https://www.youtube.com", "target URL to make requests to")
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

func checkProxies(addresses []string) error {
	fmt.Printf("target url is %s\n\n", targetURL)

	for i, address := range addresses {
		fmt.Printf("%d. %s:", i+1, address)

		speed, err := checkProxy(address)
		if err != nil {
			if errors.Is(err, errCreatingSOCKS5Dialer) || errors.Is(err, errDialerIsNotContextDialer) {
				return err
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
		if stopOnSuccess {
			break
		}
	}

	return nil
}

func handleFile() error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
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
		return err
	}

	file.Close()

	return checkProxies(addresses)
}

func handleDB() error {
	query := `SELECT address FROM proxies WHERE points > 0 ORDER BY points DESC, speed DESC`

	rows, err := db.Query(query)
	if err != nil {
		return fmt.Errorf("query '%s' failed: %w", query, err)
	}
	defer rows.Close()

	var addresses []string

	for rows.Next() {
		var address string
		if err := rows.Scan(&address); err != nil {
			return fmt.Errorf("query '%s' scan failed: %w", query, err)
		}
		addresses = append(addresses, address)
	}

	if err = rows.Err(); err != nil {
		return fmt.Errorf("query '%s' iteration failed: %w", query, err)
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

	switch {
	case filePath != "":
		err = handleFile()
	case db != nil:
		err = handleDB()
	default:
		log.Fatal("unexpected case")
	}

	if err != nil {
		log.Fatal(err)
	}
}
