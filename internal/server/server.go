package server

import (
	"context"
	"embed"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"keepassview/internal/config"
	"keepassview/internal/vault"
)

// Server holds shared state for all HTTP handlers.
type Server struct {
	vault        *vault.Vault
	cfg          *config.Config
	assets       embed.FS
	lastActivity atomic.Int64
	shutdownOnce sync.Once
	shutdownCh   chan struct{}
	httpServer   *http.Server
}

func (s *Server) touchActivity() {
	s.lastActivity.Store(time.Now().UnixNano())
}

func (s *Server) triggerShutdown() {
	s.shutdownOnce.Do(func() {
		close(s.shutdownCh)
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = s.httpServer.Shutdown(ctx)
	})
}

// Run starts the HTTP server and blocks until it shuts down.
func Run(v *vault.Vault, cfg *config.Config, port int, assets embed.FS) error {
	ln, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
	if err != nil {
		return fmt.Errorf("listen: %w", err)
	}

	s := &Server{
		vault:      v,
		cfg:        cfg,
		assets:     assets,
		shutdownCh: make(chan struct{}),
	}
	s.touchActivity()

	mux := http.NewServeMux()
	s.registerRoutes(mux)
	s.httpServer = &http.Server{Handler: mux}

	// SIGINT / SIGTERM → graceful shutdown.
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		select {
		case <-sigCh:
			fmt.Println("\nShutting down…")
			s.triggerShutdown()
		case <-s.shutdownCh:
		}
	}()

	// Idle timer: checked every 15 s.
	idleTimeout := time.Duration(cfg.IdleTimeoutSeconds) * time.Second
	go func() {
		ticker := time.NewTicker(15 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				last := time.Unix(0, s.lastActivity.Load())
				if time.Since(last) >= idleTimeout {
					fmt.Println("Idle timeout reached, shutting down…")
					s.triggerShutdown()
					return
				}
			case <-s.shutdownCh:
				return
			}
		}
	}()

	addr := ln.Addr().String()
	fmt.Printf("\nkeepassview is running at http://%s\n", addr)
	fmt.Println("Open the URL in your browser. Press Ctrl-C or click Quit to exit.")
	fmt.Printf("Auto-shutdown after %d s of inactivity.\n\n", cfg.IdleTimeoutSeconds)

	err = s.httpServer.Serve(ln)
	if err == http.ErrServerClosed {
		return nil
	}
	return err
}
