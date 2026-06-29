package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// directImports 返回某包的直接 import 列表（go list -f）。
func directImports(pkg string) ([]string, error) {
	out, err := exec.Command("go", "list", "-f", "{{range .Imports}}{{.}}\n{{end}}", pkg).Output()
	if err != nil {
		return nil, err
	}
	return strings.Split(strings.TrimSpace(string(out)), "\n"), nil
}

func main() {
	const base = "github.com/yueli-fx/aep-parser/internal/"
	type rule struct {
		pkg    string
		banned []string
	}
	rules := []rule{
		{"scene", []string{base + "rifx", base + "serializer", base + "aep"}},
		{"serializer", []string{base + "aep"}},
		{"codec", []string{base + "scene", base + "rifx"}},
	}
	var fail int
	for _, r := range rules {
		imps, err := directImports(base + r.pkg)
		if err != nil {
			fmt.Printf("SKIP %s (%v)\n", r.pkg, err) // 包未拆出
			continue
		}
		for _, imp := range imps {
			for _, b := range r.banned {
				if imp == b {
					fmt.Printf("FAIL %s imports %s\n", r.pkg, imp)
					fail++
				}
			}
		}
	}
	if fail > 0 {
		os.Exit(1)
	}
	fmt.Println("DAG OK")
}
