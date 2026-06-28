package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/example/aep-parser/internal/aeoracle"
	aep "github.com/example/aep-parser/internal/aep"
	"github.com/example/aep-parser/internal/profile"
)

func main() {
	os.Exit(run(os.Args[1:]))
}

func run(args []string) int {
	if len(args) == 0 {
		usage()
		return 2
	}
	switch args[0] {
	case "plan":
		return runPlan(args[1:])
	case "compare":
		return runCompare(args[1:])
	case "render":
		return runRender(args[1:])
	default:
		usage()
		return 2
	}
}

func runPlan(args []string) int {
	fs := flag.NewFlagSet("aeoracle plan", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	aepPath := fs.String("aep", "", "AEP file to plan frames for")
	compName := fs.String("comp", "", "composition name; empty uses first comp")
	outDir := fs.String("out", "", "output directory for request and renders")
	maxFrames := fs.Int("max-frames", 8, "maximum sentinel frames")
	jsonOut := fs.Bool("json", false, "print request JSON to stdout")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if *aepPath == "" || *outDir == "" {
		fmt.Fprintln(os.Stderr, "usage: aeoracle plan -aep file.aep -out dir [-comp name] [-max-frames n] [-json]")
		return 2
	}
	project, err := aep.Open(*aepPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "open:", err)
		return 2
	}
	prof, err := profile.Build(project, profile.Options{Path: *aepPath})
	if err != nil {
		fmt.Fprintln(os.Stderr, "profile:", err)
		return 2
	}
	frames, err := aeoracle.SelectFrames(prof, aeoracle.FrameOptions{CompName: *compName, MaxFrames: *maxFrames})
	if err != nil {
		fmt.Fprintln(os.Stderr, "frames:", err)
		return 2
	}
	req := aeoracle.NewRenderRequest(*aepPath, *compName, *outDir, frames)
	reqPath := filepath.Join(*outDir, "request.json")
	if err := aeoracle.WriteRequest(reqPath, req); err != nil {
		fmt.Fprintln(os.Stderr, "write request:", err)
		return 2
	}
	if *jsonOut {
		return writeJSON(req)
	}
	fmt.Printf("request: %s\nframes: %d\n", reqPath, len(req.Frames))
	for _, frame := range req.Frames {
		fmt.Printf("%s frame=%d seconds=%.6f reason=%s\n", frame.Tag, frame.Frame, frame.Seconds, frame.Reason)
	}
	return 0
}

func runCompare(args []string) int {
	fs := flag.NewFlagSet("aeoracle compare", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	expected := fs.String("expected", "", "expected PNG path")
	actual := fs.String("actual", "", "actual PNG path")
	threshold := fs.Int("threshold", 0, "per-channel threshold 0..255")
	jsonOut := fs.Bool("json", false, "print JSON report")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if *expected == "" || *actual == "" || *threshold < 0 || *threshold > 255 {
		fmt.Fprintln(os.Stderr, "usage: aeoracle compare -expected a.png -actual b.png [-threshold 0..255] [-json]")
		return 2
	}
	report, err := aeoracle.ComparePNG(*expected, *actual, aeoracle.CompareOptions{ChannelThreshold: uint8(*threshold)})
	if err != nil {
		fmt.Fprintln(os.Stderr, "compare:", err)
		return 2
	}
	if *jsonOut {
		return writeJSON(report)
	}
	fmt.Printf("pixels: %d different: %d (%.4f%%) max_channel_delta: %d threshold: %d\n",
		report.TotalPixels, report.DifferentPixels, report.DifferentPercent, report.MaxChannelDelta, report.ChannelThreshold)
	if report.DifferentPixels > 0 {
		return 1
	}
	return 0
}

func runRender(args []string) int {
	fs := flag.NewFlagSet("aeoracle render", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	requestPath := fs.String("request", "", "render request JSON path")
	aePath := fs.String("ae", "", "AfterFX.exe path")
	jsxPath := fs.String("jsx", "scripts/aeoracle_render.jsx", "renderer JSX path")
	timeout := fs.Int("timeout-sec", 180, "AE automation timeout seconds")
	dryRun := fs.Bool("dry-run", false, "print command without launching AE")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if *requestPath == "" {
		fmt.Fprintln(os.Stderr, "usage: aeoracle render -request request.json [-ae AfterFX.exe] [-jsx scripts/aeoracle_render.jsx] [-timeout-sec n] [-dry-run]")
		return 2
	}
	req, err := aeoracle.ReadRequest(*requestPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "request:", err)
		return 2
	}
	if *dryRun {
		fmt.Println(formatCommand("pwsh", renderCommand(*aePath, *jsxPath, req.DonePath, *timeout)))
		return 0
	}
	if *aePath == "" {
		fmt.Fprintln(os.Stderr, "render: -ae is required unless -dry-run is set")
		return 2
	}
	cmdArgs := renderCommand(*aePath, *jsxPath, req.DonePath, *timeout)
	cmd := exec.Command("pwsh", cmdArgs...)
	cmd.Env = append(os.Environ(), "AEORACLE_REQUEST="+*requestPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "render:", err)
		return 2
	}
	return 0
}

func renderCommand(aePath, jsxPath, donePath string, timeout int) []string {
	return []string{
		"-NoProfile",
		"-File", "scripts/ae_run.ps1",
		"-AeExe", aePath,
		"-Jsx", jsxPath,
		"-Done", donePath,
		"-TimeoutSec", strconv.Itoa(timeout),
	}
}

func writeJSON(v any) int {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(v); err != nil {
		fmt.Fprintln(os.Stderr, "json:", err)
		return 2
	}
	return 0
}

func quoteArg(v string) string {
	if v == "" {
		return `""`
	}
	if strings.ContainsAny(v, " \t\"") {
		return strconv.Quote(v)
	}
	return v
}

func formatCommand(name string, args []string) string {
	parts := []string{name}
	for _, arg := range args {
		parts = append(parts, quoteArg(arg))
	}
	return strings.Join(parts, " ")
}

func usage() {
	fmt.Fprintln(os.Stderr, "usage: aeoracle <plan|render|compare> [flags]")
}
