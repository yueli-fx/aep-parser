// internal/aep/solid_setters_test.go
//
// Go round-trip tests for Footage.SetSolidColor / SetSolidSize (no AE). The
// byte patches are the ones the ship-gated NewSolidLayer create path applies
// (opti @0x0A ARGB f32, sspc @0x20/@0x24 u16) — these tests prove the
// standalone setters hit the same slots and survive WriteAEP → re-parse.
package aep_test

import (
	"bytes"
	"encoding/binary"
	"testing"

	"github.com/yueli-fx/aep-parser/internal/aep"
)

// solidFixture builds a project with one solid layer and returns the
// re-parsed project, comp, and the solid's backing footage item.
func solidFixture(t *testing.T) (*aep.Project, *aep.Footage) {
	t.Helper()
	p := aep.NewProject()
	comp, err := aep.NewComposition(p, "Main", 1280, 720, 30, 5)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := aep.NewSolidLayer(comp, "BG", 1280, 720, [3]float64{1, 0, 0}); err != nil {
		t.Fatalf("NewSolidLayer: %v", err)
	}
	p, err = aep.Reopen(p)
	if err != nil {
		t.Fatalf("Reopen: %v", err)
	}
	comp = p.CompositionByName("Main")
	if comp == nil || len(comp.Layers) == 0 {
		t.Fatal("re-parse: comp/layer missing")
	}
	var solid *aep.Footage
	for _, f := range p.Footage {
		if f.ID == comp.Layers[0].SourceID {
			solid = f
		}
	}
	if solid == nil || !solid.IsSolid {
		t.Fatal("re-parse: backing solid footage missing")
	}
	return p, solid
}

func TestSetSolidColor(t *testing.T) {
	p, solid := solidFixture(t)

	if err := solid.SetSolidColor([3]float64{0.25, 0.5, 0.75}); err != nil {
		t.Fatalf("SetSolidColor: %v", err)
	}
	if solid.SolidColor != [3]float64{0.25, 0.5, 0.75} {
		t.Errorf("scene SolidColor = %v", solid.SolidColor)
	}

	var buf bytes.Buffer
	if err := p.WriteAEP(&buf); err != nil {
		t.Fatalf("WriteAEP: %v", err)
	}
	re, err := aep.FromReader(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("FromReader: %v", err)
	}
	var rf *aep.Footage
	for _, f := range re.Footage {
		if f.ID == solid.ID {
			rf = f
		}
	}
	if rf == nil {
		t.Fatal("re-parse: solid footage missing")
	}
	for i, want := range [3]float64{0.25, 0.5, 0.75} {
		got := rf.SolidColor[i]
		if got < want-1e-6 || got > want+1e-6 {
			t.Errorf("re-parse SolidColor[%d] = %v, want %v", i, got, want)
		}
	}

	// Refusals: out-of-range channel leaves state untouched.
	if err := solid.SetSolidColor([3]float64{1.5, 0, 0}); err == nil {
		t.Error("out-of-range channel accepted, want refusal")
	}
	if solid.SolidColor != [3]float64{0.25, 0.5, 0.75} {
		t.Errorf("refusal mutated SolidColor: %v", solid.SolidColor)
	}
}

func TestSetSolidSize(t *testing.T) {
	p, solid := solidFixture(t)

	if err := solid.SetSolidSize(640, 360); err != nil {
		t.Fatalf("SetSolidSize: %v", err)
	}
	if solid.Width != 640 || solid.Height != 360 {
		t.Errorf("scene dims = %dx%d, want 640x360", solid.Width, solid.Height)
	}

	var buf bytes.Buffer
	if err := p.WriteAEP(&buf); err != nil {
		t.Fatalf("WriteAEP: %v", err)
	}
	re, err := aep.FromReader(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("FromReader: %v", err)
	}
	var rf *aep.Footage
	for _, f := range re.Footage {
		if f.ID == solid.ID {
			rf = f
		}
	}
	if rf == nil {
		t.Fatal("re-parse: solid footage missing")
	}
	if rf.Width != 640 || rf.Height != 360 {
		t.Errorf("re-parse dims = %dx%d, want 640x360", rf.Width, rf.Height)
	}

	if err := solid.SetSolidSize(0, 100); err == nil {
		t.Error("zero width accepted, want refusal")
	}
	if err := solid.SetSolidSize(100, 30001); err == nil {
		t.Error("oversize height accepted, want refusal")
	}
}

// Non-solid footage refuses both setters (synthetic file-backed footage,
// same builder as TestSetFootagePath).
func TestSolidSetters_NonSolidRefused(t *testing.T) {
	rb := &rifxBuilder{}
	aliasJSON := []byte(`{"fullpath":"D:\\old\\file.png"}`)
	alasChunk := rb.chunk("alas", aliasJSON)
	als2List := rb.listChunk("LIST", "Als2", alasChunk)
	pinList := rb.listChunk("LIST", "Pin ", als2List)
	footIdta := rb.chunk("idta", buildIdta(0x07, 42))
	var footData []byte
	footData = append(footData, footIdta...)
	footData = append(footData, pinList...)
	footItem := rb.listChunk("LIST", "Item", footData)
	foldList := rb.listChunk("LIST", "Fold", footItem)

	var root bytes.Buffer
	root.WriteString("RIFX")
	if err := binary.Write(&root, binary.BigEndian, uint32(4+len(foldList))); err != nil {
		t.Fatal(err)
	}
	root.WriteString("Egg!")
	root.Write(foldList)

	proj, err := aep.FromReader(bytes.NewReader(root.Bytes()))
	if err != nil {
		t.Fatalf("FromReader: %v", err)
	}
	if len(proj.Footage) != 1 {
		t.Fatalf("expected 1 footage, got %d", len(proj.Footage))
	}
	f := proj.Footage[0]
	if f.IsSolid {
		t.Fatal("synthetic footage unexpectedly parsed as solid")
	}
	if err := f.SetSolidColor([3]float64{0, 0, 0}); err == nil {
		t.Error("SetSolidColor on non-solid accepted, want refusal")
	}
	if err := f.SetSolidSize(100, 100); err == nil {
		t.Error("SetSolidSize on non-solid accepted, want refusal")
	}
}
