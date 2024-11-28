package main

import (
	"log"
	"path/filepath"
	"sync"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

type URLMapping struct {
	ID        string `db:"id"`
	LongURL   string `db:"long_url"`
	ExpiresAt int64  `db:"expires_at"`
}

type URLDatabase struct {
	db *sqlx.DB
	mu sync.Mutex
}

func NewURLDatabase() *URLDatabase {
	dbFile := filepath.Join(".", "db.sqlite")
	db, err := sqlx.Open("sqlite3", dbFile)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}

	schema := `CREATE TABLE IF NOT EXISTS url_mapping (
		id TEXT PRIMARY KEY,
		long_url TEXT NOT NULL,
		expires_at INTEGER NOT NULL
	)`
	if _, err := db.Exec(schema); err != nil {
		log.Fatalf("Failed to create table: %v", err)
	}

	return &URLDatabase{db: db}
}

func (ud *URLDatabase) Close() {
	ud.db.Close()
}

func (ud *URLDatabase) GetURLByLongURL(longURL string) (*URLMapping, error) {
	ud.mu.Lock()
	defer ud.mu.Unlock()

	var mapping URLMapping
	err := ud.db.Get(&mapping, "SELECT * FROM url_mapping WHERE long_url = ?", longURL)
	if err != nil {
		return nil, err
	}
	return &mapping, nil
}

func (ud *URLDatabase) GetURLByShortID(shortID string) (*URLMapping, error) {
	ud.mu.Lock()
	defer ud.mu.Unlock()

	var mapping URLMapping
	err := ud.db.Get(&mapping, "SELECT * FROM url_mapping WHERE id = ?", shortID)
	if err != nil {
		return nil, err
	}
	return &mapping, nil
}

func (ud *URLDatabase) CreateShortURL(shortID, longURL string, expiresAt int64) error {
	ud.mu.Lock()
	defer ud.mu.Unlock()

	_, err := ud.db.Exec("INSERT INTO url_mapping (id, long_url, expires_at) VALUES (?, ?, ?)", shortID, longURL, expiresAt)
	return err
}

func (ud *URLDatabase) ExtendURLExpiration(shortID string, expiresAt int64) error {
	ud.mu.Lock()
	defer ud.mu.Unlock()

	_, err := ud.db.Exec("UPDATE url_mapping SET expires_at = ? WHERE id = ?", expiresAt, shortID)
	return err
}

func (ud *URLDatabase) CleanExpiredRecords() {
	for {
		time.Sleep(12 * time.Hour) // Clean every 12 hours
		ud.mu.Lock()
		now := time.Now().Unix()
		_, err := ud.db.Exec("DELETE FROM url_mapping WHERE expires_at <= ?", now)
		if err != nil {
			log.Printf("Failed to clean expired records: %v", err)
		}
		ud.mu.Unlock()
	}
}
