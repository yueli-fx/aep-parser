package registry

import (
	"os"
	"path/filepath"
	"sort"
)

type OwnershipOptions struct {
	SampleLimit int
}

type OwnershipReport struct {
	SchemaVersion int                 `json:"schema_version"`
	Summary       OwnershipSummary    `json:"summary"`
	Locations     []LocationOwnership `json:"locations"`
}

type OwnershipSummary struct {
	Locations    int `json:"locations"`
	Files        int `json:"files"`
	OwnedFiles   int `json:"owned_files"`
	UnownedFiles int `json:"unowned_files"`
	SampleLimit  int `json:"sample_limit"`
}

type LocationOwnership struct {
	ID             string         `json:"id"`
	Path           string         `json:"path"`
	Class          string         `json:"class"`
	Files          int            `json:"files"`
	OwnedFiles     int            `json:"owned_files"`
	UnownedFiles   int            `json:"unowned_files"`
	UnownedSamples []string       `json:"unowned_samples,omitempty"`
	UnownedGroups  []UnownedGroup `json:"unowned_groups,omitempty"`
}

type UnownedGroup struct {
	Name  string `json:"name"`
	Files int    `json:"files"`
}

func OwnershipRepository(root string, opts OwnershipOptions) (OwnershipReport, error) {
	reg, err := Load(root)
	if err != nil {
		return OwnershipReport{}, err
	}
	return Ownership(root, reg, opts)
}

func Ownership(root string, reg Registry, opts OwnershipOptions) (OwnershipReport, error) {
	if opts.SampleLimit < 0 {
		opts.SampleLimit = 0
	}
	refs := collectRepositoryOwnershipRefs(root, reg)
	report := OwnershipReport{
		SchemaVersion: 1,
		Summary: OwnershipSummary{
			Locations:   len(reg.Locations),
			SampleLimit: opts.SampleLimit,
		},
		Locations: make([]LocationOwnership, 0, len(reg.Locations)),
	}
	for _, location := range reg.Locations {
		entry, err := ownershipLocation(root, location, refs, opts.SampleLimit)
		if err != nil {
			return OwnershipReport{}, err
		}
		report.Locations = append(report.Locations, entry)
		report.Summary.Files += entry.Files
		report.Summary.OwnedFiles += entry.OwnedFiles
		report.Summary.UnownedFiles += entry.UnownedFiles
	}
	return report, nil
}

type ownershipRefs struct {
	files map[string]bool
	roots []string
}

func collectOwnershipRefs(root string, reg Registry) ownershipRefs {
	refs := ownershipRefs{files: map[string]bool{}}
	for _, atom := range reg.CapabilityAtoms {
		for _, dep := range atom.Dependencies {
			refs.addDependency(root, dep)
		}
	}
	for _, evidence := range reg.EvidenceSets {
		refs.add(root, evidence.ArtifactPath)
	}
	for _, boundary := range reg.VersionBoundaries {
		for _, dep := range boundary.Evidence {
			refs.addDependency(root, dep)
		}
	}
	sort.Strings(refs.roots)
	return refs
}

func collectRepositoryOwnershipRefs(root string, reg Registry) ownershipRefs {
	refs := collectOwnershipRefs(root, reg)
	for path := range refs.files {
		if dir, ok := registeredGeneratedMatrixRoot(path, reg.Locations); ok {
			refs.addExistingRoot(root, dir)
		}
	}
	sort.Strings(refs.roots)
	return refs
}

func (r *ownershipRefs) addDependency(root string, dep Dependency) {
	if dep.Kind != "glob" {
		r.add(root, dep.Path)
		return
	}
	matches, err := dependencyGlobMatches(root, dep.Path)
	if err != nil {
		return
	}
	for _, match := range matches {
		r.files[match] = true
	}
}

func (r *ownershipRefs) add(root, rel string) {
	if rel == "" || !relPathOK(rel) {
		return
	}
	clean := cleanRel(rel)
	info, err := os.Stat(filepath.Join(root, filepath.FromSlash(clean)))
	if err == nil && info.IsDir() {
		r.roots = append(r.roots, clean)
		return
	}
	r.files[clean] = true
}

func (r *ownershipRefs) addExistingRoot(root, rel string) {
	if rel == "" || !relPathOK(rel) {
		return
	}
	clean := cleanRel(rel)
	info, err := os.Stat(filepath.Join(root, filepath.FromSlash(clean)))
	if err != nil || !info.IsDir() {
		return
	}
	r.roots = append(r.roots, clean)
}

func registeredGeneratedMatrixRoot(path string, locations []Location) (string, bool) {
	clean := cleanRel(path)
	if filepath.Base(clean) != "matrix.json" {
		return "", false
	}
	dir := cleanRel(filepath.Dir(clean))
	if dir == "." || dir == clean {
		return "", false
	}
	for _, location := range locations {
		if location.Tracked || location.Class != "generated_evidence" {
			continue
		}
		root := cleanRel(location.Path)
		if dir == root || hasPathPrefix(dir, root) {
			return dir, true
		}
	}
	return "", false
}

func ownershipLocation(root string, location Location, refs ownershipRefs, sampleLimit int) (LocationOwnership, error) {
	entry := LocationOwnership{
		ID:    location.ID,
		Path:  location.Path,
		Class: location.Class,
	}
	groups := map[string]int{}
	path := filepath.Join(root, filepath.FromSlash(location.Path))
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return entry, nil
		}
		return LocationOwnership{}, err
	}
	if !info.IsDir() {
		entry.Files = 1
		if refs.owns(cleanRel(location.Path)) {
			entry.OwnedFiles = 1
		} else {
			entry.UnownedFiles = 1
			entry.addUnownedSample(cleanRel(location.Path), sampleLimit)
			groups[groupName(location.Path, location.Path)]++
		}
		entry.UnownedGroups = sortedUnownedGroups(groups)
		return entry, nil
	}
	err = filepath.WalkDir(path, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		clean := cleanRel(rel)
		entry.Files++
		if refs.owns(clean) {
			entry.OwnedFiles++
			return nil
		}
		entry.UnownedFiles++
		entry.addUnownedSample(clean, sampleLimit)
		groups[groupName(clean, location.Path)]++
		return nil
	})
	if err != nil {
		return LocationOwnership{}, err
	}
	entry.UnownedGroups = sortedUnownedGroups(groups)
	return entry, nil
}

func (r ownershipRefs) owns(path string) bool {
	clean := cleanRel(path)
	if r.files[clean] {
		return true
	}
	for _, root := range r.roots {
		if clean == root || hasPathPrefix(clean, root) {
			return true
		}
	}
	return false
}

func (l *LocationOwnership) addUnownedSample(path string, limit int) {
	if limit == 0 || len(l.UnownedSamples) >= limit {
		return
	}
	l.UnownedSamples = append(l.UnownedSamples, path)
}

func groupName(path, locationPath string) string {
	cleanPath := cleanRel(path)
	cleanLocation := cleanRel(locationPath)
	rel := cleanPath
	if cleanPath == cleanLocation {
		rel = filepath.Base(cleanPath)
	} else if hasPathPrefix(cleanPath, cleanLocation) {
		rel = cleanPath[len(cleanLocation)+1:]
	}
	parts := splitSlash(rel)
	if len(parts) > 1 {
		return parts[0]
	}
	name := parts[0]
	ext := filepath.Ext(name)
	if ext != "" {
		name = name[:len(name)-len(ext)]
	}
	if len(name) > len("minimal-") && name[:len("minimal-")] == "minimal-" {
		name = name[len("minimal-"):]
	}
	for i, r := range name {
		if r == '-' || r == '_' {
			if i > 0 {
				return name[:i]
			}
		}
	}
	if name == "" {
		return "."
	}
	return name
}

func splitSlash(path string) []string {
	var parts []string
	start := 0
	for i, r := range path {
		if r == '/' {
			if start < i {
				parts = append(parts, path[start:i])
			}
			start = i + 1
		}
	}
	if start < len(path) {
		parts = append(parts, path[start:])
	}
	if len(parts) == 0 {
		return []string{"."}
	}
	return parts
}

func sortedUnownedGroups(counts map[string]int) []UnownedGroup {
	if len(counts) == 0 {
		return nil
	}
	groups := make([]UnownedGroup, 0, len(counts))
	for name, files := range counts {
		groups = append(groups, UnownedGroup{Name: name, Files: files})
	}
	sort.Slice(groups, func(i, j int) bool {
		if groups[i].Files != groups[j].Files {
			return groups[i].Files > groups[j].Files
		}
		return groups[i].Name < groups[j].Name
	})
	return groups
}
