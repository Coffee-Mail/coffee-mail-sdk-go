package coffeemail

import (
	"context"
	"errors"
	"net/http"
	"testing"
)

func TestListAllWalksEveryPage(t *testing.T) {
	client, transport := newTestClient(t, []fakeResponse{
		respond(http.StatusOK, `{"emails":[{"id":"e1"},{"id":"e2"}],"nextCursor":"c2"}`),
		respond(http.StatusOK, `{"emails":[{"id":"e3"}],"nextCursor":null}`),
	})
	var ids []string
	for email, err := range client.Emails.ListAll(context.Background(), &ListEmailsParams{Status: EmailStatusSent, Limit: 2}) {
		if err != nil {
			t.Fatal(err)
		}
		ids = append(ids, email.ID)
	}
	requests := transport.all()
	if len(ids) != 3 || ids[2] != "e3" || len(requests) != 2 {
		t.Fatalf("ids = %v, páginas = %d", ids, len(requests))
	}
	if requests[0].query != "limit=2&status=sent" || requests[1].query != "after=c2&limit=2&status=sent" {
		t.Fatalf("queries = %q / %q", requests[0].query, requests[1].query)
	}
}

func TestListAllStopsWhenTheLoopBreaks(t *testing.T) {
	client, transport := newTestClient(t, []fakeResponse{respond(http.StatusOK, `{"suppressions":[{"id":"s1"},{"id":"s2"}],"nextCursor":"c2"}`)})
	for range client.Suppressions.ListAll(context.Background(), nil) {
		break
	}
	if n := len(transport.all()); n != 1 {
		t.Fatalf("break deveria parar a paginação, houve %d requisições", n)
	}
}

func TestListAllYieldsTheErrorAndStops(t *testing.T) {
	client, transport := newTestClient(t, []fakeResponse{
		respond(http.StatusOK, `{"broadcasts":[{"id":"b1"}],"nextCursor":"c2"}`),
		respond(http.StatusInternalServerError, `{}`),
	})
	var seen []string
	var failure error
	for broadcast, err := range client.Broadcasts.ListAll(context.Background(), &ListBroadcastsParams{After: "c1"}) {
		if err != nil {
			failure = err
			continue
		}
		seen = append(seen, broadcast.ID)
	}
	if len(seen) != 1 || !errors.Is(failure, ErrServer) || len(transport.all()) != 2 {
		t.Fatalf("vistos = %v, erro = %v", seen, failure)
	}
	if transport.all()[0].query != "after=c1" {
		t.Fatalf("o cursor inicial deveria ser respeitado: %q", transport.all()[0].query)
	}
}

func TestListAllWithEmptyCursorStops(t *testing.T) {
	client, transport := newTestClient(t, []fakeResponse{respond(http.StatusOK, `{"emails":[],"nextCursor":""}`)})
	for range client.Emails.ListAll(context.Background(), nil) {
		t.Fatal("página vazia não deveria produzir itens")
	}
	if len(transport.all()) != 1 {
		t.Fatal("cursor vazio deveria encerrar")
	}
}
