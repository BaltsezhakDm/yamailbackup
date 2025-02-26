package storage

import (
	"database/sql"
	"time"

	_ "modernc.org/sqlite"
)

// InitDB инициализирует БД и создает таблицу, если её нет
func InitDB() (*sql.DB, error) {
	db, err := sql.Open("sqlite", "messages.db")
	if err != nil {
		return nil, err
	}

	query := `
	CREATE TABLE IF NOT EXISTS emails (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		message_id TEXT UNIQUE,
		subject TEXT,
		from_email TEXT,
		date TEXT
	);`

	_, err = db.Exec(query)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func SaveEmail(db *sql.DB, email Email) error {
	query := `INSERT INTO emails (message_id, subject, from_email, date)
	          VALUES (?, ?, ?, ?);`
	_, err := db.Exec(query, email.MessageID, email.Subject, email.From, email.Date)
	return err
}

func EmailExists(db *sql.DB, messageID string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM emails WHERE message_id = ? LIMIT 1)`
	err := db.QueryRow(query, messageID).Scan(&exists)
	return exists, err
}

func GetLastEmailID(db *sql.DB) (int, error) {
	var lastID int
	query := `SELECT COALESCE((SELECT id FROM emails ORDER BY id DESC LIMIT 1), 1)`

	err := db.QueryRow(query).Scan(&lastID)
	if err == sql.ErrNoRows {
		return 1, nil // Если записей нет, возвращаем 1
	} else if err != nil {
		return 0, err // Возвращаем ошибку, если что-то пошло не так
	}

	return lastID, nil
}

func GetLastEmailDate(db *sql.DB) (time.Time, error) {
	var lastDate string

	query := `SELECT date FROM emails ORDER BY date DESC LIMIT 1`
	err := db.QueryRow(query).Scan(&lastDate)

	if err != nil {
		if err == sql.ErrNoRows {
			// Возвращаем "нулевое" значение времени (unix timestamp 0)
			return time.Now(), nil
		}
		return time.Time{}, err // Возвращаем ошибку, если другая причина
	}
	layout := "2006-01-02 15:04:05 -0700 UTC"
	parseTime, err := time.Parse(layout, lastDate)

	if err != nil {
		return time.Time{}, err
	}

	return parseTime, nil
}
