package shortener

import (
	"database/sql"
	"fmt"
	"net/url"

	// Add mysql driver
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
)

// DatabaseStorage  database storage
type DatabaseStorage struct {
	URLStorage
	db *sql.DB
}

// Open open database
func (d *DatabaseStorage) Open(url *url.URL) error {
	var db *sql.DB
	var err error

	switch url.Scheme {
	case "mysql":
		db, err = sql.Open("mysql", url.String())
	case "postgresql":
		db, err = sql.Open("postgres", url.String())
	default:
		return fmt.Errorf("Not support database %v", url.Scheme)
	}
	if err != nil {
		return err
	}
	if err = db.Ping(); err != nil {
		return err
	}
	d.db = db
	return nil
}

// Close close database connection
func (d *DatabaseStorage) Close() error {
	if d.db != nil {
		return d.db.Close()
	}
	return nil
}

// NewURL create new url entry
func (d *DatabaseStorage) NewURL(url string, id string, ttl int) error {
	var err error
	_, err = d.db.Exec(`INSERT INTO shortened_urls (id, long_url, created) VALUES(?, ?, now())`,
		id, url)
	if err != nil {
		return err
	}
	return nil
}

// FindByID find long url by id
func (d *DatabaseStorage) FindByID(id string) (string, error) {
	var longURL string
	err := d.db.QueryRow("SELECT id FROM shortened_urls WHERE id = ?", id).Scan(&longURL)
	if err != nil && sql.ErrNoRows != err {
		return "", err
	}
	if sql.ErrNoRows == err {
		return "", ErrIDNotFound
	}
	return longURL, nil
}

// DeleteURLByID delete url by id
func (d *DatabaseStorage) DeleteURLByID(id string) error {
	_, err := d.db.Exec(`DELETE FROM shortened_urls WHERE id = ?`, id)
	if err != nil && sql.ErrNoRows != err {
		return err
	}
	return nil
}
