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
	ID             string   `json:"id"`
	Path           string   `json:"path"`
	Class          string   `json:"class"`
	Files          int      `json:"files"`
	OwnedFiles     int      `json:"owned_files"`
	UnownedFiles   int      `json:"unowned_files"`
	UnownedSamples []string `json:"unowned_samples,omitempty"`
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
	refs := collectOwnershipRefs(root, reg)
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
			refs.add(root, dep.Path)
		}
	}
	for _, evidence := range reg.EvidenceSets {
		refs.add(root, evidence.ArtifactPath)
	}
	sort.Strings(refs.roots)
	return refs
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

func ownershipLocation(root string, location Location, refs ownershipRefs, sampleLimit int) (LocationOwnership, error) {
	entry := LocationOwnership{
		ID:    location.ID,
		Path:  location.Path,
		Class: location.Class,
	}
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
		}
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
		return nil
	})
	if err != nil {
		return LocationOwnership{}, err
	}
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
