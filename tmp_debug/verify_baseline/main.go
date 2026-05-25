// tmp_debug/verify_baseline/main.go
//
// Open one .aep through AE 2025 with verify_baseline.jsx; dump rich metrics
// (items, compItems, layers, activeItem, selection, layer classes). Per GPT's
// "epistemic hole" check: verify our probe measurement is sound by running
// tolerance.aep — if tolerance also reports layers.length=0, the 6 iters of
// silent-drop hypothesis were measurement artifacts.
//
// Usage:
//   go run tmp_debug/verify_baseline/main.go <path-to-.aep>
//
// Prints the entire .done content to stdout after AE quits.
package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

const (
	aeExe    = `E:/adobe/Adobe After Effects 2025/Support Files/AfterFX.exe`
	jsxPath  = `E:/projects/tools/aep-parser/test_data/verify_baseline.jsx`
	argsPath = `e:/projects/tools/aep-parser/test_data/verify_baseline_args.json`
	donePath = `e:/projects/tools/aep-parser/test_data/verify_baseline.done`
	timeout  = 90 * time.Second
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: verify_baseline <path-to-.aep>")
		os.Exit(1)
	}
	abs, err := filepath.Abs(os.Args[1])
	if err != nil {
		panic(err)
	}
	toFwd := func(p string) string { return strings.ReplaceAll(p, `\`, `/`) }
	argsJSON := fmt.Sprintf(`{"input":%q,"done":%q}`, toFwd(abs), toFwd(donePath))
	if err := os.WriteFile(argsPath, []byte(argsJSON), 0644); err != nil {
		panic(err)
	}
	_ = os.Remove(donePath)

	fmt.Printf("Opening %s in AE 2025...\n", abs)
	cmd := exec.Command(aeExe, "-r", jsxPath)
	if err := cmd.Start(); err != nil {
		panic(err)
	}

	deadline := time.Now().Add(timeout)
	for {
		if _, err := os.Stat(donePath); err == nil {
			break
		}
		if time.Now().After(deadline) {
			_ = cmd.Process.Kill()
			fmt.Fprintln(os.Stderr, "TIMEOUT — no .done after", timeout)
			os.Exit(1)
		}
		time.Sleep(1 * time.Second)
	}
	_ = cmd.Wait()

	data, err := os.ReadFile(donePath)
	if err != nil {
		panic(err)
	}
	fmt.Println("--- .done content ---")
	fmt.Println(string(data))
}
