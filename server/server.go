package server

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"net"
	"net/http"
	"os"
	"sync/atomic"

	"github.com/mdp/qrterminal/v3"
)

const tokenChars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
const tokenLen = 5

type Config struct {
	Port     int
	Dir      string
	MaxFiles int    // 0 = unlimited
	Token    string // auto-generated if empty
	RelayURL string // non-empty enables relay mode
}

func privateIP() string {
	ifaces, err := net.Interfaces()
	if err != nil {
		return "localhost"
	}
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip == nil || ip.IsLoopback() || ip.To4() == nil {
				continue
			}
			return ip.String()
		}
	}
	return "localhost"
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
	mux.HandleFunc("/ping", pingHandler)
	srv.Handler = logging(tokenAuth(cfg.Token, mux))

	limitMsg := "unlimited"
	if cfg.MaxFiles > 0 {
		limitMsg = fmt.Sprintf("limit %d", cfg.MaxFiles)
	}
	ip := privateIP()
	url := fmt.Sprintf("http://%s:%d/upload?token=%s", ip, cfg.Port, cfg.Token)

	qrterminal.GenerateHalfBlock(url, qrterminal.M, os.Stdout)

	fmt.Printf("URL:   %s\n", url)
	fmt.Printf("Dir:   %s (%s)\n", cfg.Dir, limitMsg)

	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		return err
	}
	return nil
}
