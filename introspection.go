package coffeemail

import (
	"context"
	"sync"
	"time"

	"github.com/Coffee-Mail/coffee-mail-sdk-go/internal/routes"
)

const introspectionTTL = 5 * time.Minute

// APIKeyIntrospection descreve a chave de API em uso.
type APIKeyIntrospection struct {
	ID             string   `json:"id"`
	Name           string   `json:"name"`
	OrganisationID string   `json:"organisationId"`
	Scopes         []string `json:"scopes"`
}

type introspectionCache struct {
	r       *requester
	now     func() time.Time
	mu      sync.Mutex
	value   *APIKeyIntrospection
	expires time.Time
}

func newIntrospectionCache(r *requester, now func() time.Time) *introspectionCache {
	return &introspectionCache{r: r, now: now}
}

// Introspect devolve os dados da chave de API em uso. O resultado fica em cache
// por 5 minutos; use [Client.InvalidateAPIKeyCache] para forçar nova consulta.
func (c *Client) Introspect(ctx context.Context) (*APIKeyIntrospection, error) {
	return c.introspection.get(ctx)
}

// InvalidateAPIKeyCache descarta o resultado em cache de [Client.Introspect].
func (c *Client) InvalidateAPIKeyCache() {
	c.introspection.invalidate()
}

func (ic *introspectionCache) get(ctx context.Context) (*APIKeyIntrospection, error) {
	if cached := ic.cached(); cached != nil {
		return cached, nil
	}
	fresh, err := call[APIKeyIntrospection](ctx, ic.r, get(routes.APIKeyIntrospection))
	if err != nil {
		return nil, err
	}
	ic.store(fresh)
	return fresh, nil
}

func (ic *introspectionCache) cached() *APIKeyIntrospection {
	ic.mu.Lock()
	defer ic.mu.Unlock()
	if ic.value == nil || !ic.now().Before(ic.expires) {
		return nil
	}
	return ic.value
}

func (ic *introspectionCache) store(value *APIKeyIntrospection) {
	ic.mu.Lock()
	defer ic.mu.Unlock()
	ic.value = value
	ic.expires = ic.now().Add(introspectionTTL)
}

func (ic *introspectionCache) invalidate() {
	ic.mu.Lock()
	defer ic.mu.Unlock()
	ic.value = nil
}
