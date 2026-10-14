package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

const (
	sourceVariable   = "COFFEEMAIL_OPENAPI_SPEC"
	defaultSource    = "https://api.coffeemail.com.br/v1/product/docs/json"
	snapshotPath     = "spec/product.json"
	httpPrefix       = "http://"
	httpsPrefix      = "https://"
	fetchTimeout     = 30 * time.Second
	maxSpecBytes     = 20 << 20
	snapshotFileMode = 0o644
	jsonIndent       = "  "
	exitFailure      = 1
)

var errDrift = errors.New("spec/product.json diverge da fonte; rode go run ./internal/tools/syncspec")

func main() {
	check := flag.Bool("check", false, "falha se o snapshot divergir da fonte, sem escrever")
	flag.Parse()
	if err := run(*check); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(exitFailure)
	}
}

func run(check bool) error {
	source := os.Getenv(sourceVariable)
	if source == "" {
		source = defaultSource
	}
	raw, err := read(source)
	if err != nil {
		return err
	}
	normalized, err := normalize(raw)
	if err != nil {
		return fmt.Errorf("%s não é JSON válido: %w", source, err)
	}
	if check {
		return compare(normalized)
	}
	return os.WriteFile(snapshotPath, normalized, snapshotFileMode)
}

func read(source string) ([]byte, error) {
	if strings.HasPrefix(source, httpPrefix) || strings.HasPrefix(source, httpsPrefix) {
		return fetch(source)
	}
	return os.ReadFile(source)
}

func fetch(source string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), fetchTimeout)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, source, http.NoBody)
	if err != nil {
		return nil, err
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = res.Body.Close() }()
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%s respondeu %d", source, res.StatusCode)
	}
	return io.ReadAll(io.LimitReader(res.Body, maxSpecBytes))
}

func normalize(raw []byte) ([]byte, error) {
	var out bytes.Buffer
	if err := json.Indent(&out, bytes.TrimSpace(raw), "", jsonIndent); err != nil {
		return nil, err
	}
	out.WriteByte('\n')
	return out.Bytes(), nil
}

func compare(normalized []byte) error {
	current, err := os.ReadFile(snapshotPath)
	if err != nil {
		return err
	}
	if !bytes.Equal(current, normalized) {
		return errDrift
	}
	return nil
}
