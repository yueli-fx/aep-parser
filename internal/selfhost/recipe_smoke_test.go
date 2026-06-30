package selfhost

import (
	"context"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yueli-fx/aep-parser/internal/host"
)

func TestRunRecipeDraftSmokeWritesCompileReparseAndBatchArtifacts(t *testing.T) {
	root := t.TempDir()
	draftsPath := filepath.Join(root, "full_report", "recipe_drafts.jsonl")
	writeRecipeDrafts(t, draftsPath, []string{"first.aep", "second.aep"})
	runner := &recipeSmokeFakeRunner{t: t}

	result, err := RunRecipeDraftSmoke(context.Background(), RecipeDraftSmokeOptions{
		DraftsPath: draftsPath,
		CompileDir: filepath.Join(root, "recipe_draft_compile"),
		ReparseDir: filepath.Join(root, "recipe_draft_reparse"),
		BatchDir:   filepath.Join(root, "recipe_draft_batch"),
		BatchLimit: 2,
		Runner:     runner,
		WorkingDir: root,
	})

	if err != nil {
		t.Fatalf("RunRecipeDraftSmoke: %v", err)
	}
	if result.Compile.OutputBytes == 0 {
		t.Fatalf("compile output bytes = 0")
	}
	if !result.Reparse.Passed {
		t.Fatalf("reparse summary did not pass: %+v", result.Reparse)
	}
	if result.Batch.Requested != 2 || result.Batch.Attempted != 2 || result.Batch.Passed != 2 {
		t.Fatalf("batch summary = %+v", result.Batch)
	}
	for _, path := range []string{
		filepath.Join(root, "recipe_draft_compile", "recipe_draft.json"),
		filepath.Join(root, "recipe_draft_compile", "validate.json"),
		filepath.Join(root, "recipe_draft_compile", "compile.json"),
		filepath.Join(root, "recipe_draft_reparse", "compiled_facts.json"),
		filepath.Join(root, "recipe_draft_reparse", "reparse_summary.json"),
		filepath.Join(root, "recipe_draft_batch", "summary.json"),
	} {
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("expected artifact %s: %v", path, err)
		}
	}
	if len(runner.commands) != 9 {
		t.Fatalf("runner commands = %d, want 9: %+v", len(runner.commands), runner.commands)
	}
}

func TestRunRecipeDraftSmokeFailsWhenDraftsAreEmpty(t *testing.T) {
	root := t.TempDir()
	draftsPath := filepath.Join(root, "recipe_drafts.jsonl")
	if err := os.MkdirAll(filepath.Dir(draftsPath), 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	if err := os.WriteFile(draftsPath, []byte("\n"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	_, err := RunRecipeDraftSmoke(context.Background(), RecipeDraftSmokeOptions{
		DraftsPath: draftsPath,
		CompileDir: filepath.Join(root, "compile"),
		ReparseDir: filepath.Join(root, "reparse"),
		BatchDir:   filepath.Join(root, "batch"),
		Runner:     &recipeSmokeFakeRunner{t: t},
	})

	if err == nil || !strings.Contains(err.Error(), "recipe_drafts.jsonl has no rows") {
		t.Fatalf("err = %v, want no rows", err)
	}
}

type recipeSmokeFakeRunner struct {
	t        *testing.T
	commands []host.Command
}

func (r *recipeSmokeFakeRunner) Run(_ context.Context, cmd host.Command) host.Result {
	r.commands = append(r.commands, cmd)
	joined := strings.Join(append([]string{cmd.Name}, cmd.Args...), " ")
	switch {
	case strings.Contains(joined, "./cmd/aeprecipe validate"):
		writeString(r.t, cmd.Stdout, `{"valid":true}`+"\n")
	case strings.Contains(joined, "./cmd/aeprecipe compile"):
		writeString(r.t, cmd.Stdout, `{"valid":true}`+"\n")
		outPath := argAfter(cmd.Args, "-out")
		if outPath == "" {
			r.t.Fatalf("compile command missing -out: %+v", cmd.Args)
		}
		if err := os.WriteFile(outPath, []byte("fake aep"), 0o644); err != nil {
			r.t.Fatalf("WriteFile compiled AEP: %v", err)
		}
	case strings.Contains(joined, "./cmd/aeptechnique"):
		outPath := argAfter(cmd.Args, "-out")
		if outPath == "" {
			r.t.Fatalf("technique command missing -out: %+v", cmd.Args)
		}
		writeJSONFile(r.t, outPath, map[string]any{
			"summary": map[string]any{
				"comp_count":  2,
				"layer_count": 0,
			},
		})
	default:
		r.t.Fatalf("unexpected command: %s", joined)
	}
	return host.Result{ExitCode: 0}
}

func (r *recipeSmokeFakeRunner) Start(context.Context, host.Command) (host.Process, error) {
	r.t.Fatalf("Start should not be called")
	return nil, nil
}

func writeRecipeDrafts(t *testing.T, path string, projects []string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	var b strings.Builder
	for _, project := range projects {
		row := map[string]any{
			"schema_version": 1,
			"project_path":   project,
			"recipe": map[string]any{
				"schema_version": 1,
				"project":        map[string]any{"name": project},
				"comps":          []any{map[string]any{"name": "A"}, map[string]any{"name": "B"}},
				"expected_profile": map[string]any{
					"comp_count":  2,
					"layer_count": 0,
				},
			},
		}
		data, err := json.Marshal(row)
		if err != nil {
			t.Fatalf("Marshal: %v", err)
		}
		b.Write(data)
		b.WriteByte('\n')
	}
	if err := os.WriteFile(path, []byte(b.String()), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
}

func writeString(t *testing.T, w io.Writer, value string) {
	t.Helper()
	if _, err := io.WriteString(w, value); err != nil {
		t.Fatalf("WriteString: %v", err)
	}
}

func argAfter(args []string, name string) string {
	for i := 0; i < len(args)-1; i++ {
		if args[i] == name {
			return args[i+1]
		}
	}
	return ""
}
