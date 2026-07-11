package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/yueli-fx/aep-parser/internal/aep"
)

func TestHandlerHealth(t *testing.T) {
	handler := NewHandler(Options{})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	var got map[string]string
	decodeJSON(t, rec.Body.Bytes(), &got)
	if got["status"] != "ok" {
		t.Fatalf("health = %#v", got)
	}
}

func TestHandlerCapabilitiesReportsPureServiceAndUnavailableAE(t *testing.T) {
	handler := NewHandler(Options{})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/capabilities", nil)

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	var got CapabilityReport
	decodeJSON(t, rec.Body.Bytes(), &got)
	want := map[string]string{
		"parser":       "available",
		"profile":      "available",
		"recipe_draft": "available",
		"ae_readback":  "unavailable",
		"render":       "unavailable",
		"pixel_diff":   "available",
	}
	if got.Capabilities == nil {
		t.Fatal("capabilities missing")
	}
	for key, value := range want {
		if got.Capabilities[key] != value {
			t.Fatalf("capability %s = %q, want %q in %#v", key, got.Capabilities[key], value, got.Capabilities)
		}
	}
}

func TestHandlerParseAcceptsJSONPath(t *testing.T) {
	path := writeMinimalAEP(t)
	handler := NewHandler(Options{AllowPathInput: true, AllowedPathRoots: []string{filepath.Dir(path)}})
	body := bytes.NewBufferString(`{"path":` + strconvQuote(path) + `}`)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/parse", body)
	req.Header.Set("Content-Type", "application/json")

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	var got ParseResponse
	decodeJSON(t, rec.Body.Bytes(), &got)
	if got.SchemaVersion != 1 || got.Source.Path != path {
		t.Fatalf("parse response = %+v", got)
	}
	if got.Summary.Compositions != 1 {
		t.Fatalf("summary = %+v", got.Summary)
	}
	if got.Project == nil {
		t.Fatal("project JSON missing")
	}
}

func TestHandlerRejectsPathOutsideAllowedRoots(t *testing.T) {
	path := writeMinimalAEP(t)
	handler := NewHandler(Options{AllowPathInput: true, AllowedPathRoots: []string{t.TempDir()}})
	body := bytes.NewBufferString(`{"path":` + strconvQuote(path) + `}`)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/parse", body)
	req.Header.Set("Content-Type", "application/json")

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest || !bytes.Contains(rec.Body.Bytes(), []byte("outside allowed roots")) {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
}

func TestHandlerRejectsSymlinkEscapingAllowedRoot(t *testing.T) {
	outside := writeMinimalAEP(t)
	root := t.TempDir()
	link := filepath.Join(root, "escape.aep")
	if err := os.Symlink(outside, link); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}
	handler := NewHandler(Options{AllowPathInput: true, AllowedPathRoots: []string{root}})
	body := bytes.NewBufferString(`{"path":` + strconvQuote(link) + `}`)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/parse", body)
	req.Header.Set("Content-Type", "application/json")

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest || !bytes.Contains(rec.Body.Bytes(), []byte("outside allowed roots")) {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
}

func TestRecoverHTTPContainsPanics(t *testing.T) {
	handler := recoverHTTP(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		panic("sensitive detail")
	}))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Code != http.StatusInternalServerError || bytes.Contains(rec.Body.Bytes(), []byte("sensitive detail")) {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
}

func TestLimitConcurrentRejectsExcessRequests(t *testing.T) {
	entered := make(chan struct{})
	release := make(chan struct{})
	handler := limitConcurrent(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		close(entered)
		<-release
		w.WriteHeader(http.StatusNoContent)
	}), 1)
	firstDone := make(chan struct{})
	go func() {
		defer close(firstDone)
		handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))
	}()
	select {
	case <-entered:
	case <-time.After(time.Second):
		t.Fatal("first request did not enter handler")
	}
	second := httptest.NewRecorder()
	handler.ServeHTTP(second, httptest.NewRequest(http.MethodGet, "/", nil))
	if second.Code != http.StatusServiceUnavailable {
		t.Fatalf("second status = %d, want 503", second.Code)
	}
	close(release)
	<-firstDone
}

func TestHandlerRejectsJSONPathByDefault(t *testing.T) {
	path := writeMinimalAEP(t)
	handler := NewHandler(Options{})
	body := bytes.NewBufferString(`{"path":` + strconvQuote(path) + `}`)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/parse", body)
	req.Header.Set("Content-Type", "application/json")

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	var got ErrorResponse
	decodeJSON(t, rec.Body.Bytes(), &got)
	if got.Error != "path input is disabled" {
		t.Fatalf("error = %q", got.Error)
	}
}

func TestHandlerProfileAcceptsUploadedBytes(t *testing.T) {
	data := readMinimalAEPBytes(t)
	handler := NewHandler(Options{})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/profile", bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/octet-stream")
	req.Header.Set("X-AEP-Path", "uploaded.aep")

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	var got ProfileResponse
	decodeJSON(t, rec.Body.Bytes(), &got)
	if got.SchemaVersion != 1 || got.Source.Path != "uploaded.aep" {
		t.Fatalf("profile response = %+v", got)
	}
	if got.Profile == nil || got.Profile.Fingerprint.CompCount != 1 {
		t.Fatalf("profile = %+v", got.Profile)
	}
}

func TestHandlerRejectsUnsupportedMethodAndMalformedInput(t *testing.T) {
	handler := NewHandler(Options{})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/parse", nil)

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("GET /parse status = %d, body = %s", rec.Code, rec.Body.String())
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/profile", bytes.NewBufferString(`{"path":""}`))
	req.Header.Set("Content-Type", "application/json")
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("bad profile status = %d, body = %s", rec.Code, rec.Body.String())
	}
}

func writeMinimalAEP(t *testing.T) string {
	t.Helper()
	project := aep.NewProject()
	if _, err := aep.NewComposition(project, "Main", 320, 180, 24, 1); err != nil {
		t.Fatalf("NewComposition: %v", err)
	}
	path := filepath.Join(t.TempDir(), "minimal.aep")
	file, err := os.Create(path)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	defer file.Close()
	if err := project.WriteAEP(file); err != nil {
		t.Fatalf("WriteAEP: %v", err)
	}
	return path
}

func readMinimalAEPBytes(t *testing.T) []byte {
	t.Helper()
	data, err := os.ReadFile(writeMinimalAEP(t))
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	return data
}

func decodeJSON(t *testing.T, data []byte, out any) {
	t.Helper()
	if err := json.Unmarshal(data, out); err != nil {
		t.Fatalf("Unmarshal %s: %v", string(data), err)
	}
}

func strconvQuote(value string) string {
	data, _ := json.Marshal(value)
	return string(data)
}
