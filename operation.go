package coffeemail

import (
	"net/http"
	"net/url"
)

type operation struct {
	method             string
	path               string
	query              url.Values
	body               any
	idempotencyKey     string
	sandbox            bool
	acceptsIdempotency bool
	fullAccessName     string
}

func get(path string) operation {
	return operation{method: http.MethodGet, path: path}
}

func post(path string, body any) operation {
	return operation{method: http.MethodPost, path: path, body: body}
}

func put(path string, body any) operation {
	return operation{method: http.MethodPut, path: path, body: body}
}

func patch(path string, body any) operation {
	return operation{method: http.MethodPatch, path: path, body: body}
}

func del(path string) operation {
	return operation{method: http.MethodDelete, path: path}
}

func (o operation) withQuery(query url.Values) operation {
	o.query = query
	return o
}

func (o operation) withSendHeaders(idempotencyKey string, sandbox bool) operation {
	o.idempotencyKey = idempotencyKey
	o.sandbox = sandbox
	o.acceptsIdempotency = true
	return o
}

func (o operation) requiringFullAccess(name string) operation {
	o.fullAccessName = name
	return o
}

func (o operation) url() string {
	encoded := o.query.Encode()
	if encoded == "" {
		return o.path
	}
	return o.path + "?" + encoded
}
