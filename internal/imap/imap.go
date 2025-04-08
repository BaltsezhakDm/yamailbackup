package imap

import (
	"fmt"
	"log"
	"strings"
	"time"

	utils "github.com/KrasovD/yamailbackup/internal/utils"
	imap "github.com/emersion/go-imap"
	client "github.com/emersion/go-imap/client"
)

func ConnectToIMAPServer(server, port, username, password string) (*client.Client, error) {
	conn, err := client.DialTLS(server+":"+port, nil)
	if err != nil {
		return nil, err
	}

	if err := conn.Login(username, password); err != nil {
		return nil, err
	}
	return conn, nil
}

func ListMailboxes(conn *client.Client) ([]*imap.MailboxInfo, error) {
	mailboxes := make(chan *imap.MailboxInfo, 10)
	done := make(chan error, 1)

	go func() {
		done <- conn.List("", "*", mailboxes)
	}()

	var result []*imap.MailboxInfo
	for m := range mailboxes {
		result = append(result, m)
	}

	if err := <-done; err != nil {
		return nil, err
	}
	return result, nil
}

func FormatAddresses(addresses []*imap.Address) string {
	var formatted []string
	for _, addr := range addresses {
		formatted = append(formatted, fmt.Sprintf("%s <%s>", addr.PersonalName, addr.Address()))
	}
	return strings.Join(formatted, ", ")
}

// ListInboxHeaders загружает заголовки всех писем и фильтрует нужные.
func ListInboxHeaders(conn *client.Client, cfg *utils.Config, since time.Time) ([]*imap.Message, *imap.SeqSet, error) {
	const batchSize = 50
	since = since.Add(-20 * time.Minute)

	criteria := imap.NewSearchCriteria()
	criteria.Since = since

	// Выполняем поиск писем по критериям
	uids, err := conn.Search(criteria)
	if err != nil {
		return nil, nil, fmt.Errorf("search failed: %v", err)
	}
	if len(uids) == 0 {
		return nil, nil, nil // Нет писем
	}

	var result []*imap.Message
	filteredSeqset := new(imap.SeqSet)

	// Дробим на батчи
	for i := 0; i < len(uids); i += batchSize {
		end := i + batchSize
		if end > len(uids) {
			end = len(uids)
		}

		seqset := new(imap.SeqSet)
		seqset.AddNum(uids[i:end]...)

		// Загружаем только ENVELOPE
		messages := make(chan *imap.Message, batchSize)
		done := make(chan error, 1)

		go func() {
			done <- conn.Fetch(seqset, []imap.FetchItem{imap.FetchEnvelope}, messages)
		}()

		batch := []*imap.Message{}

		log.Println("Fetching headers..." + fmt.Sprint(i) + " - " + fmt.Sprint(end))
		// Читаем заголовки
		for msg := range messages {
			if msg.Envelope == nil {
				log.Println("Envelope is nil")
				continue
			}
			fromEmail := msg.Envelope.From[0].Address()
			if utils.ShouldProcessEmail(cfg, fromEmail) {
				batch = append(batch, msg)
				filteredSeqset.AddNum(msg.SeqNum)
			}
		}

		if err := <-done; err != nil {
			return nil, nil, err
		}

		result = append(result, batch...)
		log.Println("Headers processed:", len(result))
	}

	return result, filteredSeqset, nil
}

// FetchEmailBodies загружает тела только у отфильтрованных писем.
func FetchEmailBodies(conn *client.Client, filteredSeqset *imap.SeqSet) ([]*imap.Message, error) {
	if filteredSeqset == nil {
		return nil, nil // Нет подходящих писем
	}
	if filteredSeqset.Empty() {
		return nil, nil // Нет подходящих писем
	}

	const batchSize = 50
	var result []*imap.Message

	bodyMessages := make(chan *imap.Message, batchSize)
	bodyDone := make(chan error, 1)

	go func() {
		bodyDone <- conn.Fetch(filteredSeqset, []imap.FetchItem{
			imap.FetchEnvelope,
			imap.FetchItem("BODY.PEEK[]"),
		}, bodyMessages)
	}()

	bodyMap := make(map[uint32]*imap.Message)

	// Читаем тела писем
	for msg := range bodyMessages {
		if msg == nil {
			log.Println("Received nil message in bodyMessages")
			continue
		}
		bodyMap[msg.SeqNum] = msg
	}

	if err := <-bodyDone; err != nil {
		return nil, err
	}

	// Добавляем обновленные сообщения
	for _, msg := range bodyMap {
		result = append(result, msg)
	}

	log.Println("Bodies fetched:", len(result))
	return result, nil
}
