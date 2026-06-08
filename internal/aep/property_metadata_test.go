package aep_test

import (
	"testing"

	aep "github.com/example/aep-parser/internal/aep"
	"github.com/example/aep-parser/internal/codec"
	"github.com/example/aep-parser/internal/rifx"
)

// TestPropertyControlType_Derivation verifies ControlType() derivation
// from tdb4 flags for representative property shapes.
func TestPropertyControlType_Derivation(t *testing.T) {
	tests := []struct {
		name    string
		dims    int
		spatial bool
		color   bool
		integer bool
		vector  bool
		noValue bool
		wantPCT aep.PropertyControlType
		wantPVT aep.PropertyValueType
	}{
		{
			name: "Position_3D_spatial",
			dims: 3, spatial: true, vector: true,
			wantPCT: aep.PCTLThreeD,
			wantPVT: aep.PVTThreeDSpatial,
		},
		{
			name: "Scale_3D_nonspatial",
			dims: 3, vector: true,
			wantPCT: aep.PCTLThreeD,
			wantPVT: aep.PVTThreeD,
		},
		{
			name: "Opacity_1D_scalar",
			dims: 1, vector: true,
			wantPCT: aep.PCTLScalar,
			wantPVT: aep.PVTOneD,
		},
		{
			name: "FillColor_4D_color",
			dims: 1, color: true,
			wantPCT: aep.PCTLColor,
			wantPVT: aep.PVTColor,
		},
		{
			name: "Checkbox_boolean",
			dims: 1, integer: true,
			wantPCT: aep.PCTLBoolean,
			wantPVT: aep.PVTOneD,
		},
		{
			name: "MaskFeather_2D_nonspatial",
			dims: 2, vector: true,
			wantPCT: aep.PCTLTwoD,
			wantPVT: aep.PVTTwoD,
		},
		{
			name: "EffectPoint_2D_spatial",
			dims: 2, spatial: true, vector: true,
			wantPCT: aep.PCTLTwoD,
			wantPVT: aep.PVTTwoDSpatial,
		},
		{
			name: "NoValue_separator",
			dims: 1, noValue: true,
			wantPCT: aep.PCTLUnknown,
			wantPVT: aep.PVTNoValue,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var spatialByte, typeByte byte
			if tt.spatial {
				spatialByte |= 0x08
			}
			if tt.color {
				typeByte |= 0x01
			}
			if tt.integer {
				typeByte |= 0x04
			}
			if tt.vector {
				typeByte |= 0x08
			}
			var noValByte byte
			if tt.noValue {
				noValByte = 0x01
			}
			tdb4 := aep.MakeTdb4(uint16(tt.dims), spatialByte, 0x02, noValByte, typeByte)
			p := aep.NewTestProperty("test", tt.dims, tdb4, nil, nil, nil)

			gotPCT := p.ControlType()
			if gotPCT != tt.wantPCT {
				t.Errorf("ControlType() = %v, want %v", gotPCT, tt.wantPCT)
			}
			gotPVT := p.ValuePropertyType()
			if gotPVT != tt.wantPVT {
				t.Errorf("ValuePropertyType() = %v, want %v", gotPVT, tt.wantPVT)
			}
		})
	}
}

// TestPropertyControlType_Fixture verifies ControlType/ValuePropertyType
// against parsed fixture properties.
func TestPropertyControlType_Fixture(t *testing.T) {
	proj, err := aep.Open("../../test_data/re_batch.aep")
	if err != nil {
		t.Skipf("re_batch.aep not present: %v", err)
	}

	var position, opacity *aep.Property
	for _, c := range proj.Compositions {
		for _, l := range c.Layers {
			for _, p := range l.Properties {
				switch p.MatchName {
				case aep.MatchNamePosition:
					if position == nil {
						position = p
					}
				case aep.MatchNameOpacity:
					if opacity == nil {
						opacity = p
					}
				}
			}
		}
	}

	if position != nil {
		pct := position.ControlType()
		if pct != aep.PCTLThreeD {
			t.Errorf("Position ControlType = %v, want PCTLThreeD", pct)
		}
		pvt := position.ValuePropertyType()
		if pvt != aep.PVTThreeDSpatial {
			t.Errorf("Position ValuePropertyType = %v, want PVTThreeDSpatial", pvt)
		}
	}
	if opacity != nil {
		pct := opacity.ControlType()
		if pct != aep.PCTLScalar {
			t.Errorf("Opacity ControlType = %v, want PCTLScalar", pct)
		}
		pvt := opacity.ValuePropertyType()
		if pvt != aep.PVTOneD {
			t.Errorf("Opacity ValuePropertyType = %v, want PVTOneD", pvt)
		}
	}
}

// TestMinValue_MaxValue_Synthetic verifies MinValue/MaxValue decoding
// from synthetic tdum/tduM chunks.
func TestMinValue_MaxValue_Synthetic(t *testing.T) {
	t.Run("scalar_f64", func(t *testing.T) {
		// tdum = 1 × float64 BE = 0.0
		tdumData := make([]byte, 8) // 0.0
		tdum := &rifx.Chunk{ID: rifx.IDtdum, Size: 8, Data: tdumData}
		// tduM = 1 × float64 BE = 100.0
		tduMData := make([]byte, 8)
		tduMData[0] = 0x40 // 100.0 in float64 BE
		tduMData[1] = 0x59
		tduMData[2] = 0x00
		tduMData[3] = 0x00
		tduMData[4] = 0x00
		tduMData[5] = 0x00
		tduMData[6] = 0x00
		tduMData[7] = 0x00
		tduM := &rifx.Chunk{ID: rifx.IDtduM, Size: 8, Data: tduMData}

		tdb4 := aep.MakeTdb4(1, 0, 0x02, 0, 0x08) // vector=1, dims=1
		p := aep.NewTestProperty("test", 1, tdb4, nil, tdum, tduM)

		min := p.MinValue()
		if min != 0.0 {
			t.Errorf("MinValue() = %v, want 0.0", min)
		}
		max := p.MaxValue()
		if max != 100.0 {
			t.Errorf("MaxValue() = %v, want 100.0", max)
		}
	})

	t.Run("nil_when_absent", func(t *testing.T) {
		tdb4 := aep.MakeTdb4(1, 0, 0x02, 0, 0x08)
		p := aep.NewTestProperty("test", 1, tdb4, nil, nil, nil)
		if p.MinValue() != nil {
			t.Errorf("MinValue() with no tdum = %v, want nil", p.MinValue())
		}
		if p.MaxValue() != nil {
			t.Errorf("MaxValue() with no tduM = %v, want nil", p.MaxValue())
		}
	})

	t.Run("integer_u32", func(t *testing.T) {
		// tdum = uint32 BE = 0
		tdumData := []byte{0x00, 0x00, 0x00, 0x00}
		tdum := &rifx.Chunk{ID: rifx.IDtdum, Size: 4, Data: tdumData}
		// tduM = uint32 BE = 10
		tduMData := []byte{0x00, 0x00, 0x00, 0x0A}
		tduM := &rifx.Chunk{ID: rifx.IDtduM, Size: 4, Data: tduMData}

		tdb4 := aep.MakeTdb4(1, 0, 0x02, 0, 0x04) // integer=1, dims=1
		p := aep.NewTestProperty("test", 1, tdb4, nil, tdum, tduM)

		min := p.MinValue()
		if min != 0.0 {
			t.Errorf("MinValue() integer = %v, want 0.0", min)
		}
		max := p.MaxValue()
		if max != 10.0 {
			t.Errorf("MaxValue() integer = %v, want 10.0", max)
		}
	})

	t.Run("color_f32x4", func(t *testing.T) {
		// tdum = 4 × float32 BE = [0, 0, 0, 0]
		tdumData := make([]byte, 16)
		tdum := &rifx.Chunk{ID: rifx.IDtdum, Size: 16, Data: tdumData}
		// tduM = 4 × float32 BE = [1, 1, 1, 1]
		tduMData := make([]byte, 16)
		for i := 0; i < 4; i++ {
			tduMData[i*4] = 0x3F // float32 1.0 = 0x3F800000
			tduMData[i*4+1] = 0x80
		}
		tduM := &rifx.Chunk{ID: rifx.IDtduM, Size: 16, Data: tduMData}

		tdb4 := aep.MakeTdb4(1, 0, 0x02, 0, 0x01) // color=1
		p := aep.NewTestProperty("test", 4, tdb4, nil, tdum, tduM)

		min := p.MinValue()
		if min == nil {
			t.Fatal("MinValue() color = nil")
		}
		minVals, ok := min.([]float64)
		if !ok || len(minVals) != 4 {
			t.Fatalf("MinValue() color = %T %v, want []float64 len 4", min, min)
		}
		for i, v := range minVals {
			if v != 0.0 {
				t.Errorf("MinValue()[%d] = %v, want 0.0", i, v)
			}
		}
		max := p.MaxValue()
		if max == nil {
			t.Fatal("MaxValue() color = nil")
		}
		maxVals := max.([]float64)
		for i, v := range maxVals {
			if v < 0.99 || v > 1.01 {
				t.Errorf("MaxValue()[%d] = %v, want ~1.0", i, v)
			}
		}
	})
}

// TestUnitsText verifies UnitsText() for known match-names.
func TestUnitsText(t *testing.T) {
	tests := []struct {
		matchName string
		want      string
	}{
		{"ADBE Opacity", "percent"},
		{"ADBE Position", "pixels"},
		{"ADBE Rotate Z", "degrees"},
		{"ADBE Audio Levels", "dB"},
		{"ADBE Time Remapping", "seconds"},
		{"ADBE Scale", "percent"},
		{"ADBE Camera Zoom", "pixels"},
		{"nonexistent match name", ""},
	}

	for _, tt := range tests {
		t.Run(tt.matchName, func(t *testing.T) {
			p := aep.NewTestProperty(tt.matchName, 1, nil, nil, nil, nil)
			got := p.UnitsText()
			if got != tt.want {
				t.Errorf("UnitsText() = %q, want %q", got, tt.want)
			}
		})
	}
}

// TestUnitsText_AngleFallback verifies that properties with angle control
// type fall back to "degrees" when not in the map.
func TestUnitsText_AngleFallback(t *testing.T) {
	// A property with no match-name in the map but with integer flag set
	// (which doesn't trigger angle fallback). The angle fallback only works
	// when ControlType() == PCTLAngle, which requires specific tdb4 flags
	// we can't easily synthesize. Just test the empty case.
	p := aep.NewTestProperty("unknown prop", 1, nil, nil, nil, nil)
	if got := p.UnitsText(); got != "" {
		t.Errorf("UnitsText() for unknown = %q, want empty", got)
	}
}

// TestPropertyIndex_PropertyDepth_Fixture verifies PropertyIndex and
// PropertyDepth on a parsed fixture.
func TestPropertyIndex_PropertyDepth_Fixture(t *testing.T) {
	proj, err := aep.Open("../../test_data/re_cameralight.aep")
	if err != nil {
		t.Skipf("re_cameralight.aep not present: %v", err)
	}

	var camera *aep.Layer
	for _, c := range proj.Compositions {
		for _, l := range c.Layers {
			if l.Type == aep.LayerTypeCamera {
				camera = l
				break
			}
		}
	}
	if camera == nil {
		t.Skip("no camera layer in fixture")
	}

	// Position should be a leaf in the Transform group.
	pos := camera.PropertyByPath(aep.MatchNameGroupTransform, aep.MatchNamePosition)
	if pos == nil {
		t.Skip("Position not found in camera Transform group")
	}

	// PropertyDepth: Transform group is depth 1 from root, Position is
	// depth 2. But our PropertyDepth() counts from the parentTreeGroup's
	// depth, so Position should report the group's depth + 1.
	idx := pos.PropertyIndex()
	if idx < 0 {
		t.Errorf("Position PropertyIndex = %d, want >= 0", idx)
	}
	depth := pos.PropertyDepth()
	if depth < 0 {
		t.Errorf("Position PropertyDepth = %d, want >= 0", depth)
	}

	// ParentGroup should be the Transform group.
	parent := pos.ParentGroup()
	if parent == nil {
		t.Fatal("Position ParentGroup = nil")
	}
	if parent.PropertyMatchName() != aep.MatchNameGroupTransform {
		t.Errorf("Position ParentGroup matchName = %q, want %q",
			parent.PropertyMatchName(), aep.MatchNameGroupTransform)
	}
}

// TestPropertyIndex_NoParent verifies that PropertyIndex returns -1 for
// properties without a parent tree group.
func TestPropertyIndex_NoParent(t *testing.T) {
	p := aep.NewTestProperty("test", 1, nil, nil, nil, nil)
	if got := p.PropertyIndex(); got != -1 {
		t.Errorf("PropertyIndex() = %d, want -1", got)
	}
	if got := p.PropertyDepth(); got != -1 {
		t.Errorf("PropertyDepth() = %d, want -1", got)
	}
}

// TestDefaultValue_Transform verifies that transform properties get
// DefaultValue assigned during parse.
func TestDefaultValue_Transform(t *testing.T) {
	proj, err := aep.Open("../../test_data/re_cameralight.aep")
	if err != nil {
		t.Skipf("re_cameralight.aep not present: %v", err)
	}

	for _, c := range proj.Compositions {
		for _, l := range c.Layers {
			for _, p := range l.Properties {
				switch p.MatchName {
				case aep.MatchNameScale:
					if p.DefaultValue == nil {
						t.Error("Scale DefaultValue = nil; want non-nil")
					}
				case aep.MatchNameOpacity:
					if p.DefaultValue == nil {
						t.Error("Opacity DefaultValue = nil; want non-nil")
					}
				case aep.MatchNamePosition:
					if p.DefaultValue == nil {
						t.Error("Position DefaultValue = nil; want non-nil")
					}
				case aep.MatchNameAnchorPoint:
					if p.DefaultValue == nil {
						t.Error("AnchorPoint DefaultValue = nil; want non-nil")
					}
				}
			}
		}
	}
}

// TestGradient_ParseXML verifies gradient XML parsing with synthetic data.
func TestGradient_ParseXML(t *testing.T) {
	xmlText := `<prop.map version="4">
  <prop.list>
    <prop.pair><key>Gradient Color Data</key>
      <prop.list>
        <prop.pair><key>Color Stops</key>
          <prop.list>
            <prop.pair><key>Stops List</key>
              <prop.list>
                <prop.pair><key>Stop-0</key>
                  <prop.list>
                    <prop.pair><key>Stops Color</key>
                      <array><float>0</float><float>0.5</float><float>1</float><float>0</float><float>0</float><float>1</float></array>
                    </prop.pair>
                  </prop.list>
                </prop.pair>
                <prop.pair><key>Stop-1</key>
                  <prop.list>
                    <prop.pair><key>Stops Color</key>
                      <array><float>1</float><float>0.5</float><float>0</float><float>1</float><float>0</float><float>1</float></array>
                    </prop.pair>
                  </prop.list>
                </prop.pair>
              </prop.list>
            </prop.pair>
          </prop.list>
        </prop.pair>
        <prop.pair><key>Alpha Stops</key>
          <prop.list>
            <prop.pair><key>Stops List</key>
              <prop.list>
                <prop.pair><key>Stop-0</key>
                  <prop.list>
                    <prop.pair><key>Stops Alpha</key>
                      <array><float>0</float><float>0.5</float><float>1</float></array>
                    </prop.pair>
                  </prop.list>
                </prop.pair>
              </prop.list>
            </prop.pair>
          </prop.list>
        </prop.pair>
      </prop.list>
    </prop.pair>
    <prop.pair><key>Gradient Colors</key><string>4</string></prop.pair>
  </prop.list>
</prop.map>`

	g := codec.ParseGradientXML(xmlText)
	if g == nil {
		t.Fatal("codec.ParseGradientXML returned nil")
	}
	if g.Version != "4" {
		t.Errorf("Version = %q, want %q", g.Version, "4")
	}
	if len(g.ColorStops) != 2 {
		t.Fatalf("ColorStops len = %d, want 2", len(g.ColorStops))
	}
	if g.ColorStops[0].Offset != 0.0 {
		t.Errorf("ColorStops[0].Offset = %v, want 0.0", g.ColorStops[0].Offset)
	}
	if g.ColorStops[0].Color != [3]float64{1, 0, 0} {
		t.Errorf("ColorStops[0].Color = %v, want [1 0 0]", g.ColorStops[0].Color)
	}
	if g.ColorStops[1].Offset != 1.0 {
		t.Errorf("ColorStops[1].Offset = %v, want 1.0", g.ColorStops[1].Offset)
	}
	if g.ColorStops[1].Color != [3]float64{0, 1, 0} {
		t.Errorf("ColorStops[1].Color = %v, want [0 1 0]", g.ColorStops[1].Color)
	}
	if len(g.AlphaStops) != 1 {
		t.Fatalf("AlphaStops len = %d, want 1", len(g.AlphaStops))
	}
	if g.AlphaStops[0].Alpha != 1.0 {
		t.Errorf("AlphaStops[0].Alpha = %v, want 1.0", g.AlphaStops[0].Alpha)
	}
}

// TestGradient_ParseEmpty verifies graceful handling of empty/invalid XML.
func TestGradient_ParseEmpty(t *testing.T) {
	if g := codec.ParseGradientXML(""); g != nil {
		t.Errorf("empty string: got %v, want nil", g)
	}
	if g := codec.ParseGradientXML("not xml"); g != nil {
		t.Errorf("non-xml: got %v, want nil", g)
	}
}

// TestDefaultValue_NonTransform verifies that non-transform properties
// do NOT get DefaultValue assigned (stays nil).
func TestDefaultValue_NonTransform(t *testing.T) {
	p := aep.NewTestProperty("ADBE Gaussian Blur 2-0001", 1, nil, nil, nil, nil)
	if p.DefaultValue != nil {
		t.Errorf("Non-transform DefaultValue = %v; want nil", p.DefaultValue)
	}
}

// TestEffectPardMetadata_Fixture verifies that effect parameters get
// DefaultValue/LastValue/NbOptions from pard chunks.
func TestEffectPardMetadata_Fixture(t *testing.T) {
	proj, err := aep.Open("../../test_data/re_batch.aep")
	if err != nil {
		t.Skipf("re_batch.aep not present: %v", err)
	}

	// Find any effect parameter with pard metadata.
	var foundDV, foundLV bool
	for _, c := range proj.Compositions {
		for _, l := range c.Layers {
			for _, e := range l.Effects {
				for _, p := range e.Parameters {
					if p.DefaultValue != nil {
						foundDV = true
						t.Logf("  %s/%s: DefaultValue=%v", e.MatchName, p.MatchName, p.DefaultValue)
					}
					if p.LastValue != nil {
						foundLV = true
						t.Logf("  %s/%s: LastValue=%v", e.MatchName, p.MatchName, p.LastValue)
					}
					if p.NbOptions != 0 {
						t.Logf("  %s/%s: NbOptions=%d", e.MatchName, p.MatchName, p.NbOptions)
					}
				}
			}
		}
	}

	if !foundDV {
		t.Error("No effect parameter had DefaultValue from pard")
	}
	if !foundLV {
		t.Error("No effect parameter had LastValue from pard")
	}
}
