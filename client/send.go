package client

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
)

func SendFiles(host string, port int, paths []string) error {
	url := fmt.Sprintf("http://%s:%d/upload", host, port)

	for _, path := range paths {
		if err := sendFile(url, path); err != nil {
			return fmt.Errorf("%s: %w", filepath.Base(path), err)
		}
	}
	return nil
}

func sendFile(url, path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		return err
	}

	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)

	part, err := mw.CreateFormFile("file", filepath.Base(path))
	if err != nil {
		return err
	}

	if _, err = io.Copy(part, f); err != nil {
		return err
	}
	mw.Close()

	resp, err := http.Post(url, mw.FormDataContentType(), &buf)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("server returned %s: %s", resp.Status, bytes.TrimSpace(body))
	}

	fmt.Printf("  sent: %s (%d bytes)\n", filepath.Base(path), info.Size())
	return nil
}
