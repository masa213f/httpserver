package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

const usage = `Usage: httpserver

Options:
  -listen       listen address and port
  -mode         server mode
  -v,           display version and exit
  -h, -help     display this help and exit
`

var Version = "develop"

var (
	listenAddr  string
	serverMode  string
	showVersion bool
)

func init() {
	flag.Usage = func() { fmt.Fprint(flag.CommandLine.Output(), usage) }
	flag.StringVar(&listenAddr, "listen", ":8080", "")
	flag.StringVar(&serverMode, "mode", "text=hello", "")
	flag.BoolVar(&showVersion, "v", false, "")
}

func newLogger() *slog.Logger {
	replace := func(group []string, a slog.Attr) slog.Attr {
		if a.Key == slog.TimeKey {
			// display in UTC
			return slog.Time(slog.TimeKey, a.Value.Time().UTC())
		}
		if a.Value.Kind() == slog.KindString && a.Value.String() == "" {
			// drop empty field
			return slog.Attr{}
		}
		return a
	}
	return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{ReplaceAttr: replace}))
}

func main() {
	flag.Parse()
	slog.SetDefault(newLogger())

	if showVersion {
		fmt.Println(Version)
		os.Exit(0)
	}

	handler, err := newHandler(serverMode)
	if err != nil {
		slog.Error("failed to initialize handler", "error", err)
		os.Exit(1)
	}

	var listener net.Listener
	if socket, found := strings.CutPrefix(listenAddr, "unix:"); found {
		ln, err := net.Listen("unix", socket)
		if err != nil {
			slog.Error("failed to listen unix domain socket", "error", err)
			os.Exit(1)
		}
		listener = ln
	} else {
		ln, err := net.Listen("tcp", listenAddr)
		if err != nil {
			slog.Error("failed to listen tcp", "error", err)
			os.Exit(1)
		}
		listener = ln
	}
	defer listener.Close()

	slog.Info("hello", "version", Version, "mode", serverMode, "listen", listenAddr)

	server := &http.Server{Handler: handler}
	errCh := make(chan error)
	go func() {
		errCh <- server.Serve(listener)
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-errCh:
		// Serve always returns a non-nil error. So no need for a nil check.
		slog.Error("failed to listen", "error", err)
		os.Exit(1)
	case sig := <-sigCh:
		slog.Info("catch signal", "signal", sig)
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := server.Shutdown(ctx); err != nil {
			slog.Error("failed to shutdown", "error", err)
			os.Exit(1)
		}
	}

	slog.Info("bye")
}
