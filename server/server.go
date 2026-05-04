package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"sync/atomic"
)

type Config struct {
	Port     int
	Dir      string
	MaxFiles int // 0 = unlimited
}

func Start(cfg Config) error {
	if err := os.MkdirAll(cfg.Dir, 0755); err != nil {
		return fmt.Errorf("could not create directory %s: %w", cfg.Dir, err)
	}

	var received atomic.Int32

	srv := &http.Server{Addr: fmt.Sprintf(":%d", cfg.Port)}

	shutdown := func() {
		_ = srv.Shutdown(context.Background())
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/upload", uploadHandler(cfg.Dir, cfg.MaxFiles, &received, shutdown))
	srv.Handler = logging(mux)

	limitMsg := "unlimited"
	if cfg.MaxFiles > 0 {
		limitMsg = fmt.Sprintf("limit %d", cfg.MaxFiles)
	}
	fmt.Printf("Listening on http://localhost:%d/upload, saving files to %s (%s)\n", cfg.Port, cfg.Dir, limitMsg)

	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		return err
	}
	return nil
}
