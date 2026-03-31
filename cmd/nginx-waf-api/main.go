package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/RumenDamyanov/nginx-waf-api/internal/config"
	"github.com/RumenDamyanov/nginx-waf-api/internal/handler"
	"github.com/RumenDamyanov/nginx-waf-api/internal/lists"
	"github.com/RumenDamyanov/nginx-waf-api/internal/middleware"
	"github.com/RumenDamyanov/nginx-waf-api/internal/reload"
)

var (
	version   = "dev"
	buildTime = "unknown"
)

func main() {
	configPath := flag.String("config", "/etc/nginx-waf-api/config.yaml", "Path to configuration file")
	showVersion := flag.Bool("version", false, "Show version and exit")
	flag.Parse()

	if *showVersion {
		fmt.Printf("nginx-waf-api %s (built %s)\n", version, buildTime)
		os.Exit(0)
	}

	cfg, err := config.Load(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	logger := setupLogger(cfg.Logging.Level, cfg.Logging.Format)

	logger.Info("starting nginx-waf-api",
		"version", version,
		"listen", cfg.Server.Listen,
		"lists_dir", cfg.Nginx.ListsDir,
	)

	mgr := lists.NewManager(cfg.Nginx.ListsDir)
	reloader := reload.New(cfg.Nginx.ReloadCommand, cfg.Nginx.ReloadDebounce, logger)
	defer reloader.Stop()

	h := handler.New(mgr, reloader, logger)

	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	// Build middleware chain
	var chain http.Handler = mux
	if len(cfg.Auth.APIKeys) > 0 {
		chain = middleware.Auth(cfg, logger)(chain)
	}
	chain = middleware.RequestLogger(logger)(chain)

	srv := &http.Server{
		Addr:              cfg.Server.Listen,
		Handler:           chain,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	go func() {
		var err error
		if cfg.Server.TLS.Cert != "" && cfg.Server.TLS.Key != "" {
			logger.Info("TLS enabled")
			err = srv.ListenAndServeTLS(cfg.Server.TLS.Cert, cfg.Server.TLS.Key)
		} else {
			err = srv.ListenAndServe()
		}
		if err != nil && err != http.ErrServerClosed {
			logger.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT)
	sig := <-sigCh

	logger.Info("shutting down", "signal", sig.String())
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	srv.Shutdown(ctx)
}

func setupLogger(level, format string) *slog.Logger {
	var lvl slog.Level
	switch level {
	case "debug":
		lvl = slog.LevelDebug
	case "warn":
		lvl = slog.LevelWarn
	case "error":
		lvl = slog.LevelError
	default:
		lvl = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{Level: lvl}
	var h slog.Handler
	if format == "json" {
		h = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		h = slog.NewTextHandler(os.Stdout, opts)
	}
	return slog.New(h)
}
