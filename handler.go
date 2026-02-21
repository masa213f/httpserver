package main

import (
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"math/rand"
	"net/http"
	"net/http/httputil"
	"path/filepath"
	"strings"
	"time"
)

const HeaderRequestID = "X-Request-Id"

func newRequestId() string {
	buf := make([]byte, binary.MaxVarintLen32)
	binary.BigEndian.PutUint32(buf, uint32(rand.Int31()))
	return hex.EncodeToString(buf)
}

func newHandler(serverMode string) (http.Handler, error) {
	var mode, opt string
	split := strings.SplitN(serverMode, "=", 2)
	mode = split[0]
	if len(split) == 2 {
		opt = split[1]
	}

	var handler http.Handler
	switch mode {
	case "text":
		handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			io.WriteString(w, opt)
		})

	case "file":
		absPath, err := filepath.Abs(opt)
		if err != nil {
			return nil, fmt.Errorf("failed to get absolute path: %v", err)
		}
		handler = http.FileServer(http.Dir(absPath))

	case "dump":
		handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			b, err := httputil.DumpRequest(r, true)
			if err != nil {
				io.WriteString(w, err.Error())
			} else {
				w.Write(b)
			}
		})

	default:
		return nil, errors.New("invalid mode")
	}

	return withLogger(handler), nil
}

type wrapResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *wrapResponseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func withLogger(handler http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get(HeaderRequestID)
		if requestID == "" {
			requestID = newRequestId()
		}

		wrap := &wrapResponseWriter{w, http.StatusOK}
		start := time.Now()
		handler.ServeHTTP(wrap, r)
		msec := time.Since(start).Milliseconds()

		slog.Info("s",
			slog.String("request_id", requestID),
			slog.String("path", r.URL.Path),
			slog.String("method", r.Method),
			slog.String("query", r.URL.RawQuery),
			slog.String("referer", r.Referer()),
			slog.String("user_agent", r.UserAgent()),
			slog.Int64("duration_ms", msec),
			slog.Int("status_code", wrap.statusCode),
		)
	}
}
