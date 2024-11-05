package gui

import (
	"database/sql"
	"fmt"
	"log"
	"sync"
)

type Connection struct {
	mu sync.Mutex
	db *sql.DB
}

type Transaction struct {
	Timestamp string
	Code      string
	Quantity  int
	Price     int
	Side      string
}

const DB_DRIVER = "sqlite"
const GE_DATABASE = "ge.sql"
const PAGE_SIZE = 100

func (c *Connection) WithLock(f func(*sql.DB)) {
	c.mu.Lock()
	defer c.mu.Unlock()
	f(c.db)
}

func (c *Connection) Query(query string, args ...any) (*sql.Rows, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.db.Query(query, args...)
}

func (c *Connection) Exec(query string) (sql.Result, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.db.Exec(query)
}

func (c *Connection) Close() {
	c.db.Close()
}

func NewDBConnection() (*Connection, error) {
	fmt.Print("Connecting to database...")
	db, err := sql.Open(DB_DRIVER, GE_DATABASE)
	if err != nil {
		log.Fatalf("failed to open %s db: %s", DB_DRIVER, err)
		return nil, err
	}

	fmt.Printf("Connected to database %s\n", GE_DATABASE)
	return &Connection{db: db, mu: sync.Mutex{}}, nil
}
