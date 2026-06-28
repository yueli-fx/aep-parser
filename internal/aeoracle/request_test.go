package aeoracle_test

import (
	"path/filepath"
	"testing"

	"github.com/example/aep-parser/internal/aeoracle"
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
	if len(got.Frames) != 2 || got.Frames[1].Tag != "f000030" {
		t.Fatalf("Frames = %+v", got.Frames)
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
