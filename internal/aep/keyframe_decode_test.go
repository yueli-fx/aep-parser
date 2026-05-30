package aep_test

import (
	"bytes"
	"encoding/binary"
	"math"
	"testing"

	aep "github.com/example/aep-parser/internal/aep"
)

// buildKeyframedAEP wraps three test properties (Opacity 1D + Position 3D
// spatial + Glossiness Coefficient 1D static) into a full RIFX file.
// Shared scaffolding for the keyframe roundtrip + insert/delete tests.
func buildKeyframedAEP() []byte {
	rb := &rifxBuilder{}

	// One Opacity keyframe pair: t=1.0s v=1.0, t=3.0s v=0.5.
	opacity := rb.leafKeyframed("ADBE Opacity", 0x01, 48, [][]byte{
		buildKF1D(1.0, 1.0),
		buildKF1D(3.0, 0.5),
	})
	// One Position keyframe pair (spatial): t=0.0s v=[100,200,0], t=2.0s v=[300,400,0].
	position := rb.leafKeyframed("ADBE Position", 0x03, 128, [][]byte{
		buildKF3DSpatial(0.0, 100, 200, 0),
		buildKF3DSpatial(2.0, 300, 400, 0),
	})
	// One static Glossiness Coefficient.
	gloss := rb.leafStatic("ADBE Glossiness Coefficient", 0x01, 80.0)

	// Property group: tdmn entries interleaved with payload, terminated by
	// "ADBE Group End" tdmn.
	var groupBody []byte
	groupBody = append(groupBody, opacity...)
	groupBody = append(groupBody, position...)
	groupBody = append(groupBody, gloss...)
	groupBody = append(groupBody, rb.chunk("tdmn", []byte("ADBE Group End"))...)
	tdgpList := rb.listChunk("LIST", "tdgp", groupBody)

	// Layer: ldta (minimal — SourceID at 0x28) + property tdgp.
	ldtaData := make([]byte, 0x2C)
	binary.BigEndian.PutUint32(ldtaData[0x28:], 99) // SourceID
	ldta := rb.chunk("ldta", ldtaData)
	var layerBody []byte
	layerBody = append(layerBody, ldta...)
	layerBody = append(layerBody, tdgpList...)
	layrList := rb.listChunk("LIST", "Layr", layerBody)

	// Comp Item: Utf8 + idta + cdta + Layr.
	compName := rb.chunk("Utf8", []byte("KF Comp"))
	compIdta := rb.chunk("idta", buildIdta(0x04, 1))
	compCdta := rb.chunk("cdta", buildCdta(1920, 1080, 30, 0, 300))
	var compBody []byte
	compBody = append(compBody, compName...)
	compBody = append(compBody, compIdta...)
	compBody = append(compBody, compCdta...)
	compBody = append(compBody, layrList...)
	compItem := rb.listChunk("LIST", "Item", compBody)

	foldList := rb.listChunk("LIST", "Fold", compItem)

	var root bytes.Buffer
	root.WriteString("RIFX")
	_ = binary.Write(&root, binary.BigEndian, uint32(4+len(foldList)))
	root.WriteString("Egg!")
	root.Write(foldList)
	return root.Bytes()
}

// TestKeyframeSetValueUpdatesInMemory pins the regression for the 1D
// SetValue path: bytes AND k.Value must both be updated. Earlier the
// scalar case returned early after writing bytes, leaving k.Value stale.
func TestKeyframeSetValueUpdatesInMemory(t *testing.T) {
	rb := &rifxBuilder{}
	op := rb.leafKeyframed("ADBE Opacity", 0x01, 48, [][]byte{
		buildKF1D(0.0, 1.0),
	})
	var tdgpBody []byte
	tdgpBody = append(tdgpBody, op...)
	tdgpBody = append(tdgpBody, rb.chunk("tdmn", []byte("ADBE Group End"))...)
	data := wrapAsLayer(tdgpBody)
	proj, err := aep.FromReader(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("FromReader: %v", err)
	}
	kf := proj.Compositions[0].Layers[0].Opacity().Keyframes[0]
	if err := kf.SetValue(0.42); err != nil {
		t.Fatalf("SetValue: %v", err)
	}
	if got := kf.Value.(float64); got != 0.42 {
		t.Errorf("after SetValue, k.Value = %v, want 0.42 (in-memory state diverged from bytes)", got)
	}
	var out bytes.Buffer
	if err := proj.WriteAEP(&out); err != nil {
		t.Fatalf("WriteAEP: %v", err)
	}
	proj2, _ := aep.FromReader(bytes.NewReader(out.Bytes()))
	got2 := proj2.Compositions[0].Layers[0].Opacity().Keyframes[0].Value.(float64)
	if got2 != 0.42 {
		t.Errorf("after roundtrip, k.Value = %v, want 0.42", got2)
	}
}

func TestKeyframeRoundtrip(t *testing.T) {
	data := buildKeyframedAEP()
	proj, err := aep.FromReader(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("FromReader: %v", err)
	}
	if len(proj.Compositions) != 1 || len(proj.Compositions[0].Layers) != 1 {
		t.Fatalf("expected 1 comp/1 layer, got %d/%d",
			len(proj.Compositions),
			func() int {
				if len(proj.Compositions) == 0 {
					return 0
				}
				return len(proj.Compositions[0].Layers)
			}())
	}
	layer := proj.Compositions[0].Layers[0]

	props := map[string]*aep.Property{}
	for _, p := range layer.Properties {
		props[p.MatchName] = p
	}
	for _, name := range []string{"ADBE Opacity", "ADBE Position", "ADBE Glossiness Coefficient"} {
		if props[name] == nil {
			t.Fatalf("missing property %q on layer", name)
		}
	}

	op := props["ADBE Opacity"]
	if op.Components != 1 {
		t.Errorf("Opacity Components = %d, want 1", op.Components)
	}
	if len(op.Keyframes) != 2 {
		t.Fatalf("Opacity keyframes = %d, want 2", len(op.Keyframes))
	}
	if got, want := op.Keyframes[0].Time, 1.0; math.Abs(got-want) > 1e-9 {
		t.Errorf("Opacity kf[0].Time = %v, want %v", got, want)
	}
	if got, want := op.Keyframes[0].Value.(float64), 1.0; got != want {
		t.Errorf("Opacity kf[0].Value = %v, want %v", got, want)
	}
	if got, want := op.Keyframes[1].Value.(float64), 0.5; got != want {
		t.Errorf("Opacity kf[1].Value = %v, want %v", got, want)
	}

	pos := props["ADBE Position"]
	if pos.Components != 3 {
		t.Errorf("Position Components = %d, want 3", pos.Components)
	}
	wantPos0 := []float64{100, 200, 0}
	wantPos1 := []float64{300, 400, 0}
	gotPos0 := pos.Keyframes[0].Value.([]float64)
	gotPos1 := pos.Keyframes[1].Value.([]float64)
	for i := range wantPos0 {
		if gotPos0[i] != wantPos0[i] {
			t.Errorf("Position kf[0][%d] = %v, want %v", i, gotPos0[i], wantPos0[i])
		}
		if gotPos1[i] != wantPos1[i] {
			t.Errorf("Position kf[1][%d] = %v, want %v", i, gotPos1[i], wantPos1[i])
		}
	}

	gloss := props["ADBE Glossiness Coefficient"]
	if len(gloss.Keyframes) != 0 {
		t.Errorf("Glossiness should have no keyframes, got %d", len(gloss.Keyframes))
	}
	if got, want := gloss.StaticValue.(float64), 80.0; got != want {
		t.Errorf("Glossiness static = %v, want %v", got, want)
	}

	if err := op.Keyframes[0].SetTime(2.5); err != nil {
		t.Fatalf("Opacity SetTime: %v", err)
	}
	if err := op.Keyframes[1].SetValue(0.25); err != nil {
		t.Fatalf("Opacity SetValue: %v", err)
	}
	if err := pos.Keyframes[1].SetValue([]float64{500, 600, 0}); err != nil {
		t.Fatalf("Position SetValue: %v", err)
	}
	if err := gloss.SetStaticValue(42.0); err != nil {
		t.Fatalf("Glossiness SetStaticValue: %v", err)
	}

	if err := pos.Keyframes[0].SetValue(123.0); err == nil {
		t.Error("expected error setting scalar on 3D property")
	}
	if err := op.Keyframes[0].SetValue([]float64{1, 2, 3}); err == nil {
		t.Error("expected error setting 3D slice on 1D property")
	}
	if err := op.SetStaticValue(0.0); err == nil {
		t.Error("expected error calling SetStaticValue on keyframed property")
	}

	var out bytes.Buffer
	if err := proj.WriteAEP(&out); err != nil {
		t.Fatalf("WriteAEP: %v", err)
	}
	if out.Len() != len(data) {
		t.Errorf("roundtrip length changed: in=%d out=%d (expected byte-preserving)", len(data), out.Len())
	}

	proj2, err := aep.FromReader(bytes.NewReader(out.Bytes()))
	if err != nil {
		t.Fatalf("re-parse: %v", err)
	}
	layer2 := proj2.Compositions[0].Layers[0]
	props2 := map[string]*aep.Property{}
	for _, p := range layer2.Properties {
		props2[p.MatchName] = p
	}
	if got, want := props2["ADBE Opacity"].Keyframes[0].Time, 2.5; math.Abs(got-want) > 1e-9 {
		t.Errorf("after roundtrip Opacity kf[0].Time = %v, want 2.5", got)
	}
	if got, want := props2["ADBE Opacity"].Keyframes[1].Value.(float64), 0.25; got != want {
		t.Errorf("after roundtrip Opacity kf[1].Value = %v, want 0.25", got)
	}
	gotPos := props2["ADBE Position"].Keyframes[1].Value.([]float64)
	wantPos := []float64{500, 600, 0}
	for i := range wantPos {
		if gotPos[i] != wantPos[i] {
			t.Errorf("after roundtrip Position kf[1][%d] = %v, want %v", i, gotPos[i], wantPos[i])
		}
	}
	if got, want := props2["ADBE Glossiness Coefficient"].StaticValue.(float64), 42.0; got != want {
		t.Errorf("after roundtrip Glossiness static = %v, want 42", got)
	}
}
