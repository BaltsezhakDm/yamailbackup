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
		Timeout: 30 * time.Second,
	}

	req, err := http.NewRequest("GET", oauthURL, nil)
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

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to get upload URL: %s, body: %s", resp.Status, string(body))
	}

	var data UploadResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return fmt.Errorf("failed to decode upload response: %v", err)
	}

	fileBytes := fileBuffer.Bytes()
	const maxRetries = 3
	var lastErr error

	for retries := 0; retries < maxRetries; retries++ {
		// Re-create request for each retry because the body might have been closed/drained
		uploadReq, err := http.NewRequest(data.Method, data.Href, bytes.NewReader(fileBytes))
		if err != nil {
			return fmt.Errorf("failed to create upload request: %v", err)
		}

		uploadResp, err := client.Do(uploadReq)
		if err != nil {
			lastErr = err
			time.Sleep(2 * time.Second)
			continue
		}

		if uploadResp.StatusCode != http.StatusCreated && uploadResp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(uploadResp.Body)
			uploadResp.Body.Close()
			lastErr = fmt.Errorf("failed to upload file: %s, response body: %s", uploadResp.Status, string(body))
			time.Sleep(2 * time.Second)
			continue
		}
		uploadResp.Body.Close()
		return nil
	}

	return fmt.Errorf("upload failed after %d retries: %v", maxRetries, lastErr)
}
