package main

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

const compactSpec = `{"openapi":"3.0.3","paths":{}}`

func prepare(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "spec"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Chdir(dir)
	return dir
}

func TestSyncWritesNormalizedSnapshotFromFile(t *testing.T) {
	dir := prepare(t)
	source := filepath.Join(dir, "source.json")
	if err := os.WriteFile(source, []byte(compactSpec), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv(sourceVariable, source)
	if err := run(false); err != nil {
		t.Fatal(err)
	}
	written, err := os.ReadFile(snapshotPath)
	if err != nil {
		t.Fatal(err)
	}
	want := "{\n  \"openapi\": \"3.0.3\",\n  \"paths\": {}\n}\n"
	if string(written) != want {
		t.Fatalf("snapshot = %q", written)
	}
	if err := run(true); err != nil {
		t.Fatalf("check logo após sincronizar deveria passar: %v", err)
	}
}

func TestCheckDetectsDrift(t *testing.T) {
	dir := prepare(t)
	source := filepath.Join(dir, "source.json")
	if err := os.WriteFile(source, []byte(compactSpec), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(snapshotPath, []byte("{}\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv(sourceVariable, source)
	if err := run(true); !errors.Is(err, errDrift) {
		t.Fatalf("esperado errDrift, veio %v", err)
	}
}

func TestSyncFetchesFromURL(t *testing.T) {
	prepare(t)
	api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(compactSpec))
	}))
	defer api.Close()
	t.Setenv(sourceVariable, api.URL)
	if err := run(false); err != nil {
		t.Fatal(err)
	}
}

func TestSyncRejectsBadSources(t *testing.T) {
	dir := prepare(t)
	failing := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
	}))
	defer failing.Close()
	invalid := filepath.Join(dir, "invalid.json")
	if err := os.WriteFile(invalid, []byte("nao-json"), 0o600); err != nil {
		t.Fatal(err)
	}
	for name, source := range map[string]string{
		"status de erro":      failing.URL,
		"arquivo inexistente": filepath.Join(dir, "nada.json"),
		"json inválido":       invalid,
	} {
		t.Run(name, func(t *testing.T) {
			t.Setenv(sourceVariable, source)
			if err := run(false); err == nil {
				t.Fatal("esperado erro")
			}
		})
	}
}
