package sliceworkflow_test

import (
	"reflect"
	"testing"

	"github.com/example/aep-parser/internal/aeoracle"
	"github.com/example/aep-parser/internal/profile"
	"github.com/example/aep-parser/internal/profilediff"
	"github.com/example/aep-parser/internal/sliceworkflow"
)

func TestBuildReportClassifiesProceduralComp(t *testing.T) {
	prof := testProfile("text.aep", testComp(10, "TextSlice",
		textLayer(1, "Title"),
		shapeLayer(2, "Underline"),
	))

	report, err := sliceworkflow.BuildReport(prof, sliceworkflow.Options{MaxSlices: 1})
	if err != nil {
		t.Fatal(err)
	}
	if report.Mode != "plan" {
		t.Fatalf("Mode = %q, want plan", report.Mode)
	}
	if report.SourceProject != "text.aep" {
		t.Fatalf("SourceProject = %q, want text.aep", report.SourceProject)
	}
	if len(report.Slices) != 1 {
		t.Fatalf("len(Slices) = %d, want 1", len(report.Slices))
	}
	got := report.Slices[0]
	if got.Classification != sliceworkflow.ClassificationProcedural {
		t.Fatalf("Classification = %q, want %q", got.Classification, sliceworkflow.ClassificationProcedural)
	}
	if got.TextLayerCount != 1 || got.ShapeLayerCount != 1 || got.FootageLayerCount != 0 || got.PluginDependencyCount != 0 {
		t.Fatalf("slice counts = %+v", got)
	}
	assertContains(t, got.Reasons, "text-layer")
	assertContains(t, got.Reasons, "shape-layer")
	if len(report.Commands) == 0 {
		t.Fatal("Commands empty")
	}
}

func TestBuildReportSelectsDeterministicRepresentativeSlices(t *testing.T) {
	prof := testProfile("mixed.aep",
		testComp(1, "Simple", textLayer(1, "Text")),
		testComp(2, "Footage", footageLayer(1, "Plate")),
		testComp(3, "Plugin", pluginLayer(1, "Magic")),
		testComp(4, "Other", shapeLayer(1, "Box")),
	)

	report, err := sliceworkflow.BuildReport(prof, sliceworkflow.Options{MaxSlices: 3})
	if err != nil {
		t.Fatal(err)
	}
	var ids []uint32
	for _, slice := range report.Slices {
		ids = append(ids, slice.CompID)
	}
	want := []uint32{3, 2, 1}
	if !reflect.DeepEqual(ids, want) {
		t.Fatalf("selected comp IDs = %+v, want %+v", ids, want)
	}
	if report.Slices[0].Classification != sliceworkflow.ClassificationPluginDependent {
		t.Fatalf("top classification = %q", report.Slices[0].Classification)
	}
	if report.Slices[1].Classification != sliceworkflow.ClassificationFootageAssembly {
		t.Fatalf("second classification = %q", report.Slices[1].Classification)
	}
}

func TestBuildDiagnoseReportIncludesDiffAndRenderGaps(t *testing.T) {
	source := testProfile("expected.aep", testComp(1, "Main", textLayer(1, "Title")))
	observed := testProfile("actual.aep", testComp(1, "Main", shapeLayer(1, "Title")))
	diffReport := &profilediff.Report{
		SchemaVersion: profilediff.SchemaVersion,
		DiffCount:     1,
		Diffs: []profilediff.Diff{{
			Path:       "comps.by_id[1].layers.by_id[1].type",
			Kind:       profilediff.KindWrongValue,
			Severity:   profilediff.SeverityFidelity,
			ActionType: profilediff.ActionWrite,
			Expected:   "text",
			Actual:     "shape",
			Evidence: profile.Evidence{
				Level:      profile.EvidenceL1Parsed,
				Source:     "internal/profilediff",
				Confidence: "high",
			},
		}},
	}
	renderReport := aeoracle.CompareReport{
		SchemaVersion:    aeoracle.SchemaVersion,
		ExpectedPath:     "expected.png",
		ActualPath:       "actual.png",
		Width:            2,
		Height:           2,
		TotalPixels:      4,
		DifferentPixels:  1,
		DifferentPercent: 25,
		MaxChannelDelta:  3,
	}

	report, err := sliceworkflow.BuildDiagnoseReport(source, observed, diffReport, []aeoracle.CompareReport{renderReport}, nil, sliceworkflow.Options{})
	if err != nil {
		t.Fatal(err)
	}
	if report.Mode != "diagnose" {
		t.Fatalf("Mode = %q, want diagnose", report.Mode)
	}
	if report.SourceProject != "expected.aep" || report.ObservedIn != "actual.aep" {
		t.Fatalf("context = %q/%q", report.SourceProject, report.ObservedIn)
	}
	if len(report.GapReports) != 2 {
		t.Fatalf("len(GapReports) = %d, want 2", len(report.GapReports))
	}
	if report.Summary.GapCount != 2 {
		t.Fatalf("Summary.GapCount = %d, want 2", report.Summary.GapCount)
	}
}

func TestBuildDiagnoseReportIncludesFrameSetGaps(t *testing.T) {
	source := testProfile("expected.aep", testComp(1, "Main", textLayer(1, "Title")))
	observed := testProfile("actual.aep", testComp(1, "Main", textLayer(1, "Title")))
	frameSet := aeoracle.FrameSetCompareReport{
		SchemaVersion: aeoracle.SchemaVersion,
		Frames: []aeoracle.FrameCompareRecord{{
			Tag:          "f000024",
			Frame:        24,
			Seconds:      1,
			ExpectedPath: "expected.png",
			ActualPath:   "actual.png",
			Status:       aeoracle.FrameStatusDifferent,
			Compare: &aeoracle.CompareReport{
				SchemaVersion:   aeoracle.SchemaVersion,
				TotalPixels:     4,
				DifferentPixels: 1,
			},
		}},
	}

	report, err := sliceworkflow.BuildDiagnoseReport(source, observed, nil, nil, []aeoracle.FrameSetCompareReport{frameSet}, sliceworkflow.Options{})
	if err != nil {
		t.Fatal(err)
	}
	if len(report.GapReports) != 1 {
		t.Fatalf("len(GapReports) = %d, want 1", len(report.GapReports))
	}
	if report.Summary.GapCount != 1 {
		t.Fatalf("Summary.GapCount = %d, want 1", report.Summary.GapCount)
	}
}

func testProfile(path string, comps ...profile.Composition) *profile.Profile {
	var layerCount int
	for _, comp := range comps {
		layerCount += len(comp.Layers)
	}
	return &profile.Profile{
		SchemaVersion: profile.SchemaVersion,
		Meta:          profile.Meta{Path: path},
		Fingerprint: profile.Fingerprint{
			CompCount:  len(comps),
			LayerCount: layerCount,
		},
		Comps: comps,
	}
}

func testComp(id uint32, name string, layers ...profile.Layer) profile.Composition {
	return profile.Composition{
		ID:     id,
		Name:   name,
		Layers: layers,
		Path: profile.PathRef{
			Path:        "comps.by_id[" + itoa(id) + "]",
			DisplayPath: "comps[" + name + "]",
		},
		Evidence: parsedEvidence(),
	}
}

func textLayer(id uint32, name string) profile.Layer {
	layer := baseLayer(id, name, "text")
	layer.Text = &profile.TextSource{Text: name, Evidence: parsedEvidence()}
	return layer
}

func shapeLayer(id uint32, name string) profile.Layer {
	layer := baseLayer(id, name, "shape")
	layer.Shapes = []profile.Shape{{Kind: "rect", Evidence: parsedEvidence()}}
	return layer
}

func footageLayer(id uint32, name string) profile.Layer {
	layer := baseLayer(id, name, "footage")
	layer.SourceRef = &profile.ItemRef{ID: 99, Kind: "footage", Name: "plate.png"}
	return layer
}

func pluginLayer(id uint32, name string) profile.Layer {
	layer := baseLayer(id, name, "solid")
	layer.Effects = []profile.Effect{{
		MatchName:       "Third Party Magic",
		DependencyClass: "third_party",
		Evidence:        parsedEvidence(),
	}}
	return layer
}

func baseLayer(id uint32, name, typ string) profile.Layer {
	return profile.Layer{
		ID:    id,
		Index: int(id),
		Name:  name,
		Type:  typ,
		Path: profile.PathRef{
			Path:        "layers.by_id[" + itoa(id) + "]",
			DisplayPath: "layers[" + name + "]",
		},
		Evidence: parsedEvidence(),
	}
}

func parsedEvidence() profile.Evidence {
	return profile.Evidence{Level: profile.EvidenceL1Parsed, Source: "test", Confidence: "high"}
}

func assertContains(t *testing.T, values []string, want string) {
	t.Helper()
	for _, value := range values {
		if value == want {
			return
		}
	}
	t.Fatalf("%q not found in %+v", want, values)
}

func itoa(v uint32) string {
	if v == 0 {
		return "0"
	}
	var buf [10]byte
	i := len(buf)
	for v > 0 {
		i--
		buf[i] = byte('0' + v%10)
		v /= 10
	}
	return string(buf[i:])
}
