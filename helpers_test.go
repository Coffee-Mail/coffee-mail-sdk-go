package coffeemail

import (
	"bytes"
	"io"
	"net/http"
	"sync"
	"testing"
)

const testAPIKey = "cm_test_0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"

type fakeResponse struct {
	status int
	body   string
	header map[string]string
	err    error
	block  bool
}

type recordedRequest struct {
	method string
	url    string
	path   string
	query  string
	header http.Header
	body   []byte
}

type fakeTransport struct {
	mu        sync.Mutex
	responses []fakeResponse
	requests  []recordedRequest
}

func (f *fakeTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	body, _ := io.ReadAll(req.Body)
	f.mu.Lock()
	f.requests = append(f.requests, recordedRequest{
		method: req.Method,
		url:    req.URL.String(),
		path:   req.URL.EscapedPath(),
		query:  req.URL.RawQuery,
		header: req.Header.Clone(),
		body:   body,
	})
	next := f.next()
	f.mu.Unlock()
	if next.block {
		<-req.Context().Done()
		return nil, req.Context().Err()
	}
	if next.err != nil {
		return nil, next.err
	}
	header := http.Header{}
	for name, value := range next.header {
		header.Set(name, value)
	}
	return &http.Response{
		StatusCode: next.status,
		Header:     header,
		Body:       io.NopCloser(bytes.NewBufferString(next.body)),
		Request:    req,
	}, nil
}

func (f *fakeTransport) next() fakeResponse {
	if len(f.responses) == 0 {
		return fakeResponse{status: http.StatusOK, body: "{}"}
	}
	current := f.responses[0]
	if len(f.responses) > 1 {
		f.responses = f.responses[1:]
	}
	return current
}

func (f *fakeTransport) all() []recordedRequest {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]recordedRequest(nil), f.requests...)
}

func (f *fakeTransport) last(t *testing.T) recordedRequest {
	t.Helper()
	requests := f.all()
	if len(requests) == 0 {
		t.Fatal("nenhuma requisição registrada")
	}
	return requests[len(requests)-1]
}

func newTestClient(t *testing.T, responses []fakeResponse, opts ...Option) (*Client, *fakeTransport) {
	t.Helper()
	transport := &fakeTransport{responses: responses}
	all := append([]Option{WithHTTPClient(&http.Client{Transport: transport})}, opts...)
	client, err := NewClient(testAPIKey, all...)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	return client, transport
}

func respond(status int, body string) fakeResponse {
	return fakeResponse{status: status, body: body}
}
