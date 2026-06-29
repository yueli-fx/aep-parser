package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"

	"github.com/yueli-fx/aep-parser/internal/aep"
)

func main() {
	root := "test_data"
	var fail int
	_ = filepath.Walk(root, func(p string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || filepath.Ext(p) != ".aep" {
			return nil
		}
		proj, err := aep.Open(p)
		if err != nil {
			fmt.Printf("%s OPEN-ERR %v\n", p, err)
			fail++
			return nil
		}
		var buf bytes.Buffer
		if err := proj.WriteAEP(&buf); err != nil {
			fmt.Printf("%s WRITE-ERR %v\n", p, err)
			fail++
			return nil
		}
		sum := sha256.Sum256(buf.Bytes())
		fmt.Printf("%s %s\n", p, hex.EncodeToString(sum[:]))
		return nil
	})
	if fail > 0 {
		os.Exit(1)
	}
}
