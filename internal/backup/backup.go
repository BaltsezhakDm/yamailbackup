package backup

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/KrasovD/yamailbackup/internal/utils"
)

type UploadResponse struct {
	OperationID string `json:"operation_id"`
	Href        string `json:"href"`
	Method      string `json:"method"`
	Templated   bool   `json:"templated"`
}

func GetListCloud(cfg *utils.Config) error {
	var oauthURL = fmt.Sprintf(
		"%s%s",
		cfg.Backup.Host,
		"/v1/disk/resources?path=/",
	)
	client := &http.Client{}
	req, err := http.NewRequest(
		"GET", oauthURL, nil,
	)
	if err != nil {
		return err
	}
	req.Header.Add("Authorization", cfg.Backup.AuthKey)
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	io.Copy(os.Stdout, resp.Body)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	fmt.Println(string(body))
	// var result map[string]interface{}
	// if err := json.Unmarshal(body, &result); err != nil {
	// 	log.Println("Errror parsing json!")
	// 	return err
	// }

	// // Выводим красиво отформатированный JSON
	// formattedJSON, _ := json.MarshalIndent(result, "", "  ")
	// fmt.Println("Response JSON:", string(formattedJSON))
	return nil
}

func UploadToCloud(cfg *utils.Config, filePath string, fileBuffer *bytes.Buffer) error {

	if fileBuffer.Len() == 0 {
		return fmt.Errorf("file buffer is empty")
	}

	var oauthURL = fmt.Sprintf(
		"%s%s",
		cfg.Backup.Host,
		"/v1/disk/resources/upload/",
	)
	client := &http.Client{
		Timeout: 30 * time.Second, // Установить таймаут 30 секунд
	}

	req, err := http.NewRequest(
		"GET", oauthURL, nil,
	)
	if err != nil {
		return err
	}
	req.Header.Add("Authorization", cfg.Backup.AuthKey)
	values := req.URL.Query()
	values.Add("path", "/"+filePath)
	values.Add("overwrite", "true")
	req.URL.RawQuery = values.Encode()

	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	var data UploadResponse
	err = json.Unmarshal([]byte(body), &data)
	if err != nil {
		return err
	}

	req, err = http.NewRequest(
		data.Method, data.Href, fileBuffer,
	)
	if err != nil {
		return err
	}

	const maxRetries = 3
	for retries := 0; retries < maxRetries; retries++ {
		resp, err = client.Do(req)
		if err != nil {
			if retries == maxRetries-1 {
				return err // Последняя ошибка, возвращаем
			}
			time.Sleep(2 * time.Second) // Задержка перед повторной попыткой
			continue
		}

		if resp.StatusCode != 201 {
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				return fmt.Errorf("failed to read error body: %v", err)
			}
			return fmt.Errorf("failed to upload file: %s, response body: %s", resp.Status, string(body))
		}
		break

	}
	return nil
}
