package routes

import "testing"

func TestExpandEscapesIdentifiers(t *testing.T) {
	if got := Expand(EmailEvents, "a/b"); got != "/v1/product/emails/a%2Fb/events" {
		t.Fatalf("Expand = %q", got)
	}
	if got := ExpandContact(AudienceContact, "a1", "c 1"); got != "/v1/product/audiences/a1/contacts/c%201" {
		t.Fatalf("ExpandContact = %q", got)
	}
}
