package urlshortener

import (
	"database/sql"
	"fmt"

	// This loads the postgres drivers.
	_ "github.com/lib/pq"

	"github.com/friends-of-scalability/url-shortener/pkg"
)

// shortURLPostgresRepository is an in-memory user database.
type shortURLPostgresRepository struct {
	db *sql.DB
}

// New returns a postgres backed storage service.
func newPostgresStorage(host, port, user, password, dbName string) (shortURLStorage, error) {
	// Connect postgres
	connect := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbName)
	db, err := sql.Open("postgres", connect)
	if err != nil {
		return nil, err
	}

	// Ping to connection
	err = db.Ping()
	if err != nil {
		return nil, err
	}

	// Create table if not exists
	strQuery := "CREATE TABLE IF NOT EXISTS shortener (uid BIGSERIAL NOT NULL, url VARCHAR not NULL UNIQUE, " +
		"count BIGINT DEFAULT 0);"

	_, err = db.Exec(strQuery)
	if err != nil {
		return nil, err
	}
	return &shortURLPostgresRepository{db}, nil
}

func (u *shortURLPostgresRepository) IsHealthy() (bool, error) {
	// Send a ping to make sure the database connection is alive.
	if _, err := u.db.Exec("SELECT 1"); err != nil {
		return false, err
	}

	return true, nil
}

// ByShortURL finds and URL in our databse.
func (u *shortURLPostgresRepository) ByURL(URL string) (*shortURL, error) {

	var item shortURL
	err := u.db.QueryRow("SELECT uid,url,count FROM shortener where url=$1 limit 1", URL).Scan(
		&item.ID,
		&item.URL,
		&item.VisitsCounter)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errURLNotFound
		}
		return nil, err
	}
	return &item, nil
}

func (u *shortURLPostgresRepository) ByID(id string) (*shortURL, error) {

	dbID, err := base62.Decode(id)
	if err != nil {
		return nil, err
	}

	var item shortURL
	err = u.db.QueryRow("SELECT url, count FROM shortener where uid=$1 limit 1", dbID).
		Scan(&item.URL, &item.VisitsCounter)
	item.ID = dbID

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errURLNotFound
		}
		return nil, err
	}
	if err != nil {
		return nil, err
	}

	return &item, nil
}

func (u *shortURLPostgresRepository) getURLAndUpdateVisitsCounter(item *shortURL) (*shortURL, error) {
	m, err := u.ByURL(item.URL)
	if m != nil {
		m.VisitsCounter++
		_, err = u.db.Exec("UPDATE shortener SET count=$2 WHERE uid = $1;", m.ID, m.VisitsCounter)
		if err != nil {
			return nil, err
		}
		return m, nil
	} else if err != errURLNotFound {
		return nil, err
	}
	return nil, nil
}
func (u *shortURLPostgresRepository) createNewItem(item *shortURL) (*shortURL, error) {
	var id uint64
	err := u.db.QueryRow("INSERT INTO shortener(url,count) VALUES($1,$2) returning uid;", item.URL, 0).Scan(&id)
	if err != nil {
		return nil, err
	}
	var mapping shortURL
	mapping.URL = item.URL
	mapping.VisitsCounter = 0
	mapping.ID = id
	return &mapping, nil
}

// ByShortURL finds and URL in our databse.
func (u *shortURLPostgresRepository) Save(item *shortURL) (*shortURL, error) {

	retrievedItem, err := u.getURLAndUpdateVisitsCounter(item)
	if retrievedItem != nil {
	} else if err != nil {
		return nil, err
	}

	return u.createNewItem(item)
}
