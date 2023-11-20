package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/google/uuid"
)

const HeaderRequestID = "X-Request-Id"

var (
	hello    bool
	hostname bool
	bindAddr string
	rootDir  string
)

func init() {
	const usage = `Usage: testhttpserver

Options:
  -addr             listen address and port
  -dir              root dir of file server
  -hello
  -hostname
  -h, -help         display this help and exit
`
	flag.Usage = func() { fmt.Fprint(flag.CommandLine.Output(), usage) }
	flag.BoolVar(&hello, "hello", false, "")
	flag.BoolVar(&hostname, "hostname", false, "")
	flag.StringVar(&bindAddr, "addr", ":8080", "")
	flag.StringVar(&rootDir, "dir", ".", "")
}

func main() {
	flag.Parse()

	dropEmpty := func(group []string, a slog.Attr) slog.Attr {
		if a.Value.Kind() == slog.KindString && a.Value.String() == "" {
			return slog.Attr{}
		}
		return a
	}
	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{ReplaceAttr: dropEmpty}))

	var handler http.Handler
	switch {
	case hello:
		text := os.Getenv("TEXT")
		if text == "" {
			text = "hello"
		}
		mux := http.NewServeMux()
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			io.WriteString(w, text)
		})
		handler = mux
		log.Info("run as hello server", slog.String("text", text))

	case hostname:
		name, err := os.Hostname()
		if err != nil {
			log.Error("failed to get hostname", "error", err)
			os.Exit(1)
		}
		mux := http.NewServeMux()
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			io.WriteString(w, name)
		})
		handler = mux
		log.Info("run as hostname server", slog.String("hostname", name))

	default:
		abspath, err := filepath.Abs(rootDir)
		if err != nil {
			log.Error("failed to get absolute path", "error", err)
			os.Exit(1)
		}
		handler = http.FileServer(http.Dir(abspath))
		log.Info("run as file server", slog.String("root_dir", abspath))
	}

	server := &http.Server{Addr: bindAddr, Handler: loggerHandler(log, handler)}
	errCh := make(chan error)
	go func() {
		errCh <- server.ListenAndServe()
	}()
	log.Info("start", "addr", bindAddr)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-errCh:
		// ListenAndServe always returns a non-nil error. So no need for a nil check.
		log.Error("failed to listen", "error", err)
		os.Exit(1)
	case sig := <-sigCh:
		log.Info("catch signal", "signal", sig)
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := server.Shutdown(ctx); err != nil {
			log.Error("failed to shutdown", "error", err)
			os.Exit(1)
		}
	}
	log.Info("bye")
}

type wrapResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *wrapResponseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func loggerHandler(log *slog.Logger, handler http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get(HeaderRequestID)
		if requestID == "" {
			requestID = uuid.New().String()
			r.Header.Set(HeaderRequestID, requestID)
		}
		wrap := &wrapResponseWriter{w, http.StatusOK}

		start := time.Now()
		handler.ServeHTTP(w, r)
		msec := time.Since(start).Milliseconds()

		log.Info("request",
			slog.String("request_id", requestID),
			slog.Int64("duration_ms", msec),
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.String("query", r.URL.RawQuery),
			slog.Int("status_code", wrap.statusCode),
			slog.String("referer", r.Referer()),
			slog.String("user_agent", r.UserAgent()),
		)
	}
}
