package server

import (
	"fmt"
	"net/http"
)

type Config struct {
	Port int
	Dir  string
}

func Start(cfg Config) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/upload", uploadHandler(cfg.Dir))

	addr := fmt.Sprintf(":%d", cfg.Port)
	fmt.Printf("Listening on http://localhost:%d/upload, saving files to %s\n", cfg.Port, cfg.Dir)
	return http.ListenAndServe(addr, logging(mux))
}
