package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"

	"github.com/yueli-fx/aep-parser/internal/aep"
	"github.com/yueli-fx/aep-parser/internal/projectindex"
)

type openProjectFunc func(string) (*aep.Project, error)

type SearchReport struct {
	SchemaVersion int                `json:"schema_version"`
	Mode          string             `json:"mode"`
	ProjectPath   string             `json:"project_path"`
	Kind          string             `json:"kind"`
	Query         string             `json:"query"`
	Hits          []projectindex.Hit `json:"hits,omitempty"`
}

func main() {
	os.Exit(run(os.Args[1:]))
}

func run(args []string) int {
	return runWithIO(args, os.Stdout, os.Stderr, aep.Open)
}

func runWithIO(args []string, stdout, stderr io.Writer, openProject openProjectFunc) int {
	if len(args) == 0 {
		usage(stderr)
		return 2
	}
	switch args[0] {
	case "search":
		return runSearch(args[1:], stdout, stderr, openProject)
	case "corpus":
		return runCorpus(args[1:], stdout, stderr, openProject)
	default:
		usage(stderr)
		return 2
	}
}

func runSearch(args []string, stdout, stderr io.Writer, openProject openProjectFunc) int {
	fs := flag.NewFlagSet("aepsearch search", flag.ContinueOnError)
	fs.SetOutput(stderr)
	aepPath := fs.String("aep", "", "AEP file to search")
	kind := fs.String("kind", "", "search kind: effect, source, property, expression, font")
	query := fs.String("query", "", "search query")
	jsonOut := fs.Bool("json", false, "emit JSON report")
	outPath := fs.String("out", "", "optional output path")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if *aepPath == "" || *kind == "" || *query == "" {
		fmt.Fprintln(stderr, "usage: aepsearch search -aep file.aep -kind effect|source|property|expression|font -query value [-json] [-out report.json]")
		return 2
	}
	project, err := openProject(*aepPath)
	if err != nil {
		fmt.Fprintf(stderr, "open %s: %v\n", *aepPath, err)
		return 2
	}
	hits, ok := searchProject(projectindex.Build(project), *kind, *query)
	if !ok {
		fmt.Fprintf(stderr, "unsupported search kind %q\n", *kind)
		return 2
	}
	report := SearchReport{
		SchemaVersion: 1,
		Mode:          "search",
		ProjectPath:   *aepPath,
		Kind:          *kind,
		Query:         *query,
		Hits:          hits,
	}
	if err := emitSearchReport(report, *jsonOut, *outPath, stdout); err != nil {
		fmt.Fprintln(stderr, err)
		return 2
	}
	if len(hits) == 0 {
		return 1
	}
	return 0
}

func runCorpus(args []string, stdout, stderr io.Writer, openProject openProjectFunc) int {
	fs := flag.NewFlagSet("aepsearch corpus", flag.ContinueOnError)
	fs.SetOutput(stderr)
	jsonOut := fs.Bool("json", false, "emit JSON corpus")
	outPath := fs.String("out", "", "optional output path")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	paths := fs.Args()
	if len(paths) == 0 {
		fmt.Fprintln(stderr, "usage: aepsearch corpus [-json] [-out corpus.json] file1.aep [file2.aep ...]")
		return 2
	}
	builder := projectindex.NewCorpusBuilder()
	for _, path := range paths {
		project, err := openProject(path)
		if err != nil {
			fmt.Fprintf(stderr, "open %s: %v\n", path, err)
			return 2
		}
		if err := builder.AddProject(path, project); err != nil {
			fmt.Fprintf(stderr, "index %s: %v\n", path, err)
			return 2
		}
	}
	corpus, err := builder.Build()
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 2
	}
	if err := emitCorpus(corpus, *jsonOut, *outPath, stdout); err != nil {
		fmt.Fprintln(stderr, err)
		return 2
	}
	return 0
}

func searchProject(idx *projectindex.Index, kind, query string) ([]projectindex.Hit, bool) {
	switch kind {
	case "effect":
		return idx.SearchEffectsByMatchName(query), true
	case "source":
		id, err := strconv.ParseUint(query, 10, 32)
		if err != nil || id == 0 {
			return nil, false
		}
		return idx.SearchLayersBySourceID(uint32(id)), true
	case "property":
		return idx.SearchPropertiesByMatchName(query), true
	case "expression":
		return idx.SearchExpressionsContaining(query), true
	case "font":
		return idx.SearchTextStylesByFont(query), true
	default:
		return nil, false
	}
}

func emitSearchReport(report SearchReport, jsonOut bool, outPath string, stdout io.Writer) error {
	if jsonOut || outPath != "" {
		return writeJSON(outPath, stdout, report)
	}
	for _, hit := range report.Hits {
		fmt.Fprintf(stdout, "%s\t%s=%s\tcomp=%q\tlayer=%q\n",
			hit.Kind, hit.Match.Field, hit.Match.Value, hit.Location.CompName, hit.Location.LayerName)
	}
	return nil
}

func emitCorpus(corpus *projectindex.Corpus, jsonOut bool, outPath string, stdout io.Writer) error {
	if jsonOut || outPath != "" {
		return writeJSON(outPath, stdout, corpus)
	}
	fmt.Fprintf(stdout, "projects=%d facts=%d\n", len(corpus.Projects), len(corpus.Facts))
	for _, fact := range corpus.Facts {
		fmt.Fprintf(stdout, "%s\t%s=%s\tproject=%s\tcomp=%q\tlayer=%q\n",
			fact.Kind, fact.Match.Field, fact.Match.Value, fact.ProjectID, fact.Location.CompName, fact.Location.LayerName)
	}
	return nil
}

func writeJSON(path string, stdout io.Writer, value any) error {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	if path != "" {
		return os.WriteFile(path, data, 0o644)
	}
	_, err = stdout.Write(data)
	return err
}

func usage(stderr io.Writer) {
	fmt.Fprintln(stderr, "usage: aepsearch search|corpus ...")
}
