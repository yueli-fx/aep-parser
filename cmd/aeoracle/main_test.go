package main

import (
	"context"
	"encoding/json"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"github.com/yueli-fx/aep-parser/internal/aehost"
	"github.com/yueli-fx/aep-parser/internal/aeoracle"
)

func TestResolveRenderRequestPathsAnchorsRelativePathsToCWD(t *testing.T) {
	base := t.TempDir()
	req := aeoracle.RenderRequest{
		SchemaVersion: aeoracle.SchemaVersion,
		AEPPath:       filepath.Join("samples", "input.aep"),
		OutputDir:     filepath.Join("tmp_debug", "renders"),
		Frames:        []aeoracle.FrameTarget{{Frame: 0, Tag: "f000000"}},
	}

	got, err := resolveRenderRequestPaths(req, base)
	if err != nil {
		t.Fatalf("resolveRenderRequestPaths: %v", err)
	}

	if got.AEPPath != filepath.Join(base, "samples", "input.aep") {
		t.Fatalf("AEPPath = %q", got.AEPPath)
	}
	if got.OutputDir != filepath.Join(base, "tmp_debug", "renders") {
		t.Fatalf("OutputDir = %q", got.OutputDir)
	}
	if got.DonePath != filepath.Join(got.OutputDir, "aeoracle_render.done") {
		t.Fatalf("DonePath = %q", got.DonePath)
	}
	if got.MetadataPath != filepath.Join(got.OutputDir, "metadata.json") {
		t.Fatalf("MetadataPath = %q", got.MetadataPath)
	}
}

func TestResolveRenderRequestPathsPreservesAbsolutePaths(t *testing.T) {
	base := t.TempDir()
	absAEP := filepath.Join(base, "input.aep")
	absOut := filepath.Join(base, "out")
	absDone := filepath.Join(base, "done.marker")
	absMeta := filepath.Join(base, "meta.json")
	req := aeoracle.RenderRequest{
		SchemaVersion: aeoracle.SchemaVersion,
		AEPPath:       absAEP,
		OutputDir:     absOut,
		DonePath:      absDone,
		MetadataPath:  absMeta,
		Frames:        []aeoracle.FrameTarget{{Frame: 0, Tag: "f000000"}},
	}

	got, err := resolveRenderRequestPaths(req, filepath.Join(base, "other"))
	if err != nil {
		t.Fatalf("resolveRenderRequestPaths: %v", err)
	}

	if got.AEPPath != absAEP || got.OutputDir != absOut || got.DonePath != absDone || got.MetadataPath != absMeta {
		t.Fatalf("resolved request = %+v", got)
	}
}

func TestReadRenderDoneStatusTrimsWhitespace(t *testing.T) {
	path := filepath.Join(t.TempDir(), "render.done")
	if err := os.WriteFile(path, []byte("ok\r\n"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	got, err := readRenderDoneStatus(path)
	if err != nil {
		t.Fatalf("readRenderDoneStatus: %v", err)
	}
	if got != "ok" {
		t.Fatalf("status = %q, want ok", got)
	}
}

func TestReadRenderDoneStatusKeepsErrorStatus(t *testing.T) {
	path := filepath.Join(t.TempDir(), "render.done")
	if err := os.WriteFile(path, []byte("error"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	got, err := readRenderDoneStatus(path)
	if err != nil {
		t.Fatalf("readRenderDoneStatus: %v", err)
	}
	if got != "error" {
		t.Fatalf("status = %q, want error", got)
	}
}

func TestValidateRenderOutputsRejectsMissingPNG(t *testing.T) {
	dir := t.TempDir()
	req := aeoracle.NewRenderRequest("input.aep", "Main", dir, []aeoracle.FrameTarget{
		{Frame: 0, Seconds: 0, Tag: "f000000", Reason: "comp_start"},
		{Frame: 30, Seconds: 1, Tag: "f000030", Reason: "keyframe"},
	})
	writeCmdRenderMetadata(t, dir, "metadata.json", filepath.Join(dir, "f000000.png"))
	if err := os.WriteFile(filepath.Join(dir, "f000000.png"), []byte("png"), 0o644); err != nil {
		t.Fatalf("WriteFile PNG: %v", err)
	}

	if err := validateRenderOutputs(req); err == nil {
		t.Fatal("validateRenderOutputs error = nil, want missing PNG error")
	}
}

func TestValidateRenderOutputsAcceptsRenderedPNGs(t *testing.T) {
	dir := t.TempDir()
	req := aeoracle.NewRenderRequest("input.aep", "Main", dir, []aeoracle.FrameTarget{
		{Frame: 0, Seconds: 0, Tag: "f000000", Reason: "comp_start"},
		{Frame: 30, Seconds: 1, Tag: "f000030", Reason: "keyframe"},
	})
	writeCmdRenderMetadata(t, dir, "metadata.json", filepath.Join(dir, "f000000.png"))
	for _, frame := range req.Frames {
		if err := os.WriteFile(filepath.Join(dir, frame.Tag+".png"), []byte("png"), 0o644); err != nil {
			t.Fatalf("WriteFile PNG: %v", err)
		}
	}

	if err := validateRenderOutputs(req); err != nil {
		t.Fatalf("validateRenderOutputs: %v", err)
	}
}

func TestRunCloneRequestWritesRequest(t *testing.T) {
	dir := t.TempDir()
	sourcePath := filepath.Join(dir, "source_request.json")
	source := aeoracle.NewRenderRequest("source.aep", "Main", filepath.Join(dir, "source_out"), []aeoracle.FrameTarget{
		{Frame: 0, Seconds: 0, Tag: "f000000", Reason: "comp_start"},
		{Frame: 24, Seconds: 1, Tag: "f000024", Reason: "keyframe"},
	})
	if err := aeoracle.WriteRequest(sourcePath, source); err != nil {
		t.Fatalf("WriteRequest source: %v", err)
	}
	outDir := filepath.Join(dir, "clone_out")

	code := run([]string{
		"clone-request",
		"-from", sourcePath,
		"-aep", "clone.aep",
		"-out", outDir,
		"-comp", "Clone Main",
	})
	if code != 0 {
		t.Fatalf("run clone-request exit = %d", code)
	}

	got, err := aeoracle.ReadRequest(filepath.Join(outDir, "request.json"))
	if err != nil {
		t.Fatalf("ReadRequest clone: %v", err)
	}
	if got.AEPPath != "clone.aep" || got.CompName != "Clone Main" {
		t.Fatalf("clone request = %+v", got)
	}
	if len(got.Frames) != 2 || got.Frames[1].Tag != "f000024" {
		t.Fatalf("Frames = %+v", got.Frames)
	}
}

func TestRunRenderUsesAEHost(t *testing.T) {
	dir := t.TempDir()
	outDir := filepath.Join(dir, "renders")
	reqPath := filepath.Join(dir, "request.json")
	req := aeoracle.NewRenderRequest("input.aep", "Main", outDir, []aeoracle.FrameTarget{
		{Frame: 0, Seconds: 0, Tag: "f000000", Reason: "comp_start"},
	})
	if err := aeoracle.WriteRequest(reqPath, req); err != nil {
		t.Fatalf("WriteRequest: %v", err)
	}
	host := &recordingAEHost{t: t}

	code := runWithHost([]string{
		"render",
		"-request", reqPath,
		"-ae", filepath.Join(dir, "AfterFX.exe"),
		"-jsx", filepath.Join(dir, "render.jsx"),
		"-timeout-sec", "9",
	}, host)
	if code != 0 {
		t.Fatalf("run render exit = %d", code)
	}
	if !host.called {
		t.Fatal("AE host was not called")
	}
	if host.request.TimeoutSec != 9 {
		t.Fatalf("TimeoutSec = %d", host.request.TimeoutSec)
	}
	if host.request.Env["AEORACLE_REQUEST"] == "" || host.request.Env["AEORACLE_CWD"] == "" {
		t.Fatalf("host env = %#v", host.request.Env)
	}
}

func TestRunCompareSetReturnsZeroForIdenticalFrames(t *testing.T) {
	dir := t.TempDir()
	expectedPNG := filepath.Join(dir, "expected.png")
	actualPNG := filepath.Join(dir, "actual.png")
	writeCmdTestPNG(t, expectedPNG, color.RGBA{R: 1, G: 2, B: 3, A: 255})
	writeCmdTestPNG(t, actualPNG, color.RGBA{R: 1, G: 2, B: 3, A: 255})
	expectedMeta := writeCmdRenderMetadata(t, dir, "expected.json", expectedPNG)
	actualMeta := writeCmdRenderMetadata(t, dir, "actual.json", actualPNG)

	code := run([]string{
		"compare-set",
		"-expected-meta", expectedMeta,
		"-actual-meta", actualMeta,
	})
	if code != 0 {
		t.Fatalf("run compare-set exit = %d, want 0", code)
	}
}

func TestRunCompareSetReturnsOneAndWritesReportForDifferentFrames(t *testing.T) {
	dir := t.TempDir()
	expectedPNG := filepath.Join(dir, "expected.png")
	actualPNG := filepath.Join(dir, "actual.png")
	writeCmdTestPNG(t, expectedPNG, color.RGBA{R: 1, G: 2, B: 3, A: 255})
	writeCmdTestPNG(t, actualPNG, color.RGBA{R: 1, G: 2, B: 4, A: 255})
	expectedMeta := writeCmdRenderMetadata(t, dir, "expected.json", expectedPNG)
	actualMeta := writeCmdRenderMetadata(t, dir, "actual.json", actualPNG)
	outPath := filepath.Join(dir, "report.json")

	code := run([]string{
		"compare-set",
		"-expected-meta", expectedMeta,
		"-actual-meta", actualMeta,
		"-out", outPath,
	})
	if code != 1 {
		t.Fatalf("run compare-set exit = %d, want 1", code)
	}
	report, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("ReadFile report: %v", err)
	}
	if len(report) == 0 {
		t.Fatal("report file is empty")
	}
}

type recordingAEHost struct {
	t       *testing.T
	called  bool
	request aehost.ScriptRequest
}

func (h *recordingAEHost) Available(context.Context) aehost.Availability {
	return aehost.Availability{Status: aehost.CapabilityAvailable}
}

func (h *recordingAEHost) RunScript(_ context.Context, req aehost.ScriptRequest) (aehost.ScriptResult, error) {
	h.called = true
	h.request = req

	requestPath := req.Env["AEORACLE_REQUEST"]
	renderReq, err := aeoracle.ReadRequest(requestPath)
	if err != nil {
		h.t.Fatalf("ReadRequest: %v", err)
	}
	if err := os.MkdirAll(renderReq.OutputDir, 0o755); err != nil {
		h.t.Fatalf("MkdirAll: %v", err)
	}
	if err := os.WriteFile(renderReq.DonePath, []byte("ok\n"), 0o644); err != nil {
		h.t.Fatalf("WriteFile done: %v", err)
	}
	for _, frame := range renderReq.Frames {
		writeCmdTestPNG(h.t, filepath.Join(renderReq.OutputDir, frame.Tag+".png"), color.RGBA{R: 1, G: 2, B: 3, A: 255})
	}
	writeCmdRenderMetadata(h.t, renderReq.OutputDir, filepath.Base(renderReq.MetadataPath), filepath.Join(renderReq.OutputDir, renderReq.Frames[0].Tag+".png"))
	return aehost.ScriptResult{ExitCode: 0, DonePath: req.DonePath}, nil
}

func writeCmdRenderMetadata(t *testing.T, dir, name, pngPath string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	meta := aeoracle.RenderMetadata{
		SchemaVersion: aeoracle.SchemaVersion,
		AEPPath:       name + ".aep",
		CompName:      "Main",
		OutputDir:     dir,
		Status:        "ok",
		Frames: []aeoracle.RenderedFrameRecord{{
			Frame:      0,
			Seconds:    0,
			Tag:        "f000000",
			Reason:     "comp_start",
			OutputPath: pngPath,
			Status:     "rendered",
		}},
	}
	data, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		t.Fatalf("MarshalIndent: %v", err)
	}
	if err := os.WriteFile(path, append(data, '\n'), 0o644); err != nil {
		t.Fatalf("WriteFile metadata: %v", err)
	}
	return path
}

func writeCmdTestPNG(t *testing.T, path string, pixel color.RGBA) {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, 1, 1))
	img.SetRGBA(0, 0, pixel)
	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("Create PNG: %v", err)
	}
	defer f.Close()
	if err := png.Encode(f, img); err != nil {
		t.Fatalf("Encode PNG: %v", err)
	}
}
