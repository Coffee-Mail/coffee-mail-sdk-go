package main

import (
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func runOn(t *testing.T, name, source string) []violation {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(source), 0o600); err != nil {
		t.Fatal(err)
	}
	c := &checker{fset: token.NewFileSet()}
	c.walk(dir)
	return c.violations
}

func messages(violations []violation) string {
	var parts []string
	for _, v := range violations {
		parts = append(parts, v.message)
	}
	return strings.Join(parts, "\n")
}

func TestDetectsEveryRule(t *testing.T) {
	cases := map[string]struct {
		file   string
		source string
		want   string
	}{
		"else":         {"a.go", "package a\nfunc f(x int) int {\n\tif x > 0 {\n\t\treturn 1\n\t} else {\n\t\treturn 2\n\t}\n}\n", "else proibido"},
		"switch":       {"a.go", "package a\nfunc f(x int) {\n\tswitch x {\n\t}\n}\n", "switch proibido"},
		"type switch":  {"a.go", "package a\nfunc f(x any) {\n\tswitch x.(type) {\n\t}\n}\n", "type switch proibido"},
		"comentário":   {"a.go", "package a\nfunc f() {\n\t// explicação\n}\n", "comentário proibido"},
		"doc privado":  {"a.go", "package a\n// f faz algo.\nfunc f() {}\n", "comentário proibido"},
		"println":      {"a.go", "package a\nimport \"fmt\"\nfunc f() { fmt.Println(1) }\n", "fmt.Println proibido"},
		"panic":        {"a.go", "package a\nfunc f() { panic(1) }\n", "panic proibido"},
		"profundidade": {"a.go", "package a\nfunc f(x int) {\n\tfor {\n\t\tif x > 0 {\n\t\t\tfor range x {\n\t\t\t\tif x > 1 {\n\t\t\t\t\treturn\n\t\t\t\t}\n\t\t\t}\n\t\t}\n\t}\n}\n", "profundidade maior que 3"},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := messages(runOn(t, tc.file, tc.source))
			if !strings.Contains(got, tc.want) {
				t.Fatalf("esperado %q, veio %q", tc.want, got)
			}
		})
	}
}

func TestAcceptsCompliantCode(t *testing.T) {
	source := "// Package a é um exemplo.\npackage a\n\n//go:generate echo ok\n\n// Exported é documentado.\ntype Exported struct {\n\t// Field é documentado.\n\tField int\n}\n\n// Códigos.\nconst (\n\tA = 1\n)\n\n// Do é documentado.\nfunc Do(x int) int {\n\tfn := func() int {\n\t\tfor {\n\t\t\tif x > 0 {\n\t\t\t\tfor range x {\n\t\t\t\t\treturn 1\n\t\t\t\t}\n\t\t\t}\n\t\t}\n\t}\n\treturn fn()\n}\n"
	if got := runOn(t, "a.go", source); len(got) != 0 {
		t.Fatalf("violações inesperadas: %s", messages(got))
	}
	if got := runOn(t, "a_test.go", "package a\nfunc f() { panic(1) }\n"); len(got) != 0 {
		t.Fatalf("testes podem usar panic/print: %s", messages(got))
	}
}
