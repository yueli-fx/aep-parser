package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/yueli-fx/aep-parser/internal/aep"
	"github.com/yueli-fx/aep-parser/internal/profile"
)

const (
	SchemaVersion        = 1
	defaultMaxBodyBytes  = 256 << 20
	defaultMaxConcurrent = 32
)

type Options struct {
	MaxBodyBytes     int64
	MaxConcurrent    int
	AllowPathInput   bool
	AllowedPathRoots []string
}

type CapabilityReport struct {
	SchemaVersion int               `json:"schema_version"`
	Capabilities  map[string]string `json:"capabilities"`
}

type Source struct {
	Path string `json:"path,omitempty"`
	Mode string `json:"mode"`
}

type ProjectSummary struct {
	Compositions int `json:"compositions"`
	Footage      int `json:"footage"`
	Folders      int `json:"folders"`
	Warnings     int `json:"warnings"`
}

type ParseResponse struct {
	SchemaVersion int            `json:"schema_version"`
	Source        Source         `json:"source"`
	Summary       ProjectSummary `json:"summary"`
	Project       *aep.Project   `json:"project,omitempty"`
}

type ProfileResponse struct {
	SchemaVersion int              `json:"schema_version"`
	Source        Source           `json:"source"`
	Summary       ProjectSummary   `json:"summary"`
	Profile       *profile.Profile `json:"profile,omitempty"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type handler struct {
	opts Options
}

func NewHandler(opts Options) http.Handler {
	if opts.MaxBodyBytes <= 0 {
		opts.MaxBodyBytes = defaultMaxBodyBytes
	}
	if opts.MaxConcurrent <= 0 {
		opts.MaxConcurrent = defaultMaxConcurrent
	}
	mux := http.NewServeMux()
	h := handler{opts: opts}
	mux.HandleFunc("/health", h.health)
	mux.HandleFunc("/capabilities", h.capabilities)
	mux.HandleFunc("/parse", h.parse)
	mux.HandleFunc("/profile", h.profile)
	return recoverHTTP(limitConcurrent(mux, opts.MaxConcurrent))
}

func (h handler) health(w http.ResponseWriter, r *http.Request) {
	if !allowMethod(w, r, http.MethodGet) {
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h handler) capabilities(w http.ResponseWriter, r *http.Request) {
	if !allowMethod(w, r, http.MethodGet) {
		return
	}
	writeJSON(w, http.StatusOK, CapabilityReport{
		SchemaVersion: SchemaVersion,
		Capabilities: map[string]string{
			"parser":       "available",
			"profile":      "available",
			"recipe_draft": "available",
			"ae_readback":  "unavailable",
			"render":       "unavailable",
			"pixel_diff":   "available",
		},
	})
}

func (h handler) parse(w http.ResponseWriter, r *http.Request) {
	if !allowMethod(w, r, http.MethodPost) {
		return
	}
	project, source, err := h.readProject(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	writeJSON(w, http.StatusOK, ParseResponse{
		SchemaVersion: SchemaVersion,
		Source:        source,
		Summary:       summarizeProject(project),
		Project:       project,
	})
}

func (h handler) profile(w http.ResponseWriter, r *http.Request) {
	if !allowMethod(w, r, http.MethodPost) {
		return
	}
	project, source, err := h.readProject(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	prof, err := profile.Build(project, profile.Options{Path: source.Path})
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, ProfileResponse{
		SchemaVersion: SchemaVersion,
		Source:        source,
		Summary:       summarizeProject(project),
		Profile:       prof,
	})
}

func (h handler) readProject(r *http.Request) (*aep.Project, Source, error) {
	if strings.Contains(r.Header.Get("Content-Type"), "application/json") {
		if !h.opts.AllowPathInput {
			return nil, Source{}, fmt.Errorf("path input is disabled")
		}
		var input struct {
			Path string `json:"path"`
		}
		if err := json.NewDecoder(io.LimitReader(r.Body, h.opts.MaxBodyBytes)).Decode(&input); err != nil {
			return nil, Source{}, fmt.Errorf("decode request: %w", err)
		}
		if input.Path == "" {
			return nil, Source{}, fmt.Errorf("path is required")
		}
		path, err := allowedPath(input.Path, h.opts.AllowedPathRoots)
		if err != nil {
			return nil, Source{}, err
		}
		project, err := aep.Open(path)
		if err != nil {
			return nil, Source{}, fmt.Errorf("open %q: %w", path, err)
		}
		return project, Source{Path: path, Mode: "path"}, nil
	}
	data, err := io.ReadAll(io.LimitReader(r.Body, h.opts.MaxBodyBytes+1))
	if err != nil {
		return nil, Source{}, fmt.Errorf("read upload: %w", err)
	}
	if int64(len(data)) > h.opts.MaxBodyBytes {
		return nil, Source{}, fmt.Errorf("request body exceeds %d bytes", h.opts.MaxBodyBytes)
	}
	if len(data) == 0 {
		return nil, Source{}, fmt.Errorf("request body is required")
	}
	project, err := aep.FromReader(bytes.NewReader(data))
	if err != nil {
		return nil, Source{}, fmt.Errorf("parse upload: %w", err)
	}
	return project, Source{Path: r.Header.Get("X-AEP-Path"), Mode: "upload"}, nil
}

func allowedPath(path string, roots []string) (string, error) {
	if len(roots) == 0 {
		return "", fmt.Errorf("path input requires at least one allowed root")
	}
	resolvedPath, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("resolve path: %w", err)
	}
	resolvedPath, err = filepath.EvalSymlinks(resolvedPath)
	if err != nil {
		return "", fmt.Errorf("resolve path symlinks: %w", err)
	}
	for _, root := range roots {
		resolvedRoot, err := filepath.Abs(root)
		if err != nil {
			continue
		}
		if evaluated, err := filepath.EvalSymlinks(resolvedRoot); err == nil {
			resolvedRoot = evaluated
		}
		rel, err := filepath.Rel(resolvedRoot, resolvedPath)
		if err == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) && !filepath.IsAbs(rel) {
			return resolvedPath, nil
		}
	}
	return "", fmt.Errorf("path %q is outside allowed roots", path)
}

func recoverHTTP(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if recover() != nil {
				writeError(w, http.StatusInternalServerError, fmt.Errorf("internal server error"))
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func limitConcurrent(next http.Handler, max int) http.Handler {
	semaphore := make(chan struct{}, max)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case semaphore <- struct{}{}:
			defer func() { <-semaphore }()
			next.ServeHTTP(w, r)
		default:
			writeError(w, http.StatusServiceUnavailable, fmt.Errorf("server is busy"))
		}
	})
}

func summarizeProject(project *aep.Project) ProjectSummary {
	if project == nil {
		return ProjectSummary{}
	}
	return ProjectSummary{
		Compositions: len(project.Compositions),
		Footage:      len(project.Footage),
		Folders:      len(project.Folders),
		Warnings:     len(project.Warnings),
	}
}

func allowMethod(w http.ResponseWriter, r *http.Request, method string) bool {
	if r.Method == method {
		return true
	}
	w.Header().Set("Allow", method)
	writeError(w, http.StatusMethodNotAllowed, fmt.Errorf("method %s is not allowed", r.Method))
	return false
}

func writeError(w http.ResponseWriter, status int, err error) {
	writeJSON(w, status, ErrorResponse{Error: err.Error()})
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	_ = enc.Encode(value)
}
