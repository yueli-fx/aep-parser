package main

import (
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"

	aepserver "github.com/yueli-fx/aep-parser/internal/server"
)

type serveOptions struct {
	Addr           string
	MaxBodyBytes   int64
	AllowPathInput bool
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
	allowPathInput := fs.Bool("allow-path-input", false, "allow JSON path input; disabled by default for public service safety")
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
	if serveFn == nil {
		serveFn = serve
	}
	if err := serveFn(serveOptions{
		Addr:           *addr,
		MaxBodyBytes:   *maxBodyMB << 20,
		AllowPathInput: *allowPathInput,
	}); err != nil {
		fmt.Fprintln(stderr, err)
		return 2
	}
	return 0
}

func serve(opts serveOptions) error {
	server := &http.Server{
		Addr: opts.Addr,
		Handler: aepserver.NewHandler(aepserver.Options{
			MaxBodyBytes:   opts.MaxBodyBytes,
			AllowPathInput: opts.AllowPathInput,
		}),
	}
	fmt.Printf("aepserver listening on http://%s\n", opts.Addr)
	return server.ListenAndServe()
}
