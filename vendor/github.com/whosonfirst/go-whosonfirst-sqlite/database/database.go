package database

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	_ "github.com/whosonfirst/go-spatialite"
	_ "log"
	"sync"
)

type SQLiteDatabase struct {
	conn *sql.DB
	dsn  string
	mu   *sync.Mutex
}

func NewDB(dsn string) (*SQLiteDatabase, error) {
	return NewDBWithDriver("sqlite3", dsn)
}

func NewDBWithDriver(driver string, dsn string) (*SQLiteDatabase, error) {

	conn, err := sql.Open(driver, dsn)

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
		// https://www.gaia-gis.it/gaia-sins/spatialite-cookbook/html/system.html
		"PRAGMA PAGE_SIZE=4096",
		"PRAGMA CACHE_SIZE=1000000",
	}

	for _, p := range pragma {

		_, err = conn.Exec(p)

		if err != nil {
			return err
		}
	}

	return nil
}

func (db *SQLiteDatabase) Lock() error {
	db.mu.Lock()
	return nil
}

func (db *SQLiteDatabase) Unlock() error {
	db.mu.Unlock()
	return nil
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
