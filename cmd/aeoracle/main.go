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

	"github.com/yueli-fx/aep-parser/internal/aeoracle"
	"github.com/yueli-fx/aep-parser/internal/aep"
	"github.com/yueli-fx/aep-parser/internal/profile"
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
	case "clone-request":
		return runCloneRequest(args[1:])
	case "compare":
		return runCompare(args[1:])
	case "compare-set":
		return runCompareSet(args[1:])
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

func runCloneRequest(args []string) int {
	fs := flag.NewFlagSet("aeoracle clone-request", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	fromPath := fs.String("from", "", "source render request JSON path")
	aepPath := fs.String("aep", "", "clone AEP path")
	compName := fs.String("comp", "", "override comp name; empty preserves source request comp")
	outDir := fs.String("out", "", "output directory for cloned request and renders")
	jsonOut := fs.Bool("json", false, "print cloned request JSON to stdout")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if *fromPath == "" || *aepPath == "" || *outDir == "" {
		fmt.Fprintln(os.Stderr, "usage: aeoracle clone-request -from request.json -aep clone.aep -out dir [-comp name] [-json]")
		return 2
	}
	source, err := aeoracle.ReadRequest(*fromPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "from:", err)
		return 2
	}
	req := aeoracle.CloneRenderRequest(source, *aepPath, *compName, *outDir)
	reqPath := filepath.Join(*outDir, "request.json")
	if err := aeoracle.WriteRequest(reqPath, req); err != nil {
		fmt.Fprintln(os.Stderr, "write request:", err)
		return 2
	}
	if *jsonOut {
		return writeJSON(req)
	}
	fmt.Printf("request: %s\nframes: %d\n", reqPath, len(req.Frames))
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

func runCompareSet(args []string) int {
	fs := flag.NewFlagSet("aeoracle compare-set", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	expectedMeta := fs.String("expected-meta", "", "expected render metadata JSON path")
	actualMeta := fs.String("actual-meta", "", "actual render metadata JSON path")
	threshold := fs.Int("threshold", 0, "per-channel threshold 0..255")
	jsonOut := fs.Bool("json", false, "print JSON report")
	outPath := fs.String("out", "", "write JSON report to path")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if *expectedMeta == "" || *actualMeta == "" || *threshold < 0 || *threshold > 255 {
		fmt.Fprintln(os.Stderr, "usage: aeoracle compare-set -expected-meta expected.json -actual-meta actual.json [-threshold 0..255] [-json] [-out report.json]")
		return 2
	}
	report, err := aeoracle.CompareFrameSets(*expectedMeta, *actualMeta, aeoracle.CompareOptions{ChannelThreshold: uint8(*threshold)})
	if err != nil {
		fmt.Fprintln(os.Stderr, "compare-set:", err)
		return 2
	}
	if *outPath != "" {
		if err := writeJSONFile(*outPath, report); err != nil {
			fmt.Fprintln(os.Stderr, "write report:", err)
			return 2
		}
	}
	if *jsonOut {
		if code := writeJSON(report); code != 0 {
			return code
		}
	} else {
		fmt.Printf("frames: %d ok: %d different: %d missing_expected: %d missing_actual: %d\n",
			report.Summary.TotalFrames,
			report.Summary.OKFrames,
			report.Summary.DifferentFrames,
			report.Summary.MissingExpectedFrames,
			report.Summary.MissingActualFrames)
		for _, frame := range report.Frames {
			if frame.Status == aeoracle.FrameStatusOK {
				continue
			}
			fmt.Printf("%s frame=%d seconds=%.6f status=%s\n", frame.Tag, frame.Frame, frame.Seconds, frame.Status)
		}
	}
	if report.Summary.DifferentFrames > 0 || report.Summary.MissingExpectedFrames > 0 || report.Summary.MissingActualFrames > 0 {
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
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintln(os.Stderr, "cwd:", err)
		return 2
	}
	requestAbs, err := filepath.Abs(*requestPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "request path:", err)
		return 2
	}
	req, err := aeoracle.ReadRequest(requestAbs)
	if err != nil {
		fmt.Fprintln(os.Stderr, "request:", err)
		return 2
	}
	req, err = resolveRenderRequestPaths(req, cwd)
	if err != nil {
		fmt.Fprintln(os.Stderr, "request paths:", err)
		return 2
	}
	jsxAbs, err := resolvePath(*jsxPath, cwd)
	if err != nil {
		fmt.Fprintln(os.Stderr, "jsx path:", err)
		return 2
	}
	runScript, err := resolvePath(filepath.Join("scripts", "ae_run.ps1"), cwd)
	if err != nil {
		fmt.Fprintln(os.Stderr, "runner path:", err)
		return 2
	}
	aeExe := *aePath
	if aeExe != "" {
		aeExe, err = resolvePath(aeExe, cwd)
		if err != nil {
			fmt.Fprintln(os.Stderr, "ae path:", err)
			return 2
		}
	}
	if *dryRun {
		fmt.Printf("AEORACLE_REQUEST=%s\nAEORACLE_CWD=%s\n%s\n", requestAbs, cwd, formatCommand("pwsh", renderCommand(runScript, aeExe, jsxAbs, req.DonePath, *timeout)))
		return 0
	}
	if aeExe == "" {
		fmt.Fprintln(os.Stderr, "render: -ae is required unless -dry-run is set")
		return 2
	}
	cmdArgs := renderCommand(runScript, aeExe, jsxAbs, req.DonePath, *timeout)
	cmd := exec.Command("pwsh", cmdArgs...)
	cmd.Env = append(os.Environ(), "AEORACLE_REQUEST="+requestAbs, "AEORACLE_CWD="+cwd)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "render:", err)
		return 2
	}
	status, err := readRenderDoneStatus(req.DonePath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "render status:", err)
		return 2
	}
	if status != "ok" {
		fmt.Fprintf(os.Stderr, "render: AE reported status %q\n", status)
		return 2
	}
	if err := validateRenderOutputs(req); err != nil {
		fmt.Fprintln(os.Stderr, "render outputs:", err)
		return 2
	}
	return 0
}

func renderCommand(runScript, aePath, jsxPath, donePath string, timeout int) []string {
	return []string{
		"-NoProfile",
		"-File", runScript,
		"-AeExe", aePath,
		"-Jsx", jsxPath,
		"-Done", donePath,
		"-TimeoutSec", strconv.Itoa(timeout),
	}
}

func resolveRenderRequestPaths(req aeoracle.RenderRequest, baseDir string) (aeoracle.RenderRequest, error) {
	var err error
	if req.AEPPath, err = resolvePath(req.AEPPath, baseDir); err != nil {
		return aeoracle.RenderRequest{}, err
	}
	if req.OutputDir, err = resolvePath(req.OutputDir, baseDir); err != nil {
		return aeoracle.RenderRequest{}, err
	}
	if req.DonePath == "" {
		req.DonePath = filepath.Join(req.OutputDir, "aeoracle_render.done")
	} else if req.DonePath, err = resolvePath(req.DonePath, baseDir); err != nil {
		return aeoracle.RenderRequest{}, err
	}
	if req.MetadataPath == "" {
		req.MetadataPath = filepath.Join(req.OutputDir, "metadata.json")
	} else if req.MetadataPath, err = resolvePath(req.MetadataPath, baseDir); err != nil {
		return aeoracle.RenderRequest{}, err
	}
	return req, nil
}

func resolvePath(path string, baseDir string) (string, error) {
	if path == "" {
		return "", nil
	}
	if filepath.IsAbs(path) {
		return filepath.Clean(path), nil
	}
	return filepath.Abs(filepath.Join(baseDir, path))
}

func readRenderDoneStatus(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(data)), nil
}

func validateRenderOutputs(req aeoracle.RenderRequest) error {
	if req.MetadataPath != "" {
		meta, err := aeoracle.ReadRenderMetadata(req.MetadataPath)
		if err != nil {
			return fmt.Errorf("metadata: %w", err)
		}
		if meta.Status != "ok" {
			return fmt.Errorf("metadata status %q", meta.Status)
		}
	}
	for _, frame := range req.Frames {
		path := filepath.Join(req.OutputDir, frame.Tag+".png")
		info, err := os.Stat(path)
		if err != nil {
			return fmt.Errorf("frame %s: %w", frame.Tag, err)
		}
		if info.Size() == 0 {
			return fmt.Errorf("frame %s: empty PNG %s", frame.Tag, path)
		}
	}
	return nil
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

func writeJSONFile(path string, v any) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(data, '\n'), 0o644)
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
	fmt.Fprintln(os.Stderr, "usage: aeoracle <plan|clone-request|render|compare|compare-set> [flags]")
}
