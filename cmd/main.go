package main

import (
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/KrasovD/yamailbackup/internal/storage"
	"github.com/KrasovD/yamailbackup/internal/utils"

	imap "github.com/KrasovD/yamailbackup/internal/imap"
)

func main() {
	// Загрузка конфигов
	log.Println("Starting yamailbackup...")

	config, err := utils.LoadConfig("config/config.yaml")
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			config, err = utils.LoadConfig("/app/config/config.yaml")
		}
		if err != nil {
			log.Fatal("Error loading config:", err)
		}
	}

	// Вывод конфигурации
	fmt.Println("Backup interval:", config.Backup.Interval)
	fmt.Println("Save path:", config.Backup.SavePath)

	//Подключени к базе данных
	db, err := storage.InitDB()
	if err != nil {
		log.Fatal("Error initializing database:", err)
	}
	defer db.Close()

	// Проверка последнего MailID
	lastEmailDate, err := storage.GetLastEmailDate(db)
	if err != nil {
		log.Fatal("Error getting last email Date:", err)
	}

	// Подключение к IMAP
	conn, err := imap.ConnectToIMAPServer(
		config.IMAP.Server,
		fmt.Sprintf("%d", config.IMAP.Port),
		config.IMAP.Username,
		config.IMAP.Password,
	)
	if err != nil {
		log.Fatal("Error connecting to IMAP server:", err)
	}

	defer conn.Logout()
	// Получение списка почтовых ящиков

	// Выбираем "INBOX"
	mbox, err := conn.Select("INBOX", false)
	if err != nil {
		log.Fatal("Error selecting mailbox:", err)
	}
	counter := 0

	fmt.Println("Messages:", mbox.Messages)
	fmt.Println("Last email Date:", lastEmailDate)

	_, seqset, err := imap.ListInboxHeaders(conn, config, lastEmailDate)
	if err != nil {
		log.Fatal("Failed to get headers:", err)
	}

	bodies, err := imap.FetchEmailBodies(conn, seqset)
	if err != nil {
		log.Fatal("Failed to get bodies:", err)
	}

	for _, email := range bodies {
		// Проверяем, есть ли уже такое сообщение в базе данных
		exists, err := storage.EmailExists(db, email.Envelope.MessageId)
		if err != nil {
			log.Fatal("Error checking email existence:", err)
		}

		if !exists {
			// Скачиваем вложения
			err = imap.GetAttachments(email, config)
			if err != nil {
				log.Fatal("Error downloading attachments:", err)
			}

			//Сохраняем письмо в базе данных
			err := storage.SaveEmail(
				db,
				storage.Email{
					ID:        int(email.Uid),
					MessageID: email.Envelope.MessageId,
					Subject:   email.Envelope.Subject,
					From:      imap.FormatAddresses(email.Envelope.From),
					Date:      email.Envelope.Date,
				},
			)
			if err != nil {
				log.Fatal("Error saving email:", err)
			}

			counter++
		}

	}

}
