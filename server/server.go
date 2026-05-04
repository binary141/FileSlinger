package server

import (
	"fmt"
	"net/http"
	"os"
)

type Config struct {
	Port int
	Dir  string
}

func Start(cfg Config) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/upload", uploadHandler(cfg.Dir))

	addr := fmt.Sprintf(":%d", cfg.Port)
	if err := os.MkdirAll(cfg.Dir, 0755); err != nil {
		return fmt.Errorf("could not create directory %s: %w", cfg.Dir, err)
	}
	fmt.Printf("Listening on http://localhost:%d/upload, saving files to %s\n", cfg.Port, cfg.Dir)
	return http.ListenAndServe(addr, logging(mux))
}
