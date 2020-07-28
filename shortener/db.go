package shortener

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	// Add mysql driver
	_ "github.com/go-sql-driver/mysql"
)

//InitDB inits mysql database
func InitDB(datasource string) *sql.DB {
	var db *sql.DB
	var err error
	db, err = sql.Open(datasource, os.Getenv("DB_USER")+":"+os.Getenv("DB_PASSWORD")+"@tcp(mysql:"+os.Getenv("DB_PORT")+")/"+os.Getenv("DATABASE_NAME"))
	if err != nil {
		log.Panic(err)
	}

	if err = db.Ping(); err != nil {
		log.Panic(err)
	}
	return db
}

// DatabaseStorage use database store url id
type DatabaseStorage struct {
	URLStorage
	db *sql.DB
}

// Open open database
func (d *DatabaseStorage) Open(storageType string, connectURL string) error {
	var db *sql.DB
	var err error
	switch storageType {
	case "mysql":
		db, err = sql.Open("mysql", connectURL)
	case "postgresql":
		db, err = sql.Open("postgres", connectURL)
	default:
		return fmt.Errorf("Not support database %v", storageType)
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
