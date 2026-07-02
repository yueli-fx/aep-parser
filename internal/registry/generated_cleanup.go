package registry

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type GeneratedCleanupOptions struct {
	SampleLimit int
}

type GeneratedCleanupReport struct {
	SchemaVersion int                        `json:"schema_version"`
	Summary       GeneratedCleanupSummary    `json:"summary"`
	Locations     []GeneratedCleanupLocation `json:"locations"`
}

type GeneratedCleanupSummary struct {
	Locations             int                    `json:"locations"`
	Groups                int                    `json:"groups"`
	Files                 int                    `json:"files"`
	Bytes                 int64                  `json:"bytes"`
	ReferencedFiles       int                    `json:"referenced_files"`
	UnreferencedFiles     int                    `json:"unreferenced_files"`
	CleanupCandidateFiles int                    `json:"cleanup_candidate_files"`
	RetainRegisteredFiles int                    `json:"retain_registered_files"`
	MixedGroups           int                    `json:"mixed_groups"`
	UnknownProducerGroups int                    `json:"unknown_producer_groups"`
	SampleLimit           int                    `json:"sample_limit"`
	ActionBuckets         []GeneratedGroupBucket `json:"action_buckets,omitempty"`
	ProducerBuckets       []GeneratedGroupBucket `json:"producer_buckets,omitempty"`
}

type GeneratedCleanupLocation struct {
	ID        string                  `json:"id"`
	Path      string                  `json:"path"`
	Class     string                  `json:"class"`
	Lifecycle string                  `json:"lifecycle"`
	Exists    bool                    `json:"exists"`
	Files     int                     `json:"files"`
	Bytes     int64                   `json:"bytes"`
	Groups    []GeneratedCleanupGroup `json:"groups,omitempty"`
}

type GeneratedCleanupGroup struct {
	LocationID        string   `json:"location_id"`
	Group             string   `json:"group"`
	PathPrefix        string   `json:"path_prefix"`
	ProducerCategory  string   `json:"producer_category"`
	ProducerWorkflows []string `json:"producer_workflows,omitempty"`
	Action            string   `json:"action"`
	CleanupSafety     string   `json:"cleanup_safety"`
	CleanupOperation  string   `json:"cleanup_operation"`
	CleanupTarget     string   `json:"cleanup_target,omitempty"`
	Reason            string   `json:"reason"`
	Files             int      `json:"files"`
	Bytes             int64    `json:"bytes"`
	ReferencedFiles   int      `json:"referenced_files"`
	UnreferencedFiles int      `json:"unreferenced_files"`
	EvidenceIDs       []string `json:"evidence_ids,omitempty"`
	PreservePaths     []string `json:"preserve_paths,omitempty"`
	Samples           []string `json:"samples,omitempty"`
	DeleteSamples     []string `json:"delete_samples,omitempty"`
}

type GeneratedGroupBucket struct {
	Name   string `json:"name"`
	Groups int    `json:"groups"`
	Files  int    `json:"files"`
}

func GeneratedCleanupRepository(root string, opts GeneratedCleanupOptions) (GeneratedCleanupReport, error) {
	reg, err := Load(root)
	if err != nil {
		return GeneratedCleanupReport{}, err
	}
	return GeneratedCleanup(root, reg, opts)
}

func GeneratedCleanup(root string, reg Registry, opts GeneratedCleanupOptions) (GeneratedCleanupReport, error) {
	if opts.SampleLimit < 0 {
		opts.SampleLimit = 0
	}
	refs := collectOwnershipRefs(root, reg)
	evidenceByGroup := generatedEvidenceByGroup(reg)
	report := GeneratedCleanupReport{
		SchemaVersion: 1,
		Summary: GeneratedCleanupSummary{
			SampleLimit: opts.SampleLimit,
		},
	}
	actionBuckets := map[string]GeneratedGroupBucket{}
	producerBuckets := map[string]GeneratedGroupBucket{}
	for _, location := range reg.Locations {
		if !isGeneratedCleanupLocation(location) {
			continue
		}
		entry, err := generatedCleanupLocation(root, location, refs, evidenceByGroup, opts.SampleLimit)
		if err != nil {
			return GeneratedCleanupReport{}, err
		}
		report.Locations = append(report.Locations, entry)
		report.Summary.Locations++
		report.Summary.Files += entry.Files
		report.Summary.Bytes += entry.Bytes
		for _, group := range entry.Groups {
			report.Summary.Groups++
			report.Summary.ReferencedFiles += group.ReferencedFiles
			report.Summary.UnreferencedFiles += group.UnreferencedFiles
			switch group.Action {
			case "cleanup_candidate":
				report.Summary.CleanupCandidateFiles += group.UnreferencedFiles
			case "retain_registered_evidence":
				report.Summary.RetainRegisteredFiles += group.ReferencedFiles
			case "review_mixed_registered_generated":
				report.Summary.RetainRegisteredFiles += group.ReferencedFiles
				report.Summary.CleanupCandidateFiles += group.UnreferencedFiles
				report.Summary.MixedGroups++
			}
			if group.ProducerCategory == "unknown_generated" {
				report.Summary.UnknownProducerGroups++
			}
			addGeneratedBucket(actionBuckets, group.Action, group.Files)
			addGeneratedBucket(producerBuckets, group.ProducerCategory, group.Files)
		}
	}
	report.Summary.ActionBuckets = sortedGeneratedBuckets(actionBuckets)
	report.Summary.ProducerBuckets = sortedGeneratedBuckets(producerBuckets)
	return report, nil
}

func isGeneratedCleanupLocation(location Location) bool {
	return !location.Tracked && location.Class == "generated_evidence"
}

func generatedCleanupLocation(root string, location Location, refs ownershipRefs, evidence map[string][]string, sampleLimit int) (GeneratedCleanupLocation, error) {
	entry := GeneratedCleanupLocation{
		ID:        location.ID,
		Path:      location.Path,
		Class:     location.Class,
		Lifecycle: location.Lifecycle,
	}
	path := filepath.Join(root, filepath.FromSlash(location.Path))
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return entry, nil
		}
		return GeneratedCleanupLocation{}, err
	}
	entry.Exists = true
	groups := map[string]*GeneratedCleanupGroup{}
	addFile := func(rel string, size int64) {
		clean := cleanRel(rel)
		groupName := groupName(clean, location.Path)
		group := groups[groupName]
		if group == nil {
			category, workflows := inferGeneratedProducer(location.ID, groupName)
			group = &GeneratedCleanupGroup{
				LocationID:        location.ID,
				Group:             groupName,
				PathPrefix:        generatedGroupPathPrefix(location.Path, clean, groupName),
				ProducerCategory:  category,
				ProducerWorkflows: workflows,
				EvidenceIDs:       append([]string(nil), evidence[generatedEvidenceGroupKey(location.ID, groupName)]...),
			}
			groups[groupName] = group
		}
		group.Files++
		group.Bytes += size
		if refs.owns(clean) {
			group.ReferencedFiles++
			group.PreservePaths = append(group.PreservePaths, clean)
		} else {
			group.UnreferencedFiles++
			if sampleLimit != 0 && len(group.DeleteSamples) < sampleLimit {
				group.DeleteSamples = append(group.DeleteSamples, clean)
			}
		}
		if sampleLimit != 0 && len(group.Samples) < sampleLimit {
			group.Samples = append(group.Samples, clean)
		}
	}
	if !info.IsDir() {
		entry.Files = 1
		entry.Bytes = info.Size()
		addFile(location.Path, info.Size())
	} else {
		err = filepath.WalkDir(path, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				return nil
			}
			info, err := d.Info()
			if err != nil {
				return err
			}
			rel, err := filepath.Rel(root, path)
			if err != nil {
				return err
			}
			entry.Files++
			entry.Bytes += info.Size()
			addFile(rel, info.Size())
			return nil
		})
		if err != nil {
			return GeneratedCleanupLocation{}, err
		}
	}
	entry.Groups = finalizeGeneratedGroups(groups)
	return entry, nil
}

func generatedEvidenceByGroup(reg Registry) map[string][]string {
	out := map[string][]string{}
	for _, evidence := range reg.EvidenceSets {
		for _, location := range reg.Locations {
			if !isGeneratedCleanupLocation(location) {
				continue
			}
			path := cleanRel(evidence.ArtifactPath)
			root := cleanRel(location.Path)
			if path != root && !hasPathPrefix(path, root) {
				continue
			}
			group := groupName(path, location.Path)
			key := generatedEvidenceGroupKey(location.ID, group)
			out[key] = append(out[key], evidence.ID)
		}
	}
	for key := range out {
		sort.Strings(out[key])
	}
	return out
}

func generatedEvidenceGroupKey(locationID, group string) string {
	return locationID + "\x00" + group
}

func generatedGroupPathPrefix(locationPath, filePath, group string) string {
	location := cleanRel(locationPath)
	clean := cleanRel(filePath)
	if clean == location {
		return location
	}
	if !hasPathPrefix(clean, location) {
		return location + "/" + group
	}
	rel := clean[len(location)+1:]
	parts := splitSlash(rel)
	if len(parts) > 1 {
		return location + "/" + parts[0]
	}
	return location + "/" + group + "*"
}

func finalizeGeneratedGroups(groups map[string]*GeneratedCleanupGroup) []GeneratedCleanupGroup {
	out := make([]GeneratedCleanupGroup, 0, len(groups))
	for _, group := range groups {
		group.Action, group.CleanupSafety, group.Reason = generatedCleanupDecision(*group)
		sort.Strings(group.PreservePaths)
		group.CleanupOperation, group.CleanupTarget = generatedCleanupOperation(*group)
		out = append(out, *group)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Files != out[j].Files {
			return out[i].Files > out[j].Files
		}
		if out[i].LocationID != out[j].LocationID {
			return out[i].LocationID < out[j].LocationID
		}
		return out[i].Group < out[j].Group
	})
	return out
}

func generatedCleanupDecision(group GeneratedCleanupGroup) (action, safety, reason string) {
	if group.ReferencedFiles > 0 && group.UnreferencedFiles > 0 {
		return "review_mixed_registered_generated", "mixed_registered_generated", "group contains registered evidence plus unreferenced generated siblings"
	}
	if group.ReferencedFiles > 0 {
		return "retain_registered_evidence", "registered_generated_evidence", "all files in group are referenced by registry evidence or atom dependencies"
	}
	return "cleanup_candidate", "unreferenced_rebuildable_generated", "no registry evidence or atom dependency references this generated group"
}

func generatedCleanupOperation(group GeneratedCleanupGroup) (operation, target string) {
	switch group.Action {
	case "cleanup_candidate":
		if strings.HasSuffix(group.PathPrefix, "*") {
			return "delete_file_prefix_matches", group.PathPrefix
		}
		return "delete_directory_tree", group.PathPrefix
	case "review_mixed_registered_generated":
		return "preserve_paths_then_review_unreferenced_siblings", group.PathPrefix
	case "retain_registered_evidence":
		return "no_delete_registered_evidence", ""
	default:
		return "review_unknown", group.PathPrefix
	}
}

func inferGeneratedProducer(locationID, group string) (string, []string) {
	switch locationID {
	case "generated_test_data":
		switch group {
		case "args":
			return "host_open_args", []string{"host_open"}
		case "fixtures":
			return "generated_reverse_fixtures", []string{"generate", "parse", "host_open"}
		case "ship-gate":
			return "ship_gate_fixtures", []string{"generate", "host_open"}
		default:
			return "generated_test_fixture", []string{"generate"}
		}
	case "tmp_evidence", "tmp":
		return inferTmpProducer(group)
	default:
		return "unknown_generated", nil
	}
}

func inferTmpProducer(group string) (string, []string) {
	switch {
	case group == "migration":
		return "migration_report", []string{"migrate"}
	case group == "host":
		return "host_open_gap", []string{"host_open"}
	case strings.HasPrefix(group, "migration_matrix"):
		return "version_matrix", []string{"migrate", "host_open"}
	case strings.HasPrefix(group, "host_open"):
		return "host_open_gap", []string{"host_open"}
	case strings.HasPrefix(group, "migration_"):
		return "migration_probe", []string{"migrate", "host_open"}
	case strings.HasPrefix(group, "explicit_matte"):
		return "migration_probe", []string{"migrate", "host_open"}
	case strings.HasPrefix(group, "probe"):
		return "migration_probe", []string{"migrate", "profile"}
	case strings.HasPrefix(group, "technique"):
		return "technique_learning", []string{"learn", "profile", "ai_generate"}
	case strings.HasPrefix(group, "aepselfhost"):
		return "selfhost_learning", []string{"learn", "host_open"}
	case strings.HasPrefix(group, "registry"):
		return "registry_report", nil
	case strings.HasPrefix(group, "coverage"), strings.HasPrefix(group, "current"):
		return "coverage_scratch", []string{"migrate"}
	case strings.HasPrefix(group, "inspect"):
		return "profile_probe", []string{"profile"}
	default:
		return "unknown_generated", nil
	}
}

func addGeneratedBucket(buckets map[string]GeneratedGroupBucket, name string, files int) {
	bucket := buckets[name]
	bucket.Name = name
	bucket.Groups++
	bucket.Files += files
	buckets[name] = bucket
}

func sortedGeneratedBuckets(buckets map[string]GeneratedGroupBucket) []GeneratedGroupBucket {
	if len(buckets) == 0 {
		return nil
	}
	out := make([]GeneratedGroupBucket, 0, len(buckets))
	for _, bucket := range buckets {
		out = append(out, bucket)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Files != out[j].Files {
			return out[i].Files > out[j].Files
		}
		return out[i].Name < out[j].Name
	})
	return out
}
