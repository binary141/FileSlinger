package server

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

const maxUploadSize = 10 << 30 // 10 GiB

func uploadHandler(dir string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
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
			if err := saveFile(dir, fh.Filename, func() (io.ReadCloser, error) {
				return fh.Open()
			}); err != nil {
				http.Error(w, fmt.Sprintf("failed to save %s: %v", fh.Filename, err), http.StatusInternalServerError)
				return
			}
			fmt.Printf("  saved: %s (%d bytes)\n", fh.Filename, fh.Size)
		}

		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprintf(w, "received %d file(s)\n", len(files))
	}
}

func saveFile(dir, filename string, open func() (io.ReadCloser, error)) error {
	// Sanitize: strip any directory component from the uploaded filename.
	safe := filepath.Base(filename)
	if safe == "." || safe == "/" {
		return fmt.Errorf("invalid filename")
	}

	src, err := open()
	if err != nil {
		return err
	}
	defer func() {
		_ = src.Close()
	}()

	dst, err := os.Create(filepath.Join(dir, safe))
	if err != nil {
		return err
	}
	defer func() {
		_ = dst.Close()
	}()

	_, err = io.Copy(dst, src)
	return err
}
