package toolkitcli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/yueli-fx/aep-parser/internal/aep"
)

func TestRunInspectProfileDiffAndCapabilities(t *testing.T) {
	path := filepath.Join("..", "..", "test_data", "fixtures", "v2_smoke_ae2020.aep")
	for _, test := range []struct {
		name string
		args []string
	}{
		{name: "inspect", args: []string{"inspect", "-in", path}},
		{name: "profile", args: []string{"profile", "-in", path}},
		{name: "diff", args: []string{"diff", "-expected", path, "-actual", path}},
		{name: "capabilities", args: []string{"capabilities"}},
	} {
		t.Run(test.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			if code := Run(test.args, &stdout, &stderr); code != 0 {
				t.Fatalf("code = %d, stderr = %s", code, stderr.String())
			}
			var result envelope
			if err := json.Unmarshal(stdout.Bytes(), &result); err != nil || result.Command != test.name {
				t.Fatalf("stdout = %s, err = %v", stdout.String(), err)
			}
		})
	}
}

func TestRunMigrateWritesOutputAndReport(t *testing.T) {
	project := aep.NewProject(aep.TargetAE2020)
	if _, err := aep.NewComposition(project, "Main", 320, 180, 24, 1); err != nil {
		t.Fatal(err)
	}
	input := filepath.Join(t.TempDir(), "source.aep")
	file, err := os.Create(input)
	if err != nil {
		t.Fatal(err)
	}
	if err := project.WriteAEP(file); err != nil {
		file.Close()
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}
	output := filepath.Join(t.TempDir(), "migrated.aep")
	var stdout, stderr bytes.Buffer
	code := Run([]string{"migrate", "-in", input, "-target", "AE2025", "-out", output}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("code = %d, stderr = %s, stdout = %s", code, stderr.String(), stdout.String())
	}
	if info, err := os.Stat(output); err != nil || info.Size() == 0 {
		t.Fatalf("output info = %v, err = %v", info, err)
	}
}

func TestRunErrorsUseJSONEnvelope(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if code := Run([]string{"unknown"}, &stdout, &stderr); code != 2 {
		t.Fatalf("code = %d", code)
	}
	var result errorEnvelope
	if err := json.Unmarshal(stderr.Bytes(), &result); err != nil || result.Error.Code != "unknown_command" {
		t.Fatalf("stderr = %s, err = %v", stderr.String(), err)
	}
}
