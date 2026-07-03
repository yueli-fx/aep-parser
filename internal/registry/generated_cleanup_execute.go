package registry

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type GeneratedCleanupExecutionOptions struct {
	Apply               bool
	PruneReviewSiblings bool
	IncludeProducers    []string
	ExcludeProducers    []string
}

type GeneratedCleanupExecutionReport struct {
	SchemaVersion int                              `json:"schema_version"`
	Mode          string                           `json:"mode"`
	Summary       GeneratedCleanupExecutionSummary `json:"summary"`
	Operations    []GeneratedCleanupExecutionOp    `json:"operations"`
}

type GeneratedCleanupExecutionSummary struct {
	Groups                int                               `json:"groups"`
	PlannedGroups         int                               `json:"planned_groups"`
	SkippedGroups         int                               `json:"skipped_groups"`
	DeletedGroups         int                               `json:"deleted_groups,omitempty"`
	Files                 int                               `json:"files"`
	PlannedFiles          int                               `json:"planned_files"`
	SkippedFiles          int                               `json:"skipped_files"`
	DeletedFiles          int                               `json:"deleted_files,omitempty"`
	DeleteDirectoryGroups int                               `json:"delete_directory_groups"`
	DeletePrefixGroups    int                               `json:"delete_prefix_groups"`
	Errors                int                               `json:"errors"`
	StatusBuckets         []GeneratedCleanupExecutionBucket `json:"status_buckets,omitempty"`
	SkippedReasonBuckets  []GeneratedCleanupExecutionBucket `json:"skipped_reason_buckets,omitempty"`
}

type GeneratedCleanupExecutionBucket struct {
	Name   string `json:"name"`
	Groups int    `json:"groups"`
	Files  int    `json:"files"`
}

type GeneratedCleanupExecutionOp struct {
	LocationID       string   `json:"location_id"`
	Group            string   `json:"group"`
	ProducerCategory string   `json:"producer_category"`
	CleanupOperation string   `json:"cleanup_operation"`
	CleanupTarget    string   `json:"cleanup_target"`
	PreservePaths    []string `json:"preserve_paths,omitempty"`
	Files            int      `json:"files"`
	Bytes            int64    `json:"bytes"`
	Status           string   `json:"status"`
	Reason           string   `json:"reason,omitempty"`
	Samples          []string `json:"samples,omitempty"`
	Error            string   `json:"error,omitempty"`
}

func ExecuteGeneratedCleanup(root string, report GeneratedCleanupReport, opts GeneratedCleanupExecutionOptions) GeneratedCleanupExecutionReport {
	mode := "dry_run"
	if opts.Apply {
		mode = "apply"
	}
	include := cleanupStringSet(opts.IncludeProducers)
	exclude := cleanupStringSet(opts.ExcludeProducers)
	out := GeneratedCleanupExecutionReport{
		SchemaVersion: 1,
		Mode:          mode,
	}
	statusBuckets := map[string]GeneratedCleanupExecutionBucket{}
	skippedReasonBuckets := map[string]GeneratedCleanupExecutionBucket{}
	locations := generatedCleanupLocationRoots(report)
	for _, location := range report.Locations {
		for _, group := range location.Groups {
			op := generatedCleanupExecutionOp(root, locations, group, include, exclude, opts.Apply, opts.PruneReviewSiblings)
			out.Operations = append(out.Operations, op)
			out.Summary.Groups++
			out.Summary.Files += group.Files
			addGeneratedCleanupExecutionBucket(statusBuckets, op.Status, op.Files)
			if op.Status == "skipped" {
				addGeneratedCleanupExecutionBucket(skippedReasonBuckets, valueOr(op.Reason, "unspecified"), op.Files)
			}
			switch op.Status {
			case "planned":
				out.Summary.PlannedGroups++
				out.Summary.PlannedFiles += op.Files
			case "deleted":
				out.Summary.DeletedGroups++
				out.Summary.DeletedFiles += op.Files
			case "skipped":
				out.Summary.SkippedGroups++
				out.Summary.SkippedFiles += op.Files
			case "error":
				out.Summary.Errors++
			}
			switch group.CleanupOperation {
			case "delete_directory_tree":
				out.Summary.DeleteDirectoryGroups++
			case "delete_file_prefix_matches":
				out.Summary.DeletePrefixGroups++
			}
		}
	}
	out.Summary.StatusBuckets = sortedGeneratedCleanupExecutionBuckets(statusBuckets)
	out.Summary.SkippedReasonBuckets = sortedGeneratedCleanupExecutionBuckets(skippedReasonBuckets)
	sort.Slice(out.Operations, func(i, j int) bool {
		left := out.Operations[i]
		right := out.Operations[j]
		if left.Status != right.Status {
			return left.Status < right.Status
		}
		if left.Files != right.Files {
			return left.Files > right.Files
		}
		if left.LocationID != right.LocationID {
			return left.LocationID < right.LocationID
		}
		return left.Group < right.Group
	})
	return out
}

func addGeneratedCleanupExecutionBucket(buckets map[string]GeneratedCleanupExecutionBucket, name string, files int) {
	bucket := buckets[name]
	bucket.Name = name
	bucket.Groups++
	bucket.Files += files
	buckets[name] = bucket
}

func sortedGeneratedCleanupExecutionBuckets(buckets map[string]GeneratedCleanupExecutionBucket) []GeneratedCleanupExecutionBucket {
	names := make([]string, 0, len(buckets))
	for name := range buckets {
		names = append(names, name)
	}
	sort.Strings(names)
	out := make([]GeneratedCleanupExecutionBucket, 0, len(names))
	for _, name := range names {
		out = append(out, buckets[name])
	}
	return out
}

func generatedCleanupExecutionOp(root string, locations []string, group GeneratedCleanupGroup, include, exclude map[string]bool, apply, pruneReviewSiblings bool) GeneratedCleanupExecutionOp {
	op := GeneratedCleanupExecutionOp{
		LocationID:       group.LocationID,
		Group:            group.Group,
		ProducerCategory: group.ProducerCategory,
		CleanupOperation: group.CleanupOperation,
		CleanupTarget:    group.CleanupTarget,
		PreservePaths:    append([]string(nil), group.PreservePaths...),
		Files:            group.UnreferencedFiles,
		Bytes:            group.Bytes,
		Samples:          group.DeleteSamples,
	}
	reviewPrune := pruneReviewSiblings &&
		group.CleanupOperation == "preserve_paths_then_review_unreferenced_siblings" &&
		group.UnreferencedFiles > 0 &&
		len(group.PreservePaths) > 0
	if len(include) > 0 && !include[group.ProducerCategory] {
		op.Status = "skipped"
		op.Reason = "producer_not_included"
		return op
	}
	if exclude[group.ProducerCategory] {
		op.Status = "skipped"
		op.Reason = "producer_excluded"
		return op
	}
	if group.Action != "cleanup_candidate" && !reviewPrune {
		op.Status = "skipped"
		op.Reason = "not_cleanup_candidate"
		return op
	}
	if group.CleanupOperation != "delete_directory_tree" &&
		group.CleanupOperation != "delete_file_prefix_matches" &&
		group.CleanupOperation != "preserve_paths_then_review_unreferenced_siblings" {
		op.Status = "skipped"
		op.Reason = "operation_not_deletable"
		return op
	}
	if !generatedCleanupTargetSafe(group.CleanupTarget, locations) {
		op.Status = "error"
		op.Error = "cleanup target is outside generated cleanup locations"
		return op
	}
	if !apply {
		op.Status = "planned"
		op.Reason = "dry_run"
		return op
	}
	if err := applyGeneratedCleanupOperation(root, group); err != nil {
		op.Status = "error"
		op.Error = err.Error()
		return op
	}
	op.Status = "deleted"
	return op
}

func generatedCleanupLocationRoots(report GeneratedCleanupReport) []string {
	roots := make([]string, 0, len(report.Locations))
	for _, location := range report.Locations {
		roots = append(roots, cleanRel(location.Path))
	}
	sort.Strings(roots)
	return roots
}

func generatedCleanupTargetSafe(target string, roots []string) bool {
	clean := strings.TrimSuffix(cleanRel(target), "*")
	if clean == "." || clean == "" || !relPathOK(clean) {
		return false
	}
	for _, root := range roots {
		if clean == root || hasPathPrefix(clean, root) {
			return true
		}
	}
	return false
}

func applyGeneratedCleanupOperation(root string, group GeneratedCleanupGroup) error {
	switch group.CleanupOperation {
	case "delete_directory_tree":
		target, err := safeCleanupPath(root, group.CleanupTarget)
		if err != nil {
			return err
		}
		return os.RemoveAll(target)
	case "delete_file_prefix_matches":
		target := strings.TrimSuffix(group.CleanupTarget, "*")
		if _, err := safeCleanupPath(root, target); err != nil {
			return err
		}
		pattern := filepath.Join(root, filepath.FromSlash(group.CleanupTarget))
		matches, err := filepath.Glob(pattern)
		if err != nil {
			return err
		}
		for _, match := range matches {
			info, err := os.Stat(match)
			if err != nil {
				if os.IsNotExist(err) {
					continue
				}
				return err
			}
			if info.IsDir() {
				continue
			}
			if err := os.Remove(match); err != nil {
				return err
			}
		}
		return nil
	case "preserve_paths_then_review_unreferenced_siblings":
		if strings.HasSuffix(group.CleanupTarget, "*") {
			return applyGeneratedCleanupPreservePathPrefix(root, group)
		}
		return applyGeneratedCleanupPreservePaths(root, group)
	default:
		return fmt.Errorf("unsupported cleanup operation %q", group.CleanupOperation)
	}
}

func applyGeneratedCleanupPreservePaths(root string, group GeneratedCleanupGroup) error {
	rootAbs, err := filepath.Abs(root)
	if err != nil {
		return err
	}
	target, err := safeCleanupPath(root, group.CleanupTarget)
	if err != nil {
		return err
	}
	preserve := map[string]bool{}
	for _, path := range group.PreservePaths {
		clean := cleanRel(path)
		if clean == group.CleanupTarget || hasPathPrefix(clean, group.CleanupTarget) {
			preserve[clean] = true
			continue
		}
		return fmt.Errorf("preserve path %q is outside cleanup target %q", path, group.CleanupTarget)
	}
	var dirs []string
	err = filepath.WalkDir(target, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			dirs = append(dirs, path)
			return nil
		}
		rel, err := filepath.Rel(rootAbs, path)
		if err != nil {
			return err
		}
		if preserve[cleanRel(rel)] {
			return nil
		}
		return os.Remove(path)
	})
	if err != nil {
		return err
	}
	sort.Slice(dirs, func(i, j int) bool {
		return len(dirs[i]) > len(dirs[j])
	})
	for _, dir := range dirs {
		if dir == target {
			continue
		}
		entries, err := os.ReadDir(dir)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return err
		}
		if len(entries) == 0 {
			if err := os.Remove(dir); err != nil && !os.IsNotExist(err) {
				return err
			}
		}
	}
	return nil
}

func applyGeneratedCleanupPreservePathPrefix(root string, group GeneratedCleanupGroup) error {
	prefix := strings.TrimSuffix(group.CleanupTarget, "*")
	if _, err := safeCleanupPath(root, prefix); err != nil {
		return err
	}
	preserve := map[string]bool{}
	for _, path := range group.PreservePaths {
		clean := cleanRel(path)
		if strings.HasPrefix(clean, cleanRel(prefix)) {
			preserve[clean] = true
			continue
		}
		return fmt.Errorf("preserve path %q is outside cleanup target %q", path, group.CleanupTarget)
	}
	rootAbs, err := filepath.Abs(root)
	if err != nil {
		return err
	}
	pattern := filepath.Join(rootAbs, filepath.FromSlash(group.CleanupTarget))
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return err
	}
	for _, match := range matches {
		info, err := os.Stat(match)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return err
		}
		if info.IsDir() {
			continue
		}
		rel, err := filepath.Rel(rootAbs, match)
		if err != nil {
			return err
		}
		if preserve[cleanRel(rel)] {
			continue
		}
		if err := os.Remove(match); err != nil {
			return err
		}
	}
	return nil
}

func safeCleanupPath(root, rel string) (string, error) {
	clean := cleanRel(rel)
	if !relPathOK(clean) {
		return "", fmt.Errorf("invalid cleanup target %q", rel)
	}
	rootAbs, err := filepath.Abs(root)
	if err != nil {
		return "", err
	}
	targetAbs, err := filepath.Abs(filepath.Join(rootAbs, filepath.FromSlash(clean)))
	if err != nil {
		return "", err
	}
	relToRoot, err := filepath.Rel(rootAbs, targetAbs)
	if err != nil {
		return "", err
	}
	if relToRoot == "." || strings.HasPrefix(relToRoot, ".."+string(filepath.Separator)) || relToRoot == ".." || filepath.IsAbs(relToRoot) {
		return "", fmt.Errorf("cleanup target escapes repository root: %q", rel)
	}
	return targetAbs, nil
}

func cleanupStringSet(values []string) map[string]bool {
	out := map[string]bool{}
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		out[value] = true
	}
	return out
}
