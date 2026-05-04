package server

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync/atomic"
)

const maxUploadSize = 10 << 30 // 10 GiB

func pingHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func uploadHandler(dir string, maxFiles int, received *atomic.Int32, shutdown func()) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		if maxFiles > 0 && int(received.Load()) >= maxFiles {
			http.Error(w, "file limit reached", http.StatusServiceUnavailable)
			return
		}

		r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)
		if err := r.ParseMultipartForm(32 << 20); err != nil {
			http.Error(w, "request too large or malformed", http.StatusBadRequest)
			return
		}

		files := r.MultipartForm.File["file"]
		if len(files) == 0 {
			http.Error(w, `multipart field "file" is required`, http.StatusBadRequest)
			return
		}

		for _, fh := range files {
			saved, err := saveFile(dir, fh.Filename, func() (io.ReadCloser, error) {
				return fh.Open()
			})
			if err != nil {
				http.Error(w, fmt.Sprintf("failed to save %s: %v", fh.Filename, err), http.StatusInternalServerError)
				return
			}
			fmt.Printf("  saved: %s (%d bytes)\n", filepath.Base(saved), fh.Size)
		}

		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprintf(w, "received %d file(s)\n", len(files))

		total := int(received.Add(int32(len(files))))
		if maxFiles > 0 && total >= maxFiles {
			fmt.Println("File limit reached, shutting down.")
			go shutdown()
		}
	}
}

func saveFile(dir, filename string, open func() (io.ReadCloser, error)) (string, error) {
	// Sanitize: strip any directory component from the uploaded filename.
	safe := filepath.Base(filename)
	if safe == "." || safe == "/" {
		return "", fmt.Errorf("invalid filename")
	}

	dest := deduplicateName(dir, safe)

	src, err := open()
	if err != nil {
		return "", err
	}
	defer func() {
		_ = src.Close()
	}()

	dst, err := os.Create(dest)
	if err != nil {
		return "", err
	}
	defer func() {
		_ = dst.Close()
	}()

	_, err = io.Copy(dst, src)
	return dest, err
}

// deduplicateName returns a path that does not yet exist, appending " (n)" before
// the extension when necessary — matching the behaviour of common file managers.
// e.g. "photo.jpg" → "photo (1).jpg" → "photo (2).jpg"
func deduplicateName(dir, filename string) string {
	path := filepath.Join(dir, filename)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return path
	}

	ext := filepath.Ext(filename)
	stem := filename[:len(filename)-len(ext)]

	for n := 1; ; n++ {
		candidate := filepath.Join(dir, fmt.Sprintf("%s (%d)%s", stem, n, ext))
		if _, err := os.Stat(candidate); os.IsNotExist(err) {
			return candidate
		}
	}
}
