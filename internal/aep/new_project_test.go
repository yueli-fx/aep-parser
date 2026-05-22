// internal/aep/new_project_test.go
package aep_test

import (
	"bytes"
	"encoding/hex"
	"strings"
	"testing"

	aep "github.com/example/aep-parser/internal/aep"
	"github.com/example/aep-parser/internal/rifx"
)

// svap expected bytes — frozen from template files. 防 template 被悄悄替换。
// 来源: hex dump of internal/aep/templates/{2020,2022,2025}.aep svap chunk
const (
	svapExpectedAE2020 = "0b0b862d"
	svapExpectedAE2022 = "0b330640"
	svapExpectedAE2025 = "0f088644"
)

func TestNewProject_DefaultTargetAE2020(t *testing.T) {
	p := aep.NewProject()
	if len(p.Compositions) != 0 {
		t.Errorf("Compositions = %d, want 0", len(p.Compositions))
	}
	if len(p.Footage) != 0 {
		t.Errorf("Footage = %d, want 0", len(p.Footage))
	}
	if len(p.Folders) != 0 {
		t.Errorf("Folders = %d, want 0", len(p.Folders))
	}
	// AE2020 template was authored at 32bpc; AE2022/2025 templates are 8bpc.
	// Frozen template artifact — don't re-author to match a tidier default.
	if p.BitsPerChannel != aep.BPC32 {
		t.Errorf("BitsPerChannel = %v, want BPC32 (frozen template value)", p.BitsPerChannel)
	}
	if len(p.Warnings) != 0 {
		t.Errorf("template produced %d warnings: %v", len(p.Warnings), p.Warnings)
	}
}

func TestNewProject_AllTargetsParseAndSvap(t *testing.T) {
	cases := []struct {
		target     aep.AETarget
		wantSvap   string
		wantChunkN int // root Egg! 直接 children 数（24 for 2020, 30 for 2022/2025）
	}{
		{aep.TargetAE2020, svapExpectedAE2020, 24},
		{aep.TargetAE2022, svapExpectedAE2022, 30},
		{aep.TargetAE2025, svapExpectedAE2025, 30},
	}
	for _, tc := range cases {
		t.Run(targetName(tc.target), func(t *testing.T) {
			p := aep.NewProject(tc.target)
			if len(p.Warnings) != 0 {
				t.Errorf("template AE %d produced %d warnings: %v", int(tc.target), len(p.Warnings), p.Warnings)
			}

			// roundtrip 重 serialize 看 svap byte 保留
			var buf bytes.Buffer
			if err := p.WriteAEP(&buf); err != nil {
				t.Fatalf("WriteAEP: %v", err)
			}
			root, err := rifx.Parse(bytes.NewReader(buf.Bytes()))
			if err != nil {
				t.Fatalf("re-parse: %v", err)
			}
			if len(root.Children) != tc.wantChunkN {
				t.Errorf("AE %d: root children = %d, want %d", int(tc.target), len(root.Children), tc.wantChunkN)
			}
			for _, c := range root.Children {
				if string(c.ID[:]) == "svap" {
					got := hex.EncodeToString(c.Data)
					if !strings.EqualFold(got, tc.wantSvap) {
						t.Errorf("AE %d svap = %s, want %s", int(tc.target), got, tc.wantSvap)
					}
					return
				}
			}
			t.Errorf("AE %d: svap chunk not found in output", int(tc.target))
		})
	}
}

func TestNewProject_RejectsMultipleTargets(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
			t.Error("expected panic, got none")
			return
		}
		msg, ok := r.(string)
		if !ok || !strings.Contains(msg, "accepts at most one target") {
			t.Errorf("panic msg = %v, want contains 'accepts at most one target'", r)
		}
	}()
	aep.NewProject(aep.TargetAE2020, aep.TargetAE2025)
}

func TestNewProject_RejectsUnknownTarget(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
			t.Error("expected panic, got none")
			return
		}
		msg, ok := r.(string)
		if !ok || !strings.Contains(msg, "forward-incompat") {
			t.Errorf("panic msg = %v, want contains 'forward-incompat'", r)
		}
	}()
	aep.NewProject(aep.AETarget(9999))
}

func TestNewProject_IndependentInstances(t *testing.T) {
	p1 := aep.NewProject()
	p2 := aep.NewProject()
	// mutating p1 should not affect p2's Warnings slice
	p1.Warnings = append(p1.Warnings, "fake")
	if len(p2.Warnings) != 0 {
		t.Errorf("p2.Warnings polluted by p1 mutation: %v", p2.Warnings)
	}
}

func targetName(t aep.AETarget) string {
	switch t {
	case aep.TargetAE2020:
		return "AE2020"
	case aep.TargetAE2022:
		return "AE2022"
	case aep.TargetAE2025:
		return "AE2025"
	}
	return "unknown"
}
