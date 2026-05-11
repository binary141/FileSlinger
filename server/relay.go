package server

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"strings"

	"github.com/gorilla/websocket"
	"github.com/mdp/qrterminal/v3"
)

func StartRelay(cfg Config) error {
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

	relayBase := strings.TrimRight(cfg.RelayURL, "/")
	wsURL := toWebSocketURL(relayBase) + "/session/" + cfg.Token

	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		return fmt.Errorf("relay connect: %w", err)
	}

	defer func() {
		_ = conn.Close()
	}()

	uploadURL := relayBase + "/upload/" + cfg.Token

	qrterminal.GenerateHalfBlock(uploadURL, qrterminal.M, os.Stdout)

	fmt.Printf("Relay: %s\n", uploadURL)
	fmt.Printf("Dir:   %s\n", cfg.Dir)

	var received int
	for {
		_, data, err := conn.ReadMessage()
		if err != nil {
			return fmt.Errorf("relay disconnected: %w", err)
		}

		filename, content, err := unpackRelayMessage(data)
		if err != nil {
			fmt.Printf("  warning: malformed message: %v\n", err)
			continue
		}

		saved, err := saveFile(cfg.Dir, filename, func() (io.ReadCloser, error) {
			return io.NopCloser(bytes.NewReader(content)), nil
		})
		if err != nil {
			fmt.Printf("  error: %v\n", err)
			continue
		}
		fmt.Printf("  saved: %s (%d bytes)\n", saved, len(content))
		received++

		if cfg.MaxFiles > 0 && received >= cfg.MaxFiles {
			fmt.Println("File limit reached, shutting down.")
			return nil
		}
	}
}

func toWebSocketURL(u string) string {
	switch {
	case strings.HasPrefix(u, "https://"):
		return "wss://" + u[len("https://"):]
	case strings.HasPrefix(u, "http://"):
		return "ws://" + u[len("http://"):]
	default:
		return "ws://" + u
	}
}

// unpackRelayMessage parses a multipart body forwarded by the relay.
// The relay strips headers, so we extract the boundary from the first line.
func unpackRelayMessage(data []byte) (string, []byte, error) {
	nl := bytes.IndexByte(data, '\n')
	if nl < 3 || data[0] != '-' || data[1] != '-' {
		return "", nil, fmt.Errorf("not a multipart message")
	}
	boundary := strings.TrimRight(string(data[2:nl]), "\r")

	mr := multipart.NewReader(bytes.NewReader(data), boundary)
	part, err := mr.NextPart()
	if err != nil {
		return "", nil, fmt.Errorf("multipart: %w", err)
	}
	filename := part.FileName()
	if filename == "" {
		filename = "upload"
	}
	content, err := io.ReadAll(part)
	return filename, content, err
}
