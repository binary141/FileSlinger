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
		info, err := os.Stat(path)
		if err != nil {
			return err
		}
		if info.IsDir() {
			if err := sendDir(uploadURL, path); err != nil {
				return err
			}
		} else {
			if err := sendFile(uploadURL, path, filepath.Base(path)); err != nil {
				return fmt.Errorf("%s: %w", filepath.Base(path), err)
			}
		}
	}
	return nil
}

func sendDir(uploadURL, dir string) error {
	return filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}
		if err := sendFile(uploadURL, path, rel); err != nil {
			return fmt.Errorf("%s: %w", rel, err)
		}
		return nil
	})
}

func sendFile(url, path, name string) error {
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

	if err := mw.WriteField("path", name); err != nil {
		return err
	}

	part, err := mw.CreateFormFile("file", filepath.Base(name))
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

	fmt.Printf("  sent: %s (%d bytes)\n", name, info.Size())
	return nil
}
