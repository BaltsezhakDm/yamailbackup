package imap

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"path/filepath"

	"github.com/KrasovD/yamailbackup/internal/backup"
	"github.com/KrasovD/yamailbackup/internal/utils"
	"github.com/emersion/go-imap"
	mail "github.com/emersion/go-message/mail"
)

// GetAttachments извлекает все вложения из письма и сохраняет их на диск
func GetAttachments(msg *imap.Message, cfg *utils.Config) error {

	for section := range msg.Body {
		// Получение тела секции сообщения
		r := msg.GetBody(section)
		if r == nil {
			continue
		}

		// Парсинг MIME-сообщения
		mr, err := mail.CreateReader(r)
		if err != nil {
			return fmt.Errorf("failed to create reader: %v", err)
		}

		// Чтение частей письма
		for {
			part, err := mr.NextPart()
			if err == io.EOF {
				break
			} else if err != nil {
				return fmt.Errorf("failed to read next part: %v", err)
			}

			switch h := part.Header.(type) {
			case *mail.InlineHeader:
				// Обработка встроенного содержимого
				// body, _ := io.ReadAll(part.Body)
				// fmt.Printf("Inline content: %s\n", body)

			case *mail.AttachmentHeader:
				// Обработка вложений
				filename, _ := h.Filename()
				if filename == "" {
					// Если у вложения нет имени, пропускаем его
					continue
				}

				// Полный путь для сохранения
				filePath := filepath.Join(cfg.Backup.SavePath, filename)

				fileBuffer := &bytes.Buffer{}

				_, err = io.Copy(fileBuffer, part.Body)
				if err != nil {
					return fmt.Errorf("failed to save attachment: %v", err)
				}

				err = backup.UploadToCloud(cfg, filePath, fileBuffer)
				if err != nil {
					return fmt.Errorf("failed to save attachment: %v", err)
				}
				log.Println("Download file:", filename, "Date:", msg.Envelope.Date)
				// Информация о сохранении
			}
		}
	}

	return nil
}

// getUniqueFilePath генерирует уникальное имя для файла, добавляя суффикс
// func getUniqueFilePath(filePath string, date time.Time) string {
// 	ext := filepath.Ext(filePath)
// 	base := filePath[:len(filePath)-len(ext)]
// 	newFilePath := fmt.Sprintf("%s (%s)%s", base, date.String(), ext)
// 	counter := 0

// 	// Проверка, существует ли файл с таким именем
// 	for {
// 		if _, err := os.Stat(newFilePath); os.IsNotExist(err) {
// 			break
// 		}
// 		counter++
// 		newFilePath = fmt.Sprintf("%s (%d)%s", base, counter, ext)
// 	}

// 	return newFilePath
// }
