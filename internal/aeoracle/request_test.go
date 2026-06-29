package aeoracle_test

import (
	"path/filepath"
	"testing"

	"github.com/yueli-fx/aep-parser/internal/aeoracle"
)

func TestRenderRequestRoundTrip(t *testing.T) {
	dir := t.TempDir()
	req := aeoracle.NewRenderRequest("input.aep", "Main", filepath.Join(dir, "out"), []aeoracle.FrameTarget{
		{Frame: 0, Seconds: 0, Tag: "f000000", Reason: "comp_start"},
		{Frame: 30, Seconds: 1, Tag: "f000030", Reason: "keyframe"},
	})

	path := filepath.Join(dir, "request.json")
	if err := aeoracle.WriteRequest(path, req); err != nil {
		t.Fatalf("WriteRequest: %v", err)
	}
	got, err := aeoracle.ReadRequest(path)
	if err != nil {
		t.Fatalf("ReadRequest: %v", err)
	}
	if got.SchemaVersion != aeoracle.SchemaVersion {
		t.Fatalf("SchemaVersion = %d, want %d", got.SchemaVersion, aeoracle.SchemaVersion)
	}
	if got.AEPPath != "input.aep" || got.CompName != "Main" || got.OutputDir == "" {
		t.Fatalf("round-trip request = %+v", got)
	}
	if got.DonePath == "" || got.MetadataPath == "" {
		t.Fatalf("DonePath/MetadataPath not populated: %+v", got)
	}
	if got.FrameTimeoutMS != 120000 {
		t.Fatalf("FrameTimeoutMS = %d, want 120000", got.FrameTimeoutMS)
	}
	if len(got.Frames) != 2 || got.Frames[1].Tag != "f000030" {
		t.Fatalf("Frames = %+v", got.Frames)
	}
}

func TestCloneRenderRequestPreservesFrameTargets(t *testing.T) {
	source := aeoracle.NewRenderRequest("source.aep", "Main", filepath.Join("tmp", "source"), []aeoracle.FrameTarget{
		{Frame: 0, Seconds: 0, Tag: "f000000", Reason: "comp_start"},
		{Frame: 24, Seconds: 1, Tag: "f000024", Reason: "keyframe"},
	})

	got := aeoracle.CloneRenderRequest(source, "clone.aep", "", filepath.Join("tmp", "clone"))

	if got.AEPPath != "clone.aep" {
		t.Fatalf("AEPPath = %q", got.AEPPath)
	}
	if got.CompName != "Main" {
		t.Fatalf("CompName = %q, want Main", got.CompName)
	}
	if got.OutputDir != filepath.Join("tmp", "clone") {
		t.Fatalf("OutputDir = %q", got.OutputDir)
	}
	if got.DonePath != filepath.Join("tmp", "clone", "aeoracle_render.done") {
		t.Fatalf("DonePath = %q", got.DonePath)
	}
	if got.MetadataPath != filepath.Join("tmp", "clone", "metadata.json") {
		t.Fatalf("MetadataPath = %q", got.MetadataPath)
	}
	if len(got.Frames) != 2 || got.Frames[1].Tag != "f000024" || got.Frames[1].Seconds != 1 {
		t.Fatalf("Frames = %+v", got.Frames)
	}

	got.Frames[0].Tag = "changed"
	if source.Frames[0].Tag != "f000000" {
		t.Fatalf("CloneRenderRequest aliased source frames: %+v", source.Frames)
	}
}

func TestCloneRenderRequestAllowsCompOverride(t *testing.T) {
	source := aeoracle.NewRenderRequest("source.aep", "Main", "source_out", []aeoracle.FrameTarget{
		{Frame: 0, Seconds: 0, Tag: "f000000", Reason: "comp_start"},
	})

	got := aeoracle.CloneRenderRequest(source, "clone.aep", "Clone Main", "clone_out")

	if got.CompName != "Clone Main" {
		t.Fatalf("CompName = %q, want override", got.CompName)
	}
}

func TestRenderRequestValidateRejectsRequiredFields(t *testing.T) {
	valid := aeoracle.NewRenderRequest("input.aep", "Main", t.TempDir(), []aeoracle.FrameTarget{{Frame: 0, Tag: "f000000"}})
	cases := []struct {
		name string
		edit func(*aeoracle.RenderRequest)
	}{
		{"schema", func(r *aeoracle.RenderRequest) { r.SchemaVersion = 99 }},
		{"aep", func(r *aeoracle.RenderRequest) { r.AEPPath = "" }},
		{"out", func(r *aeoracle.RenderRequest) { r.OutputDir = "" }},
		{"frames", func(r *aeoracle.RenderRequest) { r.Frames = nil }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := valid
			tc.edit(&req)
			if err := req.Validate(); err == nil {
				t.Fatal("Validate error = nil, want error")
			}
		})
	}
}
