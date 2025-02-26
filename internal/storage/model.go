package storage

import "time"

// Email представляет собой структуру для хранения писем в базе данных.
type Email struct {
	ID        int       `db:"id"`         // Уникальный идентификатор письма
	MessageID string    `db:"message_id"` // Уникальный Message-ID письма
	Subject   string    `db:"subject"`    // Тема письма
	From      string    `db:"from_email"` // Отправитель
	Date      time.Time `db:"date"`       // Дата отправки
}
