package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/yueli-fx/aep-parser/internal/aehost"
	"github.com/yueli-fx/aep-parser/internal/aepmigrate"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout, stderr io.Writer) int {
	return runWithHost(args, stdout, stderr, aehost.DefaultHost())
}

func runWithHost(args []string, stdout, stderr io.Writer, host aehost.Host) int {
	if len(args) == 0 {
		fmt.Fprintln(stderr, "usage: aepmigrate assess -in source.aep -target AE2020 [-out assess.json]")
		return 2
	}
	switch args[0] {
	case "assess":
		return runAssess(args[1:], stdout, stderr)
	case "convert":
		return runConvert(args[1:], stdout, stderr, host)
	case "matrix":
		return runMatrix(args[1:], stdout, stderr, host)
	default:
		fmt.Fprintf(stderr, "unknown command %q\n", args[0])
		return 2
	}
}

func runAssess(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("aepmigrate assess", flag.ContinueOnError)
	fs.SetOutput(stderr)
	input := fs.String("in", "", "source .aep path")
	targetRaw := fs.String("target", "", "target AE version: AE2020, AE2022, or AE2025")
	outPath := fs.String("out", "", "optional JSON output path")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if *input == "" || *targetRaw == "" {
		fmt.Fprintln(stderr, "usage: aepmigrate assess -in source.aep -target AE2020 [-out assess.json]")
		return 2
	}
	target, err := aepmigrate.ParseVersionLabel(*targetRaw)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 2
	}
	report, err := aepmigrate.Assess(aepmigrate.Options{InputPath: *input, Target: target})
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	data = append(data, '\n')
	if *outPath != "" {
		if err := os.WriteFile(*outPath, data, 0o644); err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		fmt.Fprintf(stdout, "migration assess: %s\n", *outPath)
		return statusCode(report.Summary.Status)
	}
	fmt.Fprint(stdout, string(data))
	return statusCode(report.Summary.Status)
}

func runMatrix(args []string, stdout, stderr io.Writer, host aehost.Host) int {
	fs := flag.NewFlagSet("aepmigrate matrix", flag.ContinueOnError)
	fs.SetOutput(stderr)
	var recipePaths stringListFlag
	fs.Var(&recipePaths, "recipe", "recipe JSON path; may be repeated")
	recipeDir := fs.String("recipes", "", "directory containing recipe JSON files")
	sourcesRaw := fs.String("sources", "", "comma-separated source writer versions, or recipe")
	targetsRaw := fs.String("targets", "", "comma-separated target writer versions")
	outRoot := fs.String("out", "", "matrix output root")
	aeRoot := fs.String("ae-root", "", "After Effects install root for host discovery")
	var aeHosts aeHostFlag
	fs.Var(&aeHosts, "ae", "AE host mapping like AE2025=E:\\adobe\\Adobe After Effects 2025\\Support Files\\AfterFX.exe; may be repeated")
	aeOpen := fs.Bool("ae-open", false, "run AE open verification for each converted target")
	aeOpenVersionsRaw := fs.String("ae-versions", "", "comma-separated AE host versions for -ae-open; defaults to each target version")
	aeTimeout := fs.Int("ae-timeout-sec", 180, "AE open verification timeout in seconds")
	maxAEOpenCases := fs.Int("max-ae-open-cases", 25, "maximum AE-open cases allowed in one matrix run; set 0 to disable")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if *recipeDir == "" && len(recipePaths) == 0 || *outRoot == "" {
		fmt.Fprintln(stderr, "usage: aepmigrate matrix (-recipe recipe.json | -recipes dir) -out tmp\\matrix [-sources AE2020,recipe] [-targets AE2020,AE2025] [-ae-root E:\\adobe] [-ae-open]")
		return 2
	}
	opts := aepmigrate.MatrixOptions{
		RecipePaths:      []string(recipePaths),
		RecipeDir:        *recipeDir,
		SourceLabels:     splitCSV(*sourcesRaw),
		TargetLabels:     splitCSV(*targetsRaw),
		OutRoot:          *outRoot,
		AEInstallRoot:    *aeRoot,
		AEHosts:          map[string]string(aeHosts),
		AEOpen:           *aeOpen,
		AEOpenLabels:     splitCSV(*aeOpenVersionsRaw),
		AEOpenTimeoutSec: *aeTimeout,
		MaxAEOpenCases:   *maxAEOpenCases,
		Host:             host,
	}
	report, err := aepmigrate.RunMatrix(opts)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	reportPath := filepath.Join(*outRoot, "matrix.json")
	fmt.Fprintf(stdout, "migration matrix: %s\n", reportPath)
	fmt.Fprintf(stdout, "matrix summary: total=%d pass=%d blocked=%d failed=%d skipped=%d\n",
		report.Summary.Total,
		report.Summary.Passed,
		report.Summary.Blocked,
		report.Summary.Failed,
		report.Summary.Skipped,
	)
	if report.Summary.Failed > 0 || report.Summary.Blocked > 0 {
		return 1
	}
	return 0
}

func runConvert(args []string, stdout, stderr io.Writer, host aehost.Host) int {
	fs := flag.NewFlagSet("aepmigrate convert", flag.ContinueOnError)
	fs.SetOutput(stderr)
	input := fs.String("in", "", "source .aep path")
	targetRaw := fs.String("target", "", "target AE version: AE2020, AE2022, or AE2025")
	outPath := fs.String("out", "", "target .aep output path")
	reportPath := fs.String("report", "", "JSON migration report output path")
	aeOpen := fs.Bool("ae-open", false, "run AE open verification after profile diff")
	aePath := fs.String("ae", "", "After Effects executable path for -ae-open")
	aeTimeout := fs.Int("ae-timeout-sec", 180, "AE open verification timeout in seconds")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if *input == "" || *targetRaw == "" || *outPath == "" || *reportPath == "" {
		fmt.Fprintln(stderr, "usage: aepmigrate convert -in source.aep -target AE2025 -out migrated.aep -report report.json")
		return 2
	}
	target, err := aepmigrate.ParseVersionLabel(*targetRaw)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 2
	}
	opts := aepmigrate.ConvertOptions{InputPath: *input, OutputPath: *outPath, Target: target}
	if *aeOpen {
		if *aePath == "" {
			fmt.Fprintln(stderr, "-ae-open requires -ae <AfterFX.exe>")
			return 2
		}
		jsxPath, err := filepath.Abs(filepath.Join("test_data", "generators", "verify_open.jsx"))
		if err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		argsPath, err := filepath.Abs(*outPath + ".ae_open.args.json")
		if err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		donePath, err := filepath.Abs(*outPath + ".ae_open.done")
		if err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		opts.AEOpen = &aepmigrate.AEOpenOptions{
			Host:       host,
			AEPath:     *aePath,
			JSXPath:    jsxPath,
			ArgsPath:   argsPath,
			DonePath:   donePath,
			TimeoutSec: *aeTimeout,
		}
	}
	report, err := aepmigrate.Convert(opts)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	if err := writeJSONReport(*reportPath, report); err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	if report.Summary.Status == aepmigrate.StatusBlocked {
		fmt.Fprintf(stdout, "migration convert blocked: %s\n", *reportPath)
		return 1
	}
	fmt.Fprintf(stdout, "migration convert: %s\n", *outPath)
	fmt.Fprintf(stdout, "migration report: %s\n", *reportPath)
	return statusCode(report.Summary.Status)
}

type stringListFlag []string

func (f *stringListFlag) String() string {
	return strings.Join(*f, ",")
}

func (f *stringListFlag) Set(value string) error {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	*f = append(*f, value)
	return nil
}

type aeHostFlag map[string]string

func (f *aeHostFlag) String() string {
	if f == nil {
		return ""
	}
	values := make([]string, 0, len(*f))
	for k, v := range *f {
		values = append(values, k+"="+v)
	}
	return strings.Join(values, ",")
}

func (f *aeHostFlag) Set(value string) error {
	if *f == nil {
		*f = map[string]string{}
	}
	label, path, ok := strings.Cut(value, "=")
	if !ok || strings.TrimSpace(label) == "" || strings.TrimSpace(path) == "" {
		return fmt.Errorf("-ae must be LABEL=PATH")
	}
	(*f)[strings.ToUpper(strings.TrimSpace(label))] = strings.TrimSpace(path)
	return nil
}

func splitCSV(value string) []string {
	var out []string
	for _, part := range strings.Split(value, ",") {
		part = strings.TrimSpace(part)
		if part != "" {
			out = append(out, part)
		}
	}
	return out
}

func writeJSONReport(path string, report aepmigrate.Report) error {
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(path, data, 0o644)
}

func statusCode(status aepmigrate.Status) int {
	if status == aepmigrate.StatusBlocked || status == aepmigrate.StatusError {
		return 1
	}
	return 0
}
