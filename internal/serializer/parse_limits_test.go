package serializer

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/yueli-fx/aep-parser/internal/rifx"
)

func TestReadFileBoundedRejectsBeforeParserAllocation(t *testing.T) {
	path := filepath.Join(t.TempDir(), "oversized.aep")
	if err := os.WriteFile(path, []byte("12345"), 0o600); err != nil {
		t.Fatal(err)
	}
	_, err := readFileBounded(path, 4)
	var limitErr *rifx.LimitError
	if !errors.As(err, &limitErr) || limitErr.Resource != "input bytes" {
		t.Fatalf("error = %v, want input-bytes LimitError", err)
	}
}
