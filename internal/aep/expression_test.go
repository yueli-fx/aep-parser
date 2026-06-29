package aep_test

import (
	"bytes"
	"encoding/binary"
	"math"
	"testing"

	"github.com/yueli-fx/aep-parser/internal/aep"
)

// buildEffectParade wraps one effect (with one float parameter) into an
// "ADBE Effect Parade" tdgp. The wrapper LIST formType is sspc (matches what
// the real AE serializer emits).
func (b *rifxBuilder) buildEffectParade(effectMatchName string, paramName string, paramValue float64) []byte {
	param := b.leafStatic(paramName, 0x01, paramValue)
	// Inner tdgp holds the parameter tdmn+tdbs pair + Group End sentinel.
	var innerBody []byte
	innerBody = append(innerBody, param...)
	innerBody = append(innerBody, b.chunk("tdmn", []byte("ADBE Group End"))...)
	innerTdgp := b.listChunk("LIST", "tdgp", innerBody)
	// Wrapper LIST (sspc).
	sspcWrap := b.listChunk("LIST", "sspc", innerTdgp)
	// Parade tdgp: tdmn(effect name) + sspc wrapper + Group End.
	var paradeBody []byte
	paradeBody = append(paradeBody, b.chunk("tdmn", []byte(effectMatchName))...)
	paradeBody = append(paradeBody, sspcWrap...)
	paradeBody = append(paradeBody, b.chunk("tdmn", []byte("ADBE Group End"))...)
	paradeTdgp := b.listChunk("LIST", "tdgp", paradeBody)
	// Outer: tdmn("ADBE Effect Parade") + paradeTdgp.
	var out []byte
	out = append(out, b.chunk("tdmn", []byte("ADBE Effect Parade"))...)
	out = append(out, paradeTdgp...)
	return out
}

// buildTextSourceWrapper wraps a fake opaque btds LIST with the supplied
// payload bytes into an "ADBE Text Document" tdmn+btds pair.
func (b *rifxBuilder) buildTextSourceWrapper(payload []byte) []byte {
	// We need to assemble the LIST manually since rifxBuilder.listChunk
	// treats body as already-formed chunks (which an opaque payload isn't).
	var out bytes.Buffer
	out.WriteString("LIST")
	_ = binary.Write(&out, binary.BigEndian, uint32(4+len(payload)))
	out.WriteString("btds")
	out.Write(payload)
	if len(payload)%2 != 0 {
		out.WriteByte(0)
	}
	var ret []byte
	ret = append(ret, b.chunk("tdmn", []byte("ADBE Text Document"))...)
	ret = append(ret, out.Bytes()...)
	return ret
}

// buildExtendedAEP creates a synthetic .aep with one layer carrying:
//   - ADBE Opacity (1D keyframed) with expression "time*2"
//   - ADBE Effect Parade containing "ADBE Test Effect" with one param = 42
//   - ADBE Marker with 2 markers
//   - ADBE Text Document with opaque btds payload "HELLO-TEXT-PAYLOAD"
func buildExtendedAEP(textPayload []byte, markerTimes []float64, markerComments [][5]string) []byte {
	rb := &rifxBuilder{}

	opacity := rb.leafKeyframedWithExpr("ADBE Opacity", 0x01, 48, [][]byte{
		buildKF1D(1.0, 1.0),
		buildKF1D(2.0, 0.5),
	}, "time*2")

	effectParade := rb.buildEffectParade("ADBE Test Effect", "ADBE Test Effect-0001", 42.0)
	markers := rb.buildMarkers(markerTimes, markerComments)
	textWrap := rb.buildTextSourceWrapper(textPayload)

	var groupBody []byte
	groupBody = append(groupBody, opacity...)
	groupBody = append(groupBody, effectParade...)
	groupBody = append(groupBody, markers...)
	groupBody = append(groupBody, textWrap...)
	groupBody = append(groupBody, rb.chunk("tdmn", []byte("ADBE Group End"))...)
	tdgpList := rb.listChunk("LIST", "tdgp", groupBody)

	ldtaData := make([]byte, 0x2C)
	binary.BigEndian.PutUint32(ldtaData[0x28:], 7)
	ldta := rb.chunk("ldta", ldtaData)
	var layerBody []byte
	layerBody = append(layerBody, ldta...)
	layerBody = append(layerBody, tdgpList...)
	layrList := rb.listChunk("LIST", "Layr", layerBody)

	compName := rb.chunk("Utf8", []byte("Ext Comp"))
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

// TestEffectParamSetStaticValueRoundtrip exercises Property.SetStaticValue
// on an effect parameter (Layer.Effects[].Parameters[]), verifying that
// the generic Property setter works directly on effect knobs — no
// effect-specific API needed. Covers scalar (Gaussian Blur Blurriness)
// and 4D (Tritone color) cases.
func TestEffectParamSetStaticValueRoundtrip(t *testing.T) {
	proj, err := aep.Open("../../test_data/re_batch.aep")
	if err != nil {
		t.Skipf("re_batch.aep not present")
	}
	var blur, tritone *aep.Property
	for _, c := range proj.Compositions {
		for _, l := range c.Layers {
			for _, e := range l.Effects {
				for _, p := range e.Parameters {
					switch p.MatchName {
					case "ADBE Gaussian Blur 2-0001":
						if v, ok := p.StaticValue.(float64); ok && v == 88 {
							blur = p
						}
					case "ADBE Tritone-0002":
						if v, ok := p.StaticValue.([]float64); ok && len(v) == 4 {
							tritone = p
						}
					}
				}
			}
		}
	}
	if blur == nil {
		t.Fatal("ADBE Gaussian Blur 2-0001 static=88 not found in re_batch.aep")
	}
	if tritone == nil {
		t.Fatal("ADBE Tritone-0002 4D static not found in re_batch.aep")
	}
	if err := blur.SetStaticValue(42.0); err != nil {
		t.Fatalf("blur.SetStaticValue: %v", err)
	}
	if err := tritone.SetStaticValue([]float64{10, 20, 30, 40}); err != nil {
		t.Fatalf("tritone.SetStaticValue: %v", err)
	}

	// In-memory mirror updated.
	if got, _ := blur.StaticValue.(float64); got != 42.0 {
		t.Errorf("blur in-mem after set: %v, want 42.0", blur.StaticValue)
	}

	// Roundtrip.
	var buf bytes.Buffer
	if err := proj.WriteAEP(&buf); err != nil {
		t.Fatalf("WriteAEP: %v", err)
	}
	re, err := aep.FromReader(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("re-parse: %v", err)
	}
	var blur2, tritone2 *aep.Property
	for _, c := range re.Compositions {
		for _, l := range c.Layers {
			for _, e := range l.Effects {
				for _, p := range e.Parameters {
					switch p.MatchName {
					case "ADBE Gaussian Blur 2-0001":
						if v, ok := p.StaticValue.(float64); ok && (v == 42.0 || v == 88) {
							blur2 = p
						}
					case "ADBE Tritone-0002":
						if _, ok := p.StaticValue.([]float64); ok {
							tritone2 = p
						}
					}
				}
			}
		}
	}
	if blur2 == nil {
		t.Fatal("post-roundtrip blur param not found")
	}
	if got, _ := blur2.StaticValue.(float64); got != 42.0 {
		t.Errorf("post-roundtrip blur StaticValue = %v, want 42.0", blur2.StaticValue)
	}
	if tritone2 == nil {
		t.Fatal("post-roundtrip tritone param not found")
	}
	got, _ := tritone2.StaticValue.([]float64)
	want := []float64{10, 20, 30, 40}
	if len(got) != 4 || got[0] != want[0] || got[1] != want[1] || got[2] != want[2] || got[3] != want[3] {
		t.Errorf("post-roundtrip tritone StaticValue = %v, want %v", got, want)
	}
}

// TestEffectParamKeyframeSetValueRoundtrip exercises Keyframe.SetValue
// on a keyframed effect parameter (ADBE Tritone-0001, the 4D color),
// confirming the same path as direct-property keyframe edit works on
// effects too.
func TestEffectParamKeyframeSetValueRoundtrip(t *testing.T) {
	proj, err := aep.Open("../../test_data/re_batch.aep")
	if err != nil {
		t.Skipf("re_batch.aep not present")
	}
	var tritoneKfd *aep.Property
	for _, c := range proj.Compositions {
		for _, l := range c.Layers {
			for _, e := range l.Effects {
				for _, p := range e.Parameters {
					if p.MatchName == "ADBE Tritone-0001" && len(p.Keyframes) >= 2 {
						tritoneKfd = p
					}
				}
			}
		}
	}
	if tritoneKfd == nil {
		t.Fatal("keyframed ADBE Tritone-0001 not found in re_batch.aep")
	}
	want := []float64{1.5, 2.5, 3.5, 4.5}
	if err := tritoneKfd.Keyframes[0].SetValue(want); err != nil {
		t.Fatalf("Keyframes[0].SetValue: %v", err)
	}

	var buf bytes.Buffer
	if err := proj.WriteAEP(&buf); err != nil {
		t.Fatalf("WriteAEP: %v", err)
	}
	re, err := aep.FromReader(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("re-parse: %v", err)
	}
	for _, c := range re.Compositions {
		for _, l := range c.Layers {
			for _, e := range l.Effects {
				for _, p := range e.Parameters {
					if p.MatchName != "ADBE Tritone-0001" || len(p.Keyframes) == 0 {
						continue
					}
					got, _ := p.Keyframes[0].Value.([]float64)
					if len(got) != 4 || got[0] != want[0] || got[1] != want[1] || got[2] != want[2] || got[3] != want[3] {
						t.Errorf("post-roundtrip Tritone kf[0] = %v, want %v", got, want)
					}
					return
				}
			}
		}
	}
	t.Error("post-roundtrip Tritone keyframed param not found")
}

func TestExpressionAndEffectsAndMarkersAndText(t *testing.T) {
	textPayload := []byte("HELLO-TEXT-PAYLOAD-OPAQUE")
	markerTimes := []float64{1.5, 3.0}
	markerComments := [][5]string{
		{"first", "chap-A", "https://example.com", "ft-1", "cue-X"},
		{"second", "", "", "", ""},
	}
	data := buildExtendedAEP(textPayload, markerTimes, markerComments)

	proj, err := aep.FromReader(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("FromReader: %v", err)
	}
	if len(proj.Compositions) != 1 || len(proj.Compositions[0].Layers) != 1 {
		t.Fatalf("expected 1 comp/1 layer")
	}
	layer := proj.Compositions[0].Layers[0]

	// Expression
	op := layer.Opacity()
	if op == nil {
		t.Fatal("Opacity() = nil")
	}
	if op.Expression != "time*2" {
		t.Errorf("Opacity.Expression = %q, want %q", op.Expression, "time*2")
	}
	if len(op.Keyframes) != 2 {
		t.Errorf("Opacity keyframes = %d, want 2", len(op.Keyframes))
	}

	// Effects
	if len(layer.Effects) != 1 {
		t.Fatalf("Effects = %d, want 1", len(layer.Effects))
	}
	fx := layer.Effects[0]
	if fx.MatchName != "ADBE Test Effect" {
		t.Errorf("Effect.MatchName = %q, want ADBE Test Effect", fx.MatchName)
	}
	if len(fx.Parameters) != 1 {
		t.Fatalf("effect params = %d, want 1", len(fx.Parameters))
	}
	if got, want := fx.Parameters[0].StaticValue.(float64), 42.0; got != want {
		t.Errorf("effect param value = %v, want %v", got, want)
	}

	// Markers
	if len(layer.Markers) != 2 {
		t.Fatalf("Markers = %d, want 2", len(layer.Markers))
	}
	if got, want := layer.Markers[0].Time, 1.5; math.Abs(got-want) > 1e-9 {
		t.Errorf("Marker[0].Time = %v, want %v", got, want)
	}
	if layer.Markers[0].Comment != "first" {
		t.Errorf("Marker[0].Comment = %q, want %q", layer.Markers[0].Comment, "first")
	}
	if layer.Markers[0].Chapter != "chap-A" {
		t.Errorf("Marker[0].Chapter = %q, want chap-A", layer.Markers[0].Chapter)
	}
	if layer.Markers[0].URL != "https://example.com" {
		t.Errorf("Marker[0].URL = %q", layer.Markers[0].URL)
	}
	if layer.Markers[0].FrameTarget != "ft-1" {
		t.Errorf("Marker[0].FrameTarget = %q", layer.Markers[0].FrameTarget)
	}
	if layer.Markers[0].CuePointName != "cue-X" {
		t.Errorf("Marker[0].CuePointName = %q", layer.Markers[0].CuePointName)
	}
	if layer.Markers[1].Comment != "second" {
		t.Errorf("Marker[1].Comment = %q", layer.Markers[1].Comment)
	}

	// Text source
	if !bytes.Equal(layer.TextSourceRaw, textPayload) {
		t.Errorf("TextSourceRaw mismatch: got %q want %q", layer.TextSourceRaw, textPayload)
	}
	if layer.Type != aep.LayerTypeText {
		t.Errorf("Layer.Type = %q, want %q (text detected via btds)", layer.Type, aep.LayerTypeText)
	}

	// Bogus "ADBE Marker" property should NOT appear in layer.Properties
	// (it's diverted to Markers instead).
	for _, p := range layer.Properties {
		if p.MatchName == "ADBE Marker" {
			t.Error("ADBE Marker should be diverted, not appear in Properties")
		}
	}

	// Byte-perfect roundtrip — exercises opaque-LIST write path for btds.
	var buf bytes.Buffer
	if err := proj.WriteAEP(&buf); err != nil {
		t.Fatalf("WriteAEP: %v", err)
	}
	if !bytes.Equal(buf.Bytes(), data) {
		t.Errorf("roundtrip mismatch: in=%d out=%d bytes", len(data), buf.Len())
		minLen := len(data)
		if buf.Len() < minLen {
			minLen = buf.Len()
		}
		for i := 0; i < minLen; i++ {
			if data[i] != buf.Bytes()[i] {
				t.Logf("first diff @0x%X: in=0x%02X out=0x%02X", i, data[i], buf.Bytes()[i])
				break
			}
		}
	}
}
