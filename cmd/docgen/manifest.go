package main

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// manifest：每个输出文件一节。所有相对路径相对 manifest 文件自身目录解析（CWD 无关）。
type manifest struct {
	Pkg         string         `json:"pkg"`
	Pkgs        []string       `json:"pkgs"`         // 额外扫描的包（M8 多包分层后：scene 含真类型/方法、codec 含值类型；Pkg=aep facade 含 re-export 自由函数）
	ExamplePkgs []string       `json:"example_pkgs"` // 只扫描 Example 函数，不参与 API 类型/函数抽取
	Files       []fileManifest `json:"files"`
	Index       string         `json:"index"` // 可选：JSON 符号索引输出路径（相对 baseDir）
	baseDir     string         // = filepath.Dir(manifestPath)，不序列化
}

// pkgDirs 返回所有待扫描包目录（已解析），Pkg 在前（自由函数优先从 facade 取），
// 其后是 Pkgs。空 Pkg 跳过。
func (m *manifest) pkgDirs() []string {
	var dirs []string
	if m.Pkg != "" {
		dirs = append(dirs, m.resolve(m.Pkg))
	}
	for _, p := range m.Pkgs {
		dirs = append(dirs, m.resolve(p))
	}
	return dirs
}

// exampleDirs 返回用于抽取 Example 函数的目录。默认复用 API 包目录，
// ExamplePkgs 可补充外部测试包目录，避免源码包被大量 example 测试淹没。
func (m *manifest) exampleDirs() []string {
	var dirs []string
	seen := map[string]bool{}
	appendDir := func(d string) {
		if d == "" {
			return
		}
		clean := filepath.Clean(d)
		if seen[clean] {
			return
		}
		seen[clean] = true
		dirs = append(dirs, d)
	}
	for _, d := range m.pkgDirs() {
		appendDir(d)
	}
	for _, p := range m.ExamplePkgs {
		appendDir(m.resolve(p))
	}
	return dirs
}

type fileManifest struct {
	Out   string   `json:"out"`
	Roots []string `json:"roots"`
	Funcs []string `json:"funcs"` // 可选：渲进 "## Functions" 的包级函数名
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
