package mysql

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/go-sql-driver/mysql"
	"url-shortener/internal/storage"
)

const (
	duplicateEntryErrCode = 1062
)

type Storage struct {
	Database *sql.DB
}

func (s *Storage) SaveURL(urlToSave string, alias string) (_ int64, err error) {
	const fn = "storage.mysql.SaveURL"

	stmt, err := s.Database.Prepare(`INSERT INTO url(url, alias) VALUES(?, ?)`)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", fn, err)
	}

	defer func(stmt *sql.Stmt) {
		if err != nil {
			_ = stmt.Close()
			return
		}

		err = stmt.Close()
	}(stmt)

	res, err := stmt.Exec(urlToSave, alias)
	if err != nil {
		var mysqlErr *mysql.MySQLError
		if errors.As(err, &mysqlErr) && mysqlErr.Number == duplicateEntryErrCode {
			return 0, fmt.Errorf("%s: %w", fn, storage.ErrURLExists)
		}

		return 0, fmt.Errorf("%s: %w", fn, err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("%s: failed to get last insert id: %w", fn, err)
	}

	return id, nil
}

func (s *Storage) GetURL(alias string) (_ string, err error) {
	const fn = "storage.mysql.GetURL"

	stmt, err := s.Database.Prepare(`SELECT url FROM url WHERE alias = ?`)
	if err != nil {
		return "", fmt.Errorf("%s: %w", fn, err)
	}

	defer func(stmt *sql.Stmt) {
		if err != nil {
			_ = stmt.Close()
			return
		}

		err = stmt.Close()
	}(stmt)

	var url string
	err = stmt.QueryRow(alias).Scan(&url)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", storage.ErrURLNotFound
		}

		return "", fmt.Errorf("%s: %w", fn, err)
	}

	return url, nil
}

func (s *Storage) DeleteURL(alias string) (err error) {
	const fn = "storage.mysql.DeleteURL"

	stmt, err := s.Database.Prepare(`DELETE FROM url WHERE alias = ?`)
	if err != nil {
		return fmt.Errorf("%s: %w", fn, err)
	}

	defer func(stmt *sql.Stmt) {
		if err != nil {
			_ = stmt.Close()
			return
		}

		err = stmt.Close()
	}(stmt)

	res, err := stmt.Exec(alias)
	if err != nil {
		return fmt.Errorf("%s: %w", fn, err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("%s: %w", fn, err)
	}

	if rowsAffected == 0 {
		return storage.ErrURLNotFound
	}

	return nil
}

func New(connectionString string) (*Storage, error) {
	const fn = "storage.mysql.New"

	db, err := sql.Open("mysql", connectionString)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", fn, err)
	}

	_, err = db.Exec(`
	CREATE TABLE IF NOT EXISTS url(
	    id INTEGER PRIMARY KEY AUTO_INCREMENT,
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
