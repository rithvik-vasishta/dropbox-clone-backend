package aws

import (
	"bytes"
	"fmt"
	"mime/multipart"
	"net/http"
	"time"
)

const (
	RequestTimeout = 5 * time.Second
)

func UploadShardToNode(nodeURL, shardFilename string, data []byte) error {
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	part, err := writer.CreateFormFile("file", shardFilename)
	if err != nil {
		return err
	}
	if _, err = part.Write(data); err != nil {
		return err
	}
	writer.Close()

	req, err := http.NewRequest("POST", nodeURL+"/upload-shard", &buf)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := http.Client{Timeout: RequestTimeout}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("upload shard failed, status %d", resp.StatusCode)
	}
	return nil
}

func DownloadShardFromNode(nodeURL, shardPath string) (*http.Response, error) {
	client := http.Client{Timeout: RequestTimeout}
	resp, err := client.Get(nodeURL + shardPath)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("download shard failed, status %d", resp.StatusCode)
	}
	return resp, nil
}
