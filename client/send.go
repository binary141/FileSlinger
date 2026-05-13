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

func SendFiles(uploadURL string, paths []string, excludeDirs []string) error {
	var count int
	for _, path := range paths {
		info, err := os.Stat(path)
		if err != nil {
			return err
		}
		if info.IsDir() {
			if err := sendDir(uploadURL, path, excludeDirs, &count); err != nil {
				return err
			}
		} else {
			count++
			if err := sendFile(uploadURL, path, filepath.Base(path), count); err != nil {
				return fmt.Errorf("%s: %w", filepath.Base(path), err)
			}
		}
	}
	return nil
}

func sendDir(uploadURL, dir string, excludeDirs []string, count *int) error {
	excluded := make(map[string]bool, len(excludeDirs))
	for _, d := range excludeDirs {
		excluded[d] = true
	}
	return filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if excluded[d.Name()] {
				return filepath.SkipDir
			}
			return nil
		}
		rel, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}
		*count++
		if err := sendFile(uploadURL, path, rel, *count); err != nil {
			return fmt.Errorf("%s: %w", rel, err)
		}
		return nil
	})
}

func sendFile(url, path, name string, n int) error {
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

	fmt.Printf("  #%d: %s (%d bytes)\n", n, name, info.Size())
	return nil
}
