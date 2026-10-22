package server

import (
	"io"
	"log"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAPIValidationAndUI(t *testing.T) {
	handler := New(log.New(io.Discard, "", 0)).Handler()
	cases := []struct {
		method, path, body, origin string
		status                     int
	}{
		{"GET", "/", "", "", 200}, {"GET", "/app.js", "", "", 200}, {"GET", "/api/health", "", "", 200},
		{"POST", "/api/convert", `{"html":""}`, "", 400},
		{"POST", "/api/convert", `{"html":"<p>x</p>","options":{"margin_mm":-1}}`, "", 400},
		{"POST", "/api/convert", `{} {}`, "", 400},
		{"POST", "/api/convert", `{}`, "https://untrusted.example", 403},
	}
	for _, tc := range cases {
		t.Run(tc.method+tc.path+tc.body, func(t *testing.T) {
			r := httptest.NewRequest(tc.method, tc.path, strings.NewReader(tc.body))
			if tc.origin != "" {
				r.Header.Set("Origin", tc.origin)
			}
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, r)
			if w.Code != tc.status {
				t.Errorf("got %d: %s", w.Code, w.Body.String())
			}
		})
	}
}

func TestLocalBatchArtifacts(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "report.html"), []byte("batch report evidence"), 0600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("HTMLWORD_BATCH_DIR", dir)
	handler := New(log.New(io.Discard, "", 0)).Handler()
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, httptest.NewRequest("GET", "/batch/report.html", nil))
	if w.Code != 200 || !strings.Contains(w.Body.String(), "batch report evidence") {
		t.Fatalf("batch report unavailable: %d %s", w.Code, w.Body.String())
	}
}
