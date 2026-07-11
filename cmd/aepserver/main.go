package main

import (
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	aepserver "github.com/yueli-fx/aep-parser/internal/server"
)

type serveOptions struct {
	Addr              string
	MaxBodyBytes      int64
	MaxConcurrent     int
	AllowPathInput    bool
	AllowedPathRoots  []string
	ReadHeaderTimeout time.Duration
	ReadTimeout       time.Duration
	WriteTimeout      time.Duration
	IdleTimeout       time.Duration
}

type serveFunc func(serveOptions) error

func main() {
	os.Exit(runWithIO(os.Args[1:], os.Stdout, os.Stderr, serve))
}

func runWithIO(args []string, stdout, stderr io.Writer, serveFn serveFunc) int {
	fs := flag.NewFlagSet("aepserver", flag.ContinueOnError)
	fs.SetOutput(stdout)
	fs.Usage = func() {
		fmt.Fprintln(stdout, "usage: aepserver [-addr host:port] [-max-body-mb n] [-allow-path-input]")
		fs.PrintDefaults()
	}
	addr := fs.String("addr", "127.0.0.1:8080", "HTTP listen address")
	maxBodyMB := fs.Int64("max-body-mb", 256, "maximum uploaded AEP body size in MiB")
	maxConcurrent := fs.Int("max-concurrent", 32, "maximum concurrent requests")
	allowPathInput := fs.Bool("allow-path-input", false, "allow JSON path input; disabled by default for public service safety")
	allowedPathRoots := fs.String("allowed-path-roots", "", "comma-separated roots for JSON path input")
	readHeaderTimeout := fs.Duration("read-header-timeout", 5*time.Second, "maximum time to read request headers")
	readTimeout := fs.Duration("read-timeout", 30*time.Second, "maximum time to read a request")
	writeTimeout := fs.Duration("write-timeout", 60*time.Second, "maximum time to write a response")
	idleTimeout := fs.Duration("idle-timeout", 120*time.Second, "maximum keep-alive idle time")
	if err := fs.Parse(args); err != nil {
		if err == flag.ErrHelp {
			return 0
		}
		return 2
	}
	if *addr == "" {
		fmt.Fprintln(stderr, "-addr is required")
		return 2
	}
	if *maxBodyMB <= 0 {
		fmt.Fprintln(stderr, "-max-body-mb must be positive")
		return 2
	}
	if *maxConcurrent <= 0 {
		fmt.Fprintln(stderr, "-max-concurrent must be positive")
		return 2
	}
	if *readHeaderTimeout <= 0 || *readTimeout <= 0 || *writeTimeout <= 0 || *idleTimeout <= 0 {
		fmt.Fprintln(stderr, "server timeouts must be positive")
		return 2
	}
	roots := splitRoots(*allowedPathRoots)
	if *allowPathInput && len(roots) == 0 {
		fmt.Fprintln(stderr, "-allow-path-input requires -allowed-path-roots")
		return 2
	}
	if serveFn == nil {
		serveFn = serve
	}
	if err := serveFn(serveOptions{
		Addr:              *addr,
		MaxBodyBytes:      *maxBodyMB << 20,
		MaxConcurrent:     *maxConcurrent,
		AllowPathInput:    *allowPathInput,
		AllowedPathRoots:  roots,
		ReadHeaderTimeout: *readHeaderTimeout,
		ReadTimeout:       *readTimeout,
		WriteTimeout:      *writeTimeout,
		IdleTimeout:       *idleTimeout,
	}); err != nil {
		fmt.Fprintln(stderr, err)
		return 2
	}
	return 0
}

func serve(opts serveOptions) error {
	server := &http.Server{
		Addr:              opts.Addr,
		ReadHeaderTimeout: opts.ReadHeaderTimeout,
		ReadTimeout:       opts.ReadTimeout,
		WriteTimeout:      opts.WriteTimeout,
		IdleTimeout:       opts.IdleTimeout,
		Handler: aepserver.NewHandler(aepserver.Options{
			MaxBodyBytes:     opts.MaxBodyBytes,
			MaxConcurrent:    opts.MaxConcurrent,
			AllowPathInput:   opts.AllowPathInput,
			AllowedPathRoots: opts.AllowedPathRoots,
		}),
	}
	fmt.Printf("aepserver listening on http://%s\n", opts.Addr)
	return server.ListenAndServe()
}

func splitRoots(value string) []string {
	var roots []string
	for _, root := range strings.Split(value, ",") {
		if root = strings.TrimSpace(root); root != "" {
			roots = append(roots, root)
		}
	}
	return roots
}
