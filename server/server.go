package server

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"net/http"
	"os"
	"sync/atomic"
)

const tokenChars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
const tokenLen = 5

type Config struct {
	Port     int
	Dir      string
	MaxFiles int    // 0 = unlimited
	Token    string // auto-generated if empty
}

func generateToken() (string, error) {
	b := make([]byte, tokenLen)
	for i := range b {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(tokenChars))))
		if err != nil {
			return "", err
		}
		b[i] = tokenChars[n.Int64()]
	}
	return string(b), nil
}

func Start(cfg Config) error {
	if err := os.MkdirAll(cfg.Dir, 0755); err != nil {
		return fmt.Errorf("could not create directory %s: %w", cfg.Dir, err)
	}

	if cfg.Token == "" {
		tok, err := generateToken()
		if err != nil {
			return fmt.Errorf("could not generate token: %w", err)
		}
		cfg.Token = tok
	}

	var received atomic.Int32

	srv := &http.Server{Addr: fmt.Sprintf(":%d", cfg.Port)}

	shutdown := func() {
		_ = srv.Shutdown(context.Background())
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/upload", uploadHandler(cfg.Dir, cfg.MaxFiles, &received, shutdown))
	srv.Handler = logging(tokenAuth(cfg.Token, mux))

	limitMsg := "unlimited"
	if cfg.MaxFiles > 0 {
		limitMsg = fmt.Sprintf("limit %d", cfg.MaxFiles)
	}
	fmt.Printf("Listening on http://localhost:%d/upload, saving files to %s (%s)\n", cfg.Port, cfg.Dir, limitMsg)
	fmt.Printf("Token: %s\n", cfg.Token)

	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		return err
	}
	return nil
}
