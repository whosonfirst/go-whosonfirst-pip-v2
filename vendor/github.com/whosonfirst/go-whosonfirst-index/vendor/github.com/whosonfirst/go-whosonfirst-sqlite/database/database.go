package database

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	_ "log"
	"sync"
)

type SQLiteDatabase struct {
	conn *sql.DB
	dsn  string
	mu   *sync.Mutex
}

func NewDB(dsn string) (*SQLiteDatabase, error) {

	conn, err := sql.Open("sqlite3", dsn)

	if err != nil {
		return nil, err
	}

	mu := new(sync.Mutex)

	db := SQLiteDatabase{
		conn: conn,
		dsn:  dsn,
		mu:   mu,
	}

	return &db, err
}

// https://blog.devart.com/increasing-sqlite-performance.html
// https://www.sqlite.org/pragma.html#pragma_journal_mode

func (db *SQLiteDatabase) LiveHardDieFast() error {

	conn, err := db.Conn()

	if err != nil {
		return err
	}

	pragma := []string{
		"PRAGMA JOURNAL_MODE=OFF",
		"PRAGMA SYNCHRONOUS=OFF",
		"PRAGMA LOCKING_MODE=EXCLUSIVE",
	}

	for _, p := range pragma {

		_, err = conn.Exec(p)

		if err != nil {
			return err
		}
	}

	return nil
}

func (db *SQLiteDatabase) Lock() {
	db.mu.Lock()
}

func (db *SQLiteDatabase) Unlock() {
	db.mu.Unlock()
}

func (db *SQLiteDatabase) Conn() (*sql.DB, error) {
	return db.conn, nil
}

func (db *SQLiteDatabase) DSN() string {
	return db.dsn
}

func (db *SQLiteDatabase) Close() error {
	return db.conn.Close()
}
