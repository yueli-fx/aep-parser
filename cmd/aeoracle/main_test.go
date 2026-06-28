package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/example/aep-parser/internal/aeoracle"
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
