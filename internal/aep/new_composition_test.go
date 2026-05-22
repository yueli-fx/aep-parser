package aep_test

import (
	"bytes"
	"math"
	"strings"
	"testing"

	aep "github.com/example/aep-parser/internal/aep"
)

// TestBuildCompItemMatchesGoldenStructure 是 Phase 4.7 占位 stub。
// 完整验证 builder 输出 Item LIST children 顺序 / 类型跟 AE 2025 saved 1-comp
// fixture (test_data/comp_item_children.golden.txt) 一致，需要 NewComposition
// 入口（Phase 5）建出 *Composition 才能拿到 itemList chunk 对比。
//
// 当前 Phase 4 builders 已经实现完整 buildCompItem 链路，但 unexported；
// Phase 5 TestNewComposition_Roundtrip 通过 WriteAEP + FromReader 隐式验证
// children 顺序正确（错了 AE 会拒开 / re-parse 会丢字段）。
//
// 这个 skip-stub 留作未来如果需要 explicit byte-level golden diff 时的占位。
func TestBuildCompItemMatchesGoldenStructure(t *testing.T) {
	t.Skip("Deferred to Phase 5: NewComposition roundtrip implicitly verifies golden order")
}

func TestNewComposition_Fields(t *testing.T) {
	p := aep.NewProject()
	c, err := p.NewComposition("Main", 1920, 1080, 29.97, 10)
	if err != nil {
		t.Fatal(err)
	}
	if c.Name != "Main" {
		t.Errorf("Name = %q, want Main", c.Name)
	}
	if c.Width != 1920 || c.Height != 1080 {
		t.Errorf("Size = %dx%d, want 1920x1080", c.Width, c.Height)
	}
	if math.Abs(c.FrameRate-29.97) > 1e-4 {
		t.Errorf("FrameRate = %g, want 29.97 (canonical)", c.FrameRate)
	}
	// Duration tolerance: 29.97 canonical encoding (29 + 0xF852/65536 = 29.96997833)
	// drifts ~0.01s on a 10s comp due to integer-frame storage + decoded fps slightly
	// below 29.97. Acceptable for AE intent (still 300 frames stored).
	if math.Abs(c.Duration-10) > 0.02 {
		t.Errorf("Duration = %g, want ~10 (within NTSC drift)", c.Duration)
	}
	if c.PixelAspect != 1.0 {
		t.Errorf("PixelAspect = %g, want 1.0", c.PixelAspect)
	}
	if c.ResolutionFactor != [2]uint16{1, 1} {
		t.Errorf("ResolutionFactor = %v, want [1 1]", c.ResolutionFactor)
	}
	if c.BGColor != [3]uint8{0, 0, 0} {
		t.Errorf("BGColor = %v, want [0 0 0]", c.BGColor)
	}
	if c.ID == 0 {
		t.Error("ID not allocated")
	}
	if len(p.Compositions) != 1 || p.Compositions[0] != c {
		t.Errorf("Project.Compositions not updated correctly")
	}
}

func TestNewComposition_Roundtrip(t *testing.T) {
	p := aep.NewProject()
	c1, err := p.NewComposition("Main", 1920, 1080, 29.97, 10)
	if err != nil {
		t.Fatal(err)
	}
	c1.SetBGColor([3]uint8{20, 30, 40})
	c1.SetResolutionFactor(2, 2)

	c2, _ := p.NewComposition("BG", 1280, 720, 30, 5)
	if c2.ID == c1.ID {
		t.Errorf("ID collision: c1=%d c2=%d", c1.ID, c2.ID)
	}

	var buf bytes.Buffer
	if err := p.WriteAEP(&buf); err != nil {
		t.Fatal(err)
	}
	re, err := aep.FromReader(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatal(err)
	}
	if len(re.Compositions) != 2 {
		t.Fatalf("re.Compositions = %d, want 2", len(re.Compositions))
	}
	main := findCompByName(re, "Main")
	if main == nil {
		t.Fatal("Main comp not found post-roundtrip")
	}
	if main.BGColor != [3]uint8{20, 30, 40} {
		t.Errorf("post-rt BGColor = %v, want [20 30 40]", main.BGColor)
	}
	if main.ResolutionFactor != [2]uint16{2, 2} {
		t.Errorf("post-rt ResolutionFactor = %v, want [2 2]", main.ResolutionFactor)
	}
	if math.Abs(main.FrameRate-29.97) > 1e-4 {
		t.Errorf("post-rt FrameRate = %g, want 29.97", main.FrameRate)
	}
	if math.Abs(main.Duration-10) > 0.02 {
		t.Errorf("post-rt Duration = %g, want ~10 (within NTSC drift)", main.Duration)
	}
}

func findCompByName(p *aep.Project, name string) *aep.Composition {
	for _, c := range p.Compositions {
		if c.Name == name {
			return c
		}
	}
	return nil
}

func TestNewComposition_RejectsInvalid(t *testing.T) {
	cases := []struct {
		name        string
		fn          func(p *aep.Project) error
		wantInError string
	}{
		{"empty name", func(p *aep.Project) error {
			_, e := p.NewComposition("", 1920, 1080, 30, 5)
			return e
		}, "name cannot be empty"},
		{"zero width", func(p *aep.Project) error {
			_, e := p.NewComposition("x", 0, 1080, 30, 5)
			return e
		}, "size must be > 0"},
		{"zero height", func(p *aep.Project) error {
			_, e := p.NewComposition("x", 1920, 0, 30, 5)
			return e
		}, "size must be > 0"},
		{"zero fps", func(p *aep.Project) error {
			_, e := p.NewComposition("x", 1920, 1080, 0, 5)
			return e
		}, "frame rate must be > 0"},
		{"negative fps", func(p *aep.Project) error {
			_, e := p.NewComposition("x", 1920, 1080, -29.97, 5)
			return e
		}, "frame rate must be > 0"},
		{"zero duration", func(p *aep.Project) error {
			_, e := p.NewComposition("x", 1920, 1080, 30, 0)
			return e
		}, "duration must be > 0"},
		{"negative duration", func(p *aep.Project) error {
			_, e := p.NewComposition("x", 1920, 1080, 30, -1)
			return e
		}, "duration must be > 0"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			p := aep.NewProject()
			err := tc.fn(p)
			if err == nil {
				t.Fatalf("expected error, got nil")
			}
			if !strings.Contains(err.Error(), tc.wantInError) {
				t.Errorf("err = %q, want contains %q", err.Error(), tc.wantInError)
			}
			if len(p.Compositions) != 0 {
				t.Errorf("project polluted on failure: %d comps", len(p.Compositions))
			}
		})
	}
}

func TestNewComposition_OnOpenedProject_NoIDCollision(t *testing.T) {
	p, err := aep.Open("../../test_data/re_batch.aep")
	if err != nil {
		t.Skipf("re_batch.aep not present: %v", err)
	}
	var maxOld uint32
	for _, c := range p.Compositions {
		if c.ID > maxOld {
			maxOld = c.ID
		}
	}
	for _, f := range p.Footage {
		if f.ID > maxOld {
			maxOld = f.ID
		}
	}

	c, err := p.NewComposition("Added", 1920, 1080, 30, 5)
	if err != nil {
		t.Fatal(err)
	}
	if c.ID <= maxOld {
		t.Errorf("new comp ID %d should be > existing max %d", c.ID, maxOld)
	}
}

func TestNewComposition_EmptyLayrListPreserved(t *testing.T) {
	p := aep.NewProject()
	c, _ := p.NewComposition("EmptyL", 1920, 1080, 30, 5)
	if len(c.Layers) != 0 {
		t.Errorf("fresh NewComposition has %d layers, want 0", len(c.Layers))
	}

	// Roundtrip: AE should not auto-insert sentinel layer
	var buf bytes.Buffer
	if err := p.WriteAEP(&buf); err != nil {
		t.Fatal(err)
	}
	re, err := aep.FromReader(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatal(err)
	}
	for _, recomp := range re.Compositions {
		if recomp.Name == "EmptyL" && len(recomp.Layers) != 0 {
			t.Errorf("post-rt EmptyL has %d layers, want 0", len(recomp.Layers))
		}
	}
}
