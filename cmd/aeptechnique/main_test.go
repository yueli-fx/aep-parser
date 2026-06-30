package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yueli-fx/aep-parser/internal/technique"
)

func TestRunEmitsTechniqueFactsJSON(t *testing.T) {
	input := filepath.Join("..", "..", "flightdeck", "showcase", "text", "text.aep")
	var stdout, stderr bytes.Buffer

	code := run([]string{"-in", input}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("run exit = %d, stderr=%s", code, stderr.String())
	}

	var facts technique.FactSet
	if err := json.Unmarshal(stdout.Bytes(), &facts); err != nil {
		t.Fatalf("json.Unmarshal: %v\nstdout=%s", err, stdout.String())
	}
	if facts.SchemaVersion != technique.SchemaVersion {
		t.Fatalf("SchemaVersion = %d, want %d", facts.SchemaVersion, technique.SchemaVersion)
	}
	if facts.SourcePath != input {
		t.Fatalf("SourcePath = %q, want %q", facts.SourcePath, input)
	}
	if facts.Summary.CompCount == 0 || facts.Summary.LayerCount == 0 || len(facts.Layers) == 0 {
		t.Fatalf("facts summary/layers = %+v / %d", facts.Summary, len(facts.Layers))
	}
}

func TestRunAcceptsJSONFlag(t *testing.T) {
	input := filepath.Join("..", "..", "flightdeck", "showcase", "text", "text.aep")
	var stdout, stderr bytes.Buffer

	code := run([]string{"-in", input, "-json"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("run exit = %d, stderr=%s", code, stderr.String())
	}
	if !json.Valid(stdout.Bytes()) {
		t.Fatalf("stdout is not valid json: %s", stdout.String())
	}
}

func TestRunEmitsPortraitJSONWithMode(t *testing.T) {
	input := filepath.Join("..", "..", "flightdeck", "showcase", "text", "text.aep")
	var stdout, stderr bytes.Buffer

	code := run([]string{"-in", input, "-mode", "portrait"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("run exit = %d, stderr=%s", code, stderr.String())
	}

	var portrait technique.Portrait
	if err := json.Unmarshal(stdout.Bytes(), &portrait); err != nil {
		t.Fatalf("json.Unmarshal: %v\nstdout=%s", err, stdout.String())
	}
	if portrait.SchemaVersion != technique.SchemaVersion || portrait.SourcePath != input {
		t.Fatalf("portrait identity = %+v", portrait)
	}
	if portrait.Fingerprint.CompCount == 0 || portrait.Fingerprint.LayerCount == 0 {
		t.Fatalf("portrait fingerprint = %+v", portrait.Fingerprint)
	}
	if portrait.Mechanisms.ShapeFamilyCounts == nil || portrait.Graph.RelationCounts == nil {
		t.Fatalf("portrait maps not initialized: %+v", portrait)
	}
}

func TestRunEmitsExplanationJSONWithMode(t *testing.T) {
	input := filepath.Join("..", "..", "flightdeck", "showcase", "text", "text.aep")
	var stdout, stderr bytes.Buffer

	code := run([]string{"-in", input, "-mode", "explain"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("run exit = %d, stderr=%s", code, stderr.String())
	}

	var explanation technique.Explanation
	if err := json.Unmarshal(stdout.Bytes(), &explanation); err != nil {
		t.Fatalf("json.Unmarshal: %v\nstdout=%s", err, stdout.String())
	}
	if explanation.SchemaVersion != technique.SchemaVersion || explanation.SourcePath != input {
		t.Fatalf("explanation identity = %+v", explanation)
	}
	if explanation.Portrait.Fingerprint.LayerCount == 0 || len(explanation.Overview) == 0 {
		t.Fatalf("explanation = %+v", explanation)
	}
}

func TestRunAcceptsPortraitFlag(t *testing.T) {
	input := filepath.Join("..", "..", "flightdeck", "showcase", "text", "text.aep")
	var stdout, stderr bytes.Buffer

	code := run([]string{"-in", input, "-portrait"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("run exit = %d, stderr=%s", code, stderr.String())
	}

	var portrait technique.Portrait
	if err := json.Unmarshal(stdout.Bytes(), &portrait); err != nil {
		t.Fatalf("json.Unmarshal: %v\nstdout=%s", err, stdout.String())
	}
	if portrait.Fingerprint.LayerCount == 0 {
		t.Fatalf("portrait fingerprint = %+v", portrait.Fingerprint)
	}
}

func TestRunEmitsCorpusPortraitJSONL(t *testing.T) {
	fixture := filepath.Join("..", "..", "flightdeck", "showcase", "text", "text.aep")
	root := t.TempDir()
	writeFixtureCopy(t, fixture, filepath.Join(root, "b.aep"))
	nested := filepath.Join(root, "nested")
	if err := os.Mkdir(nested, 0o755); err != nil {
		t.Fatalf("Mkdir: %v", err)
	}
	writeFixtureCopy(t, fixture, filepath.Join(nested, "a.aep"))

	var stdout, stderr bytes.Buffer
	code := run([]string{"-in", root, "-mode", "portrait", "-corpus", "-recursive"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("run exit = %d, stderr=%s", code, stderr.String())
	}

	lines := nonEmptyLines(stdout.String())
	if len(lines) != 2 {
		t.Fatalf("jsonl lines = %d, stdout=%s", len(lines), stdout.String())
	}
	var first, second corpusRecord
	if err := json.Unmarshal([]byte(lines[0]), &first); err != nil {
		t.Fatalf("unmarshal first: %v\n%s", err, lines[0])
	}
	if err := json.Unmarshal([]byte(lines[1]), &second); err != nil {
		t.Fatalf("unmarshal second: %v\n%s", err, lines[1])
	}
	if !strings.HasSuffix(first.Path, "b.aep") || !strings.HasSuffix(second.Path, filepath.Join("nested", "a.aep")) {
		t.Fatalf("records not sorted by path: first=%s second=%s", first.Path, second.Path)
	}
	if first.Mode != "portrait" || first.Portrait == nil || first.Portrait.Fingerprint.LayerCount == 0 {
		t.Fatalf("first record = %+v", first)
	}
}

func TestRunRejectsDirectoryCorpusWithoutRecursive(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := run([]string{"-in", t.TempDir(), "-mode", "portrait", "-corpus"}, &stdout, &stderr)
	if code != 2 {
		t.Fatalf("run exit = %d, stderr=%s stdout=%s", code, stderr.String(), stdout.String())
	}
	if !strings.Contains(stderr.String(), "requires -recursive") {
		t.Fatalf("stderr = %s, want requires -recursive", stderr.String())
	}
}

func TestRunEmitsCorpusSummary(t *testing.T) {
	fixture := filepath.Join("..", "..", "flightdeck", "showcase", "text", "text.aep")
	root := t.TempDir()
	writeFixtureCopy(t, fixture, filepath.Join(root, "one.aep"))
	writeFixtureCopy(t, fixture, filepath.Join(root, "two.aep"))

	var stdout, stderr bytes.Buffer
	code := run([]string{"-in", root, "-mode", "portrait", "-corpus", "-recursive", "-summary"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("run exit = %d, stderr=%s", code, stderr.String())
	}

	var summary corpusSummary
	if err := json.Unmarshal(stdout.Bytes(), &summary); err != nil {
		t.Fatalf("json.Unmarshal: %v\nstdout=%s", err, stdout.String())
	}
	if summary.ProjectCount != 2 || summary.ErrorCount != 0 {
		t.Fatalf("summary counts = %+v", summary)
	}
	if summary.Totals.LayerCount != 4 || summary.Totals.ShapeOperatorCount == 0 {
		t.Fatalf("summary totals = %+v", summary.Totals)
	}
	if summary.HintCounts["shape_operator_stack"] != 2 {
		t.Fatalf("hint counts = %+v", summary.HintCounts)
	}
}

func TestRunEmitsCorpusExplanationSummary(t *testing.T) {
	fixture := filepath.Join("..", "..", "flightdeck", "showcase", "text", "text.aep")
	root := t.TempDir()
	writeFixtureCopy(t, fixture, filepath.Join(root, "one.aep"))

	var stdout, stderr bytes.Buffer
	code := run([]string{"-in", root, "-mode", "explain", "-corpus", "-recursive", "-summary"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("run exit = %d, stderr=%s", code, stderr.String())
	}

	var summary corpusSummary
	if err := json.Unmarshal(stdout.Bytes(), &summary); err != nil {
		t.Fatalf("json.Unmarshal: %v\nstdout=%s", err, stdout.String())
	}
	if summary.Mode != "explain" || summary.ProjectCount != 1 || summary.Totals.LayerCount == 0 {
		t.Fatalf("summary = %+v", summary)
	}
	if len(summary.ArchetypeCounts) == 0 {
		t.Fatalf("archetype counts = %+v", summary.ArchetypeCounts)
	}
	if len(summary.PatternCounts) == 0 {
		t.Fatalf("pattern counts = %+v", summary.PatternCounts)
	}
	if len(summary.PatternExamples) == 0 {
		t.Fatalf("pattern examples = %+v", summary.PatternExamples)
	}
	for id, examples := range summary.PatternExamples {
		if len(examples) == 0 {
			t.Fatalf("pattern examples[%s] is empty", id)
		}
		if examples[0].Path == "" || examples[0].Score <= 0 || examples[0].Readiness == "" {
			t.Fatalf("pattern examples[%s][0] = %+v", id, examples[0])
		}
	}
	if len(summary.PatternProfiles) == 0 {
		t.Fatalf("pattern profiles = %+v", summary.PatternProfiles)
	}
	for id, count := range summary.PatternCounts {
		profile := summary.PatternProfiles[id]
		if profile == nil {
			t.Fatalf("pattern profile %s missing from %+v", id, summary.PatternProfiles)
		}
		if profile.Count != count || len(profile.Examples) == 0 {
			t.Fatalf("pattern profile %s = %+v, count=%d", id, profile, count)
		}
		if len(profile.EffectCounts)+len(profile.ShapeFamilies)+len(profile.TextAnimators) == 0 {
			t.Fatalf("pattern profile %s has no mechanism counts: %+v", id, profile)
		}
		if len(profile.RecreationStepCounts) == 0 {
			t.Fatalf("pattern profile %s has no recreation step counts: %+v", id, profile)
		}
	}
	if len(summary.ReadinessCounts) == 0 {
		t.Fatalf("readiness counts = %+v", summary.ReadinessCounts)
	}
	if summary.PluginEffectCounts == nil {
		t.Fatalf("plugin effect counts map is nil")
	}
}

func TestRunEmitsCorpusExplanationRecordsWithFacts(t *testing.T) {
	fixture := filepath.Join("..", "..", "flightdeck", "showcase", "pseudo-effect", "pseudo_default.aep")
	root := t.TempDir()
	writeFixtureCopy(t, fixture, filepath.Join(root, "pseudo_default.aep"))

	var stdout, stderr bytes.Buffer
	code := run([]string{"-in", root, "-mode", "explain", "-corpus", "-recursive"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("run exit = %d, stderr=%s", code, stderr.String())
	}

	lines := nonEmptyLines(stdout.String())
	if len(lines) != 1 {
		t.Fatalf("jsonl lines = %d, stdout=%s", len(lines), stdout.String())
	}
	var record corpusRecord
	if err := json.Unmarshal([]byte(lines[0]), &record); err != nil {
		t.Fatalf("unmarshal record: %v\n%s", err, lines[0])
	}
	if record.Explanation == nil || record.Explanation.Portrait.Fingerprint.EffectCount == 0 {
		t.Fatalf("record explanation missing effect fingerprint: %+v", record.Explanation)
	}
	if record.Facts == nil || len(record.Facts.Effects) == 0 {
		t.Fatalf("record facts missing effects: %+v", record.Facts)
	}
}

func TestRunEmitsPatternPluginProfiles(t *testing.T) {
	fixture := filepath.Join("..", "..", "flightdeck", "showcase", "pseudo-effect", "pseudo_default.aep")
	root := t.TempDir()
	writeFixtureCopy(t, fixture, filepath.Join(root, "pseudo_default.aep"))

	var stdout, stderr bytes.Buffer
	code := run([]string{"-in", root, "-mode", "explain", "-corpus", "-recursive", "-summary"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("run exit = %d, stderr=%s", code, stderr.String())
	}

	var summary corpusSummary
	if err := json.Unmarshal(stdout.Bytes(), &summary); err != nil {
		t.Fatalf("json.Unmarshal: %v\nstdout=%s", err, stdout.String())
	}
	profile := summary.PatternProfiles["plugin_dependent_effect_stack"]
	if profile == nil {
		t.Fatalf("plugin pattern profile missing from %+v", summary.PatternProfiles)
	}
	if len(profile.PluginEffectCounts) == 0 {
		t.Fatalf("plugin pattern profile has no plugin effects: %+v", profile)
	}
}

func TestRunWritesOutputFile(t *testing.T) {
	input := filepath.Join("..", "..", "flightdeck", "showcase", "text", "text.aep")
	outPath := filepath.Join(t.TempDir(), "portrait.json")
	var stdout, stderr bytes.Buffer

	code := run([]string{"-in", input, "-mode", "portrait", "-out", outPath}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("run exit = %d, stderr=%s", code, stderr.String())
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout = %s, want empty when -out is used", stdout.String())
	}
	data, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("ReadFile output: %v", err)
	}
	var portrait technique.Portrait
	if err := json.Unmarshal(data, &portrait); err != nil {
		t.Fatalf("json.Unmarshal output: %v\n%s", err, string(data))
	}
	if portrait.Fingerprint.LayerCount == 0 {
		t.Fatalf("portrait fingerprint = %+v", portrait.Fingerprint)
	}
}

func TestRunWritesCorpusSummaryOutputFile(t *testing.T) {
	fixture := filepath.Join("..", "..", "flightdeck", "showcase", "text", "text.aep")
	root := t.TempDir()
	writeFixtureCopy(t, fixture, filepath.Join(root, "one.aep"))
	outPath := filepath.Join(t.TempDir(), "summary.json")
	var stdout, stderr bytes.Buffer

	code := run([]string{"-in", root, "-mode", "portrait", "-corpus", "-recursive", "-summary", "-out", outPath}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("run exit = %d, stderr=%s", code, stderr.String())
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout = %s, want empty when -out is used", stdout.String())
	}
	data, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("ReadFile output: %v", err)
	}
	var summary corpusSummary
	if err := json.Unmarshal(data, &summary); err != nil {
		t.Fatalf("json.Unmarshal output: %v\n%s", err, string(data))
	}
	if summary.ProjectCount != 1 {
		t.Fatalf("summary = %+v", summary)
	}
}

func TestRunWritesCorpusJSONLAndSummaryInOnePass(t *testing.T) {
	fixture := filepath.Join("..", "..", "flightdeck", "showcase", "text", "text.aep")
	root := t.TempDir()
	writeFixtureCopy(t, fixture, filepath.Join(root, "one.aep"))
	writeFixtureCopy(t, fixture, filepath.Join(root, "two.aep"))
	outDir := t.TempDir()
	corpusPath := filepath.Join(outDir, "corpus.jsonl")
	summaryPath := filepath.Join(outDir, "summary.json")
	var stdout, stderr bytes.Buffer

	code := run([]string{
		"-in", root,
		"-mode", "explain",
		"-corpus",
		"-recursive",
		"-out", corpusPath,
		"-summary-out", summaryPath,
	}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("run exit = %d, stderr=%s", code, stderr.String())
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout = %s, want empty when -out is used", stdout.String())
	}
	lines := nonEmptyLines(string(mustReadFile(t, corpusPath)))
	if len(lines) != 2 {
		t.Fatalf("jsonl lines = %d, want 2\n%s", len(lines), strings.Join(lines, "\n"))
	}
	var summary corpusSummary
	if err := json.Unmarshal(mustReadFile(t, summaryPath), &summary); err != nil {
		t.Fatalf("json.Unmarshal summary: %v", err)
	}
	if summary.Mode != "explain" || summary.ProjectCount != 2 || len(summary.PatternCounts) == 0 {
		t.Fatalf("summary = %+v", summary)
	}
}

func writeFixtureCopy(t *testing.T, src, dst string) {
	t.Helper()
	data, err := os.ReadFile(src)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if err := os.WriteFile(dst, data, 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
}

func mustReadFile(t *testing.T, path string) []byte {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile %s: %v", path, err)
	}
	return data
}

func nonEmptyLines(text string) []string {
	var out []string
	for _, line := range strings.Split(text, "\n") {
		if strings.TrimSpace(line) != "" {
			out = append(out, line)
		}
	}
	return out
}
