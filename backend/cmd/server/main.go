// Package main is the entrypoint for the barq-slides ConnectRPC server.
package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"

	barqv1 "github.com/YASSERRMD/barq-slides/gen/barq/v1"
	"github.com/YASSERRMD/barq-slides/internal/api"
	"github.com/YASSERRMD/barq-slides/internal/config"
	"github.com/YASSERRMD/barq-slides/internal/llm"
	"github.com/YASSERRMD/barq-slides/internal/logger"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
		os.Exit(1)
	}

	log := logger.New(cfg.Log.Level, cfg.Log.Format)
	log.Info("barq-slides server starting",
		"host", cfg.Server.Host,
		"port", cfg.Server.Port,
		"provider", cfg.LLM.DefaultProvider,
	)

	factory := llm.NewFactory(cfg)
	// We no longer fatally exit if the default provider fails to initialize
	// because the API key could be dynamically provided at runtime by the user.
	if _, err := factory.Default(); err != nil {
		log.Warn("default llm provider failed to initialize (will require frontend headers)", "error", err)
	}

	handler := api.New(log, factory)
	path, svcHandler := barqv1.NewHeliosServiceHandler(handler)

	mux := http.NewServeMux()
	mux.Handle(path, withCORS(cfg.Server.CORSAllowOrigin, svcHandler))
	mux.HandleFunc("/healthz", healthHandler)
	mux.HandleFunc("/readyz", healthHandler)

	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	srv := &http.Server{
		Addr:         addr,
		Handler:      h2c.NewHandler(mux, &http2.Server{}),
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 0, // streaming — no write timeout
		IdleTimeout:  120 * time.Second,
	}

	go func() {
		log.Info("server listening", "addr", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Error("server forced shutdown", "error", err)
		os.Exit(1)
	}

	log.Info("server exited gracefully")
}

func healthHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = fmt.Fprintf(w, `{"status":"ok"}`)
}

// withCORS adds permissive CORS headers required by browser ConnectRPC clients.
// Connect protocol requires application/connect+json content-type for streaming
// and specific headers for the Connect envelope protocol.
func withCORS(allowOrigin string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		// Allow the configured origin or any origin if allowOrigin is "*".
		if allowOrigin == "*" || origin == allowOrigin {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		} else if allowOrigin != "" {
			w.Header().Set("Access-Control-Allow-Origin", allowOrigin)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
		}
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		// All headers required by the Connect protocol (unary + streaming).
		w.Header().Set("Access-Control-Allow-Headers",
			"Accept, Content-Type, Content-Length, Accept-Encoding, "+
				"Connect-Protocol-Version, Connect-Timeout-Ms, "+
				"Connect-Accept-Encoding, Connect-Content-Encoding, "+
				"Grpc-Timeout, X-Grpc-Web, X-User-Agent, "+
				"X-Request-ID, X-Llm-Provider, X-Llm-Api-Key, X-Llm-Model, X-Llm-Base-Url")
		// Expose Connect trailer headers so the browser can read them.
		w.Header().Set("Access-Control-Expose-Headers",
			"Connect-Content-Encoding, Grpc-Status, Grpc-Message, "+
				"Grpc-Status-Details-Bin, Trailer, Trailers")
		w.Header().Set("Access-Control-Max-Age", "7200")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
