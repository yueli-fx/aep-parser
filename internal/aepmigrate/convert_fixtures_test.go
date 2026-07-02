package aepmigrate

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/yueli-fx/aep-parser/internal/aep"
)

func writeProjectFile(t *testing.T, project *aep.Project, path string) {
	t.Helper()
	out, err := os.Create(path)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	defer out.Close()
	if err := project.WriteAEP(out); err != nil {
		t.Fatalf("WriteAEP: %v", err)
	}
}

func writeTempProjectWithOneComp(t *testing.T, target aep.AETarget) string {
	t.Helper()
	project := aep.NewProject(target)
	if _, err := aep.NewComposition(project, "Main", 640, 360, 24, 2.5); err != nil {
		t.Fatalf("NewComposition: %v", err)
	}
	path := filepath.Join(t.TempDir(), "one-comp.aep")
	writeProjectFile(t, project, path)
	return path
}
