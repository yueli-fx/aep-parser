package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/example/aep-parser/internal/aep"
)

type entry struct {
	OpenOK bool   `json:"open_ok"`
	SHA    string `json:"sha,omitempty"`
	Err    string `json:"err,omitempty"`
}

func main() {
	roots := []string{"test_data", "data"}
	out := map[string]entry{}
	for _, root := range roots {
		_ = filepath.Walk(root, func(p string, info os.FileInfo, err error) error {
			if err != nil || info.IsDir() || filepath.Ext(p) != ".aep" {
				return nil
			}
			key := filepath.ToSlash(p)
			proj, oerr := aep.Open(p)
			if oerr != nil {
				out[key] = entry{OpenOK: false, Err: oerr.Error()}
				return nil
			}
			var buf bytes.Buffer
			if werr := proj.WriteAEP(&buf); werr != nil {
				out[key] = entry{OpenOK: true, Err: "WriteAEP: " + werr.Error()}
				return nil
			}
			sum := sha256.Sum256(buf.Bytes())
			out[key] = entry{OpenOK: true, SHA: hex.EncodeToString(sum[:])}
			return nil
		})
	}
	keys := make([]string, 0, len(out))
	for k := range out {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	ordered := map[string]entry{}
	for _, k := range keys {
		ordered[k] = out[k]
	}
	b, _ := json.MarshalIndent(ordered, "", "  ")
	if err := os.WriteFile("tmp_debug/reorg_baseline/baseline.json", b, 0o644); err != nil {
		panic(err)
	}
	fmt.Printf("baseline: %d fixtures\n", len(out))
}
