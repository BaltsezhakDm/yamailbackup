package storage

import (
	"database/sql"
	"fmt"
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
			now := time.Now()
			startOfDay := time.Date(now.Year(), now.Month(), now.Day()-1, 0, 0, 0, 0, now.Location())
			return startOfDay, nil
		}
		return time.Time{}, err // Возвращаем ошибку, если другая причина
	}
	parseTime, err := parseDate(lastDate)

	if err != nil {
		return time.Time{}, err
	}

	return parseTime, nil
}

func parseDate(lastDate string) (time.Time, error) {
	// Форматы для парсинга
	layouts := []string{
		"2006-01-02 15:04:05",             // YYYY-MM-DD HH:MM:SS
		"2006-01-02T15:04:05Z07:00",       // ISO 8601
		"02/01/2006 15:04:05",             // DD/MM/YYYY HH:MM:SS
		"2006-01-02 15:04:05 -0700 MST",   // 2025-02-27 03:51:09 +0000 UTC
		"2006-01-02 15:04:05 -0700 -0700", // 2025-02-27 03:51:09 +0000 +0000
	}
	var parseTime time.Time
	var err error

	// Пробуем каждый формат
	for _, layout := range layouts {
		parseTime, err = time.Parse(layout, lastDate)
		if err == nil {
			return parseTime, nil // Успех, возвращаем результат
		}
	}

	// Если ни один формат не подошел
	return time.Time{}, fmt.Errorf("invalid date format")
}
