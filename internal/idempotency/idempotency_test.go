package idempotency

import (
	"regexp"
	"testing"
)

var uuidV4 = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)

func TestNewKeyIsAUniqueUUIDv4(t *testing.T) {
	seen := map[string]bool{}
	for range 1000 {
		key := NewKey()
		if !uuidV4.MatchString(key) || seen[key] {
			t.Fatalf("chave inválida ou repetida: %q", key)
		}
		seen[key] = true
	}
}
