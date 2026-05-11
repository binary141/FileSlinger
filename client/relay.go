package client

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func SendFilesRelay(relayURL, token string, paths []string) error {
	relayURL = strings.TrimRight(relayURL, "/")
	uploadURL := fmt.Sprintf("%s/upload/%s", relayURL, token)

	for _, path := range paths {
		if err := sendFileRelay(uploadURL, path); err != nil {
			return fmt.Errorf("%s: %w", filepath.Base(path), err)
		}
	}
	return nil
}

func sendFileRelay(uploadURL, path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer func() {
		_ = f.Close()
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

	resp, err := http.Post(uploadURL, mw.FormDataContentType(), &buf)
	if err != nil {
		return err
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("relay returned %s: %s", resp.Status, bytes.TrimSpace(body))
	}

	fmt.Printf("  sent: %s (%d bytes)\n", info.Name(), info.Size())
	return nil
}
