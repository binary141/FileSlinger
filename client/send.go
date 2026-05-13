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

func SendFiles(uploadURL string, paths []string) error {
	for _, path := range paths {
		if err := sendFile(uploadURL, path); err != nil {
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

	defer func() {
		err = f.Close()
	}()

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

	err = mw.Close()
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, url, &buf)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", mw.FormDataContentType())

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("server returned %s: %s", resp.Status, bytes.TrimSpace(body))
	}

	fmt.Printf("  sent: %s (%d bytes)\n", filepath.Base(path), info.Size())
	return nil
}
