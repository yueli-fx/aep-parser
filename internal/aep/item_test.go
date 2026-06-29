package aep_test

import (
	"bytes"
	"encoding/binary"
	"testing"

	"github.com/yueli-fx/aep-parser/internal/aep"
)

func TestSetFootagePath(t *testing.T) {
	// Build a synthetic file with a Pin/Als2/alas JSON containing fullpath.
	rb := &rifxBuilder{}

	aliasJSON := []byte(`{"fullpath":"D:\\old\\file.png","other":"keep"}`)
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
	_ = binary.Write(&root, binary.BigEndian, uint32(4+len(foldList)))
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
	if f.Path != `D:\old\file.png` {
		t.Errorf("initial path = %q, want D:\\old\\file.png", f.Path)
	}

	newPath := `E:\much\longer\new\path\to\renamed-asset.png`
	if err := f.SetPath(newPath); err != nil {
		t.Fatalf("SetPath: %v", err)
	}
	if f.Path != newPath {
		t.Errorf("after SetPath, Path = %q, want %q", f.Path, newPath)
	}
	if f.Name != "renamed-asset.png" {
		t.Errorf("after SetPath, Name = %q, want renamed-asset.png", f.Name)
	}

	var out bytes.Buffer
	if err := proj.WriteAEP(&out); err != nil {
		t.Fatalf("WriteAEP: %v", err)
	}
	proj2, err := aep.FromReader(bytes.NewReader(out.Bytes()))
	if err != nil {
		t.Fatalf("re-parse after SetPath: %v", err)
	}
	if len(proj2.Footage) != 1 {
		t.Fatalf("re-parsed footage count = %d, want 1", len(proj2.Footage))
	}
	if proj2.Footage[0].Path != newPath {
		t.Errorf("re-parsed path = %q, want %q", proj2.Footage[0].Path, newPath)
	}
	if !bytes.Contains(out.Bytes(), []byte(`"other":"keep"`)) {
		t.Error(`sibling JSON field "other":"keep" was lost`)
	}
}

func TestProjectIDs(t *testing.T) {
	data := buildMinimalAEP()
	proj, err := aep.FromReader(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("FromReader failed: %v", err)
	}

	if len(proj.Compositions) > 0 && proj.Compositions[0].ID != 1 {
		t.Errorf("comp ID = %d, want 1", proj.Compositions[0].ID)
	}
	if len(proj.Footage) > 0 && proj.Footage[0].ID != 2 {
		t.Errorf("footage ID = %d, want 2", proj.Footage[0].ID)
	}
	if len(proj.Folders) > 0 && proj.Folders[0].ID != 3 {
		t.Errorf("folder ID = %d, want 3", proj.Folders[0].ID)
	}
}

// TestProjectSetBitsPerChannel verifies the dual-write to nhed @0x0F
// and nnhd @0x18 against the AE 2020 fixture re_bpc_*.aep set.
// Skips when fixtures aren't present.
func TestProjectSetBitsPerChannel(t *testing.T) {
	cases := []struct {
		file string
		bpc  aep.BitsPerChannel
	}{
		{"../../test_data/fixtures/re_bpc_8.aep", aep.BPC8},
		{"../../test_data/fixtures/re_bpc_16.aep", aep.BPC16},
		{"../../test_data/fixtures/re_bpc_32.aep", aep.BPC32},
	}
	for _, tc := range cases {
		proj, err := aep.Open(tc.file)
		if err != nil {
			t.Skipf("%s not present", tc.file)
			return
		}
		if proj.BitsPerChannel != tc.bpc {
			t.Errorf("%s: read BitsPerChannel = %d, want %d", tc.file, proj.BitsPerChannel, tc.bpc)
		}
	}

	proj, err := aep.Open(cases[0].file)
	if err != nil {
		t.Skipf("%s not present", cases[0].file)
		return
	}
	if err := proj.SetBitsPerChannel(aep.BPC32); err != nil {
		t.Fatalf("SetBitsPerChannel: %v", err)
	}
	if proj.BitsPerChannel != aep.BPC32 {
		t.Errorf("in-mem after Set: %d, want %d", proj.BitsPerChannel, aep.BPC32)
	}
	var buf bytes.Buffer
	if err := proj.WriteAEP(&buf); err != nil {
		t.Fatalf("WriteAEP: %v", err)
	}
	proj2, err := aep.FromReader(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("re-parse: %v", err)
	}
	if proj2.BitsPerChannel != aep.BPC32 {
		t.Errorf("after roundtrip: BitsPerChannel = %d", proj2.BitsPerChannel)
	}
}

// TestItemCommentAndLabel covers Item-level SetComment / SetLabel
// against re_batch3.aep (which has real 84-byte idta chunks).
// The synthetic buildMinimalAEP's 20-byte idta is too short for
// the @0x3A label byte write, hence the real-file fixture.
func TestItemCommentAndLabel(t *testing.T) {
	proj, err := aep.Open("../../test_data/fixtures/re_batch3.aep")
	if err != nil {
		t.Skipf("re_batch3.aep not present")
	}
	var comp *aep.Composition
	for _, c := range proj.Compositions {
		if c.Name == "RE_B3_J_commentLabel" {
			comp = c
			break
		}
	}
	if comp == nil {
		t.Fatal("RE_B3_J_commentLabel comp not found")
	}
	if err := comp.SetComment("rewritten via setter\nline two"); err != nil {
		t.Fatalf("comp SetComment: %v", err)
	}
	if err := comp.SetLabel(11); err != nil {
		t.Fatalf("comp SetLabel: %v", err)
	}
	if comp.Comment != "rewritten via setter\nline two" {
		t.Errorf("in-mem Comment = %q", comp.Comment)
	}
	if comp.Label != 11 {
		t.Errorf("in-mem Label = %d", comp.Label)
	}

	var buf bytes.Buffer
	if err := proj.WriteAEP(&buf); err != nil {
		t.Fatalf("WriteAEP: %v", err)
	}
	proj2, err := aep.FromReader(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("re-parse: %v", err)
	}
	for _, c := range proj2.Compositions {
		if c.Name != "RE_B3_J_commentLabel" {
			continue
		}
		if c.Comment != "rewritten via setter\nline two" {
			t.Errorf("roundtrip Comment = %q", c.Comment)
		}
		if c.Label != 11 {
			t.Errorf("roundtrip Label = %d", c.Label)
		}
		return
	}
	t.Error("comp missing after roundtrip")
}
