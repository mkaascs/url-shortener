package mysql

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
)

type Storage struct {
	Database *sql.DB
}

func New(connectionString string) (*Storage, error) {
	const fn = "storage.mysql.New"

	db, err := sql.Open("mysql", connectionString)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", fn, err)
	}

	_, err = db.Exec(`
	CREATE TABLE IF NOT EXISTS url(
	    id INTEGER PRIMARY KEY,
	    alias VARCHAR(255) UNIQUE NOT NULL,
	    url TEXT NOT NULL);`)

	if err != nil {
		return nil, fmt.Errorf("%s: %w", fn, err)
	}

	var indexExists bool
	err = db.QueryRow(`
        SELECT COUNT(1) > 0
        FROM information_schema.statistics
        WHERE table_schema = DATABASE()
          AND table_name = 'url'
          AND index_name = 'idx_alias'
    `).Scan(&indexExists)

	if err != nil {
		indexExists = false
	}

	if !indexExists {
		_, err = db.Exec(`CREATE INDEX idx_alias ON url(alias)`)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", fn, err)
		}
	}

	return &Storage{db}, err
}
