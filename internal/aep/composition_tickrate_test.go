package aep_test

import (
	"bytes"
	"encoding/binary"
	"math"
	"testing"

	aep "github.com/example/aep-parser/internal/aep"
)

// buildAEPWithCustomCdta wraps a single keyframed Opacity property into a
// comp with custom cdta tick-rate bytes. Returns the bytes ready for
// aep.FromReader.
func buildAEPWithCustomCdta(cdta []byte, kfTime float64, kfTickRate float64) []byte {
	rb := &rifxBuilder{}
	// Build a 1D Opacity keyframe at the requested time, where the time
	// will be encoded as ticks at kfTickRate.
	kf := make([]byte, 48)
	binary.BigEndian.PutUint32(kf[0:], uint32(math.Round(kfTime*kfTickRate)))
	binary.BigEndian.PutUint64(kf[0x08:], math.Float64bits(0.5))
	op := rb.leafKeyframed("ADBE Opacity", 0x01, 48, [][]byte{kf})
	var tdgpBody []byte
	tdgpBody = append(tdgpBody, op...)
	tdgpBody = append(tdgpBody, rb.chunk("tdmn", []byte("ADBE Group End"))...)
	tdgpList := rb.listChunk("LIST", "tdgp", tdgpBody)

	ldtaData := make([]byte, 0x2C)
	binary.BigEndian.PutUint32(ldtaData[0x28:], 1)
	ldta := rb.chunk("ldta", ldtaData)
	var layerBody []byte
	layerBody = append(layerBody, ldta...)
	layerBody = append(layerBody, tdgpList...)
	layrList := rb.listChunk("LIST", "Layr", layerBody)

	compName := rb.chunk("Utf8", []byte("Test"))
	compIdta := rb.chunk("idta", buildIdta(0x04, 1))
	compCdta := rb.chunk("cdta", cdta)
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

func TestPerCompTickRate(t *testing.T) {
	cases := []struct {
		name        string
		fpsWhole    uint16
		fpsFrac     uint16
		cdtaTick    uint32 // cdta_0x08
		legacyScale uint32 // cdta_0xA8
		wantRate    float64
	}{
		{"modern_30fps", 30, 0, 30720, 1, 30720},
		{"modern_24fps", 24, 0, 24576, 1, 24576},
		{"modern_29_97", 29, 63570, 23976, 1, 23976},
		{"legacy_29_97", 29, 63570, 23976, 2997, 8000},
		{"missing_scale_falls_back", 30, 0, 8000, 0, 8000},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cdta := buildCdtaWithRate(1920, 1080, tc.fpsWhole, tc.fpsFrac, 100,
				tc.cdtaTick, tc.legacyScale)
			// Keyframe at user-intent t=2.0s, encoded at the comp's tick rate
			// so the parser MUST derive the rate correctly to read t=2.0s
			// back out.
			data := buildAEPWithCustomCdta(cdta, 2.0, tc.wantRate)
			proj, err := aep.FromReader(bytes.NewReader(data))
			if err != nil {
				t.Fatalf("FromReader: %v", err)
			}
			if len(proj.Compositions) != 1 {
				t.Fatalf("got %d comps", len(proj.Compositions))
			}
			comp := proj.Compositions[0]
			if comp.TickRate != tc.wantRate {
				t.Errorf("TickRate = %v, want %v", comp.TickRate, tc.wantRate)
			}
			if len(comp.Layers) != 1 || comp.Layers[0].Opacity() == nil {
				t.Fatalf("layer/opacity missing")
			}
			kf := comp.Layers[0].Opacity().Keyframes[0]
			if math.Abs(kf.Time-2.0) > 1e-3 {
				t.Errorf("kf.Time = %.4f, want ~2.0 (rate=%v)", kf.Time, tc.wantRate)
			}

			// Roundtrip: SetTime to 3.5s, write, reparse, verify.
			if err := kf.SetTime(3.5); err != nil {
				t.Fatalf("SetTime: %v", err)
			}
			var out bytes.Buffer
			if err := proj.WriteAEP(&out); err != nil {
				t.Fatalf("WriteAEP: %v", err)
			}
			proj2, err := aep.FromReader(bytes.NewReader(out.Bytes()))
			if err != nil {
				t.Fatalf("FromReader: %v", err)
			}
			kf2 := proj2.Compositions[0].Layers[0].Opacity().Keyframes[0]
			if math.Abs(kf2.Time-3.5) > 1e-3 {
				t.Errorf("after SetTime, kf2.Time = %.4f, want 3.5", kf2.Time)
			}
		})
	}
}
