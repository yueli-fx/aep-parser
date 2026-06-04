package main

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// manifest：每个输出文件一节。所有相对路径相对 manifest 文件自身目录解析（CWD 无关）。
type manifest struct {
	Pkg     string         `json:"pkg"`
	Files   []fileManifest `json:"files"`
	baseDir string         // = filepath.Dir(manifestPath)，不序列化
}

type fileManifest struct {
	Out   string   `json:"out"`
	Roots []string `json:"roots"`
	Head  string   `json:"head"`
	Tail  string   `json:"tail"`
}

func loadManifest(path string) (*manifest, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var m manifest
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, err
	}
	m.baseDir = filepath.Dir(path)
	return &m, nil
}

// resolve 把 manifest 内相对路径锚到 baseDir。
func (m *manifest) resolve(p string) string {
	if p == "" || filepath.IsAbs(p) {
		return p
	}
	return filepath.Join(m.baseDir, p)
}
