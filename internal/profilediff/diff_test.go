package profilediff_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/example/aep-parser/internal/profile"
	"github.com/example/aep-parser/internal/profilediff"
)

func TestCompareReportsLayerPresenceAndScalarDiffs(t *testing.T) {
	expected := testProfile(
		testLayer(10, "Hero", func(l *profile.Layer) {
			l.Flags.Visible = true
			l.Timing.InPoint = 0
			l.Timing.OutPoint = 3
		}),
		testLayer(11, "Missing", nil),
	)
	actual := testProfile(
		testLayer(10, "Hero", func(l *profile.Layer) {
			l.Flags.Visible = false
			l.Timing.InPoint = 0.5
			l.Timing.OutPoint = 3
		}),
		testLayer(12, "Extra", nil),
	)

	report, err := profilediff.Compare(expected, actual, profilediff.Options{})
	if err != nil {
		t.Fatalf("Compare: %v", err)
	}
	if report.SchemaVersion != profilediff.SchemaVersion {
		t.Fatalf("SchemaVersion = %d, want %d", report.SchemaVersion, profilediff.SchemaVersion)
	}
	assertDiff(t, report, "comps.by_id[1].layers.by_id[11]", profilediff.KindMissingObject)
	assertDiff(t, report, "comps.by_id[1].layers.by_id[12]", profilediff.KindExtraObject)
	assertDiff(t, report, "comps.by_id[1].layers.by_id[10].flags.visible", profilediff.KindWrongValue)
	assertDiff(t, report, "comps.by_id[1].layers.by_id[10].timing.in_point_seconds", profilediff.KindWrongValue)
}

func TestCompareReportsEffectParamAndKeyframeDiffs(t *testing.T) {
	expected := testProfile(testLayer(10, "Hero", func(l *profile.Layer) {
		l.Effects = []profile.Effect{testEffect(0, "ADBE Fill", []profile.Property{
			testProperty(`comps.by_id[1].layers.by_id[10].effects.by_match_name["ADBE Fill"]#0.params.by_match_name["ADBE Fill-0002"]#0`, "ADBE Fill-0002", "#ff0000", []profile.Keyframe{
				{Time: 0, Value: 0},
				{Time: 1, Value: 100},
			}),
		})}
	}))
	actual := testProfile(testLayer(10, "Hero", func(l *profile.Layer) {
		l.Effects = []profile.Effect{testEffect(0, "ADBE Fill", []profile.Property{
			testProperty(`comps.by_id[1].layers.by_id[10].effects.by_match_name["ADBE Fill"]#0.params.by_match_name["ADBE Fill-0002"]#0`, "ADBE Fill-0002", "#00ff00", []profile.Keyframe{
				{Time: 0, Value: 0},
			}),
		})}
	}))

	report, err := profilediff.Compare(expected, actual, profilediff.Options{})
	if err != nil {
		t.Fatalf("Compare: %v", err)
	}
	assertDiff(t, report, `comps.by_id[1].layers.by_id[10].effects.by_match_name["ADBE Fill"]#0.params.by_match_name["ADBE Fill-0002"]#0.static_value`, profilediff.KindWrongValue)
	assertDiff(t, report, `comps.by_id[1].layers.by_id[10].effects.by_match_name["ADBE Fill"]#0.params.by_match_name["ADBE Fill-0002"]#0.keyframes`, profilediff.KindWrongValue)
}

func TestCompareMatchesCloneObjectsByNameAndIndexWhenIDsDiffer(t *testing.T) {
	expected := testProfile(testLayer(10, "Hero", func(l *profile.Layer) {
		l.Timing.OutPoint = 3
	}))
	actual := &profile.Profile{
		SchemaVersion: profile.SchemaVersion,
		Fingerprint:   profile.Fingerprint{CompCount: 1, LayerCount: 1},
		Comps: []profile.Composition{{
			ID:       99,
			Name:     "Main",
			Path:     profile.PathRef{Path: "comps.by_id[99]", DisplayPath: `comps["Main"]`},
			Evidence: profile.Evidence{Level: profile.EvidenceL1Parsed, Source: "test", Confidence: "high"},
			Layers: []profile.Layer{testLayer(88, "Hero", func(l *profile.Layer) {
				l.Index = 10
				l.Path = profile.PathRef{Path: "comps.by_id[99].layers.by_id[88]"}
				l.Timing.OutPoint = 4
			})},
		}},
	}

	report, err := profilediff.Compare(expected, actual, profilediff.Options{})
	if err != nil {
		t.Fatalf("Compare: %v", err)
	}
	assertDiff(t, report, "comps.by_id[1].layers.by_id[10].timing.out_point_seconds", profilediff.KindWrongValue)
	assertNoDiff(t, report, "comps.by_id[1].layers.by_id[10]", profilediff.KindMissingObject)
	assertNoDiff(t, report, "comps.by_id[99].layers.by_id[88]", profilediff.KindExtraObject)
}

func TestCompareFallsBackToUniqueLayerIndexWhenGeneratedNamesDiffer(t *testing.T) {
	expected := testProfile(testLayer(10, "", func(l *profile.Layer) {
		l.Index = 3
		l.Timing.OutPoint = 3
	}))
	actual := testProfile(testLayer(88, "L03", func(l *profile.Layer) {
		l.Index = 3
		l.Path = profile.PathRef{Path: "comps.by_id[1].layers.by_id[88]"}
		l.Timing.OutPoint = 4
	}))

	report, err := profilediff.Compare(expected, actual, profilediff.Options{})
	if err != nil {
		t.Fatalf("Compare: %v", err)
	}
	assertDiff(t, report, "comps.by_id[1].layers.by_id[10].timing.out_point_seconds", profilediff.KindWrongValue)
	assertNoDiff(t, report, "comps.by_id[1].layers.by_id[10]", profilediff.KindMissingObject)
	assertNoDiff(t, report, "comps.by_id[1].layers.by_id[88]", profilediff.KindExtraObject)
}

func TestCompareSourceRefsByKindAndNameWhenIDsDiffer(t *testing.T) {
	expected := testProfile(testLayer(10, "Precomp", func(l *profile.Layer) {
		l.SourceRef = &profile.ItemRef{ID: 27, Kind: "composition", Name: "Child"}
	}))
	actual := testProfile(testLayer(10, "Precomp", func(l *profile.Layer) {
		l.SourceRef = &profile.ItemRef{ID: 99, Kind: "composition", Name: "Child"}
	}))

	report, err := profilediff.Compare(expected, actual, profilediff.Options{})
	if err != nil {
		t.Fatalf("Compare: %v", err)
	}
	assertNoDiff(t, report, "comps.by_id[1].layers.by_id[10].source_ref", profilediff.KindWrongValue)
}

func TestCompareAppliesIgnoreRules(t *testing.T) {
	expected := testProfile(testLayer(10, "Hero", func(l *profile.Layer) {
		l.Flags.Visible = true
	}))
	actual := testProfile(testLayer(10, "Hero", func(l *profile.Layer) {
		l.Flags.Visible = false
	}))

	report, err := profilediff.Compare(expected, actual, profilediff.Options{
		IgnoreRules: &profilediff.IgnoreRules{
			SchemaVersion: 1,
			Rules: []profilediff.IgnoreRule{{
				Path:   "comps.by_id[1].layers.by_id[10].flags.visible",
				Kind:   profilediff.KindWrongValue,
				Reason: "fixture intentionally hides this layer",
			}},
		},
	})
	if err != nil {
		t.Fatalf("Compare: %v", err)
	}
	if len(report.Diffs) != 0 {
		t.Fatalf("len(Diffs) = %d, want 0: %+v", len(report.Diffs), report.Diffs)
	}
	if report.IgnoredCount != 1 {
		t.Fatalf("IgnoredCount = %d, want 1", report.IgnoredCount)
	}
}

func TestCompareRejectsInvalidIgnoreRules(t *testing.T) {
	_, err := profilediff.Compare(testProfile(), testProfile(), profilediff.Options{
		IgnoreRules: &profilediff.IgnoreRules{
			SchemaVersion: 1,
			Rules: []profilediff.IgnoreRule{{
				Path: "comps.by_id[1].layers.by_id[10].flags.visible",
				Kind: profilediff.KindWrongValue,
			}},
		},
	})
	if err == nil {
		t.Fatal("Compare error = nil, want invalid ignore rule error")
	}
}

func TestLoadIgnoreRulesFromJSON(t *testing.T) {
	path := filepath.Join(t.TempDir(), "ignore.json")
	data := []byte(`{
  "schema_version": 1,
  "rules": [
    {
      "path": "comps.by_id[1].layers.by_id[10].flags.visible",
      "kind": "wrong_value",
      "reason": "fixture intentionally hides this layer"
    }
  ]
}`)
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	rules, err := profilediff.LoadIgnoreRules(path)
	if err != nil {
		t.Fatalf("LoadIgnoreRules: %v", err)
	}
	if len(rules.Rules) != 1 {
		t.Fatalf("len(Rules) = %d, want 1", len(rules.Rules))
	}
	if rules.Rules[0].Kind != profilediff.KindWrongValue {
		t.Fatalf("Kind = %q, want %q", rules.Rules[0].Kind, profilediff.KindWrongValue)
	}
}

func testProfile(layers ...profile.Layer) *profile.Profile {
	return &profile.Profile{
		SchemaVersion: profile.SchemaVersion,
		Fingerprint: profile.Fingerprint{
			CompCount:  1,
			LayerCount: len(layers),
		},
		Comps: []profile.Composition{{
			ID:       1,
			Name:     "Main",
			Path:     profile.PathRef{Path: "comps.by_id[1]", DisplayPath: `comps["Main"]`},
			Evidence: profile.Evidence{Level: profile.EvidenceL1Parsed, Source: "test", Confidence: "high"},
			Layers:   layers,
		}},
	}
}

func testLayer(id uint32, name string, edit func(*profile.Layer)) profile.Layer {
	l := profile.Layer{
		ID:    id,
		Index: int(id),
		Name:  name,
		Type:  "av",
		Path: profile.PathRef{
			Path:        "comps.by_id[1].layers.by_id[" + itoa(id) + "]",
			DisplayPath: `comps["Main"].layers["` + name + `"]`,
		},
		Evidence: profile.Evidence{Level: profile.EvidenceL1Parsed, Source: "test", Confidence: "high"},
		Flags:    profile.LayerFlags{Visible: true},
	}
	if edit != nil {
		edit(&l)
	}
	return l
}

func testEffect(occurrence int, matchName string, params []profile.Property) profile.Effect {
	return profile.Effect{
		MatchName:       matchName,
		Occurrence:      occurrence,
		DependencyClass: "native",
		Path: profile.PathRef{
			Path: `comps.by_id[1].layers.by_id[10].effects.by_match_name["` + matchName + `"]#0`,
		},
		Evidence: profile.Evidence{Level: profile.EvidenceL1Parsed, Source: "test", Confidence: "high"},
		Params:   params,
	}
}

func testProperty(path string, matchName string, value any, keyframes []profile.Keyframe) profile.Property {
	return profile.Property{
		MatchName:   matchName,
		StaticValue: value,
		Keyframes:   keyframes,
		Path:        profile.PathRef{Path: path},
		Evidence:    profile.Evidence{Level: profile.EvidenceL1Parsed, Source: "test", Confidence: "high"},
	}
}

func assertDiff(t *testing.T, report *profilediff.Report, path string, kind profilediff.Kind) {
	t.Helper()
	for _, diff := range report.Diffs {
		if diff.Path == path && diff.Kind == kind {
			return
		}
	}
	t.Fatalf("diff %s %s not found in %+v", kind, path, report.Diffs)
}

func assertNoDiff(t *testing.T, report *profilediff.Report, path string, kind profilediff.Kind) {
	t.Helper()
	for _, diff := range report.Diffs {
		if diff.Path == path && diff.Kind == kind {
			t.Fatalf("unexpected diff %s %s in %+v", kind, path, report.Diffs)
		}
	}
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
