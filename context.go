package coffeemail

import "context"

type requestIDKey struct{}

// WithRequestID devolve um context que envia id no header X-Request-Id, útil para
// correlacionar logs da sua aplicação com os da API.
func WithRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, requestIDKey{}, id)
}

func requestIDFrom(ctx context.Context) string {
	id, _ := ctx.Value(requestIDKey{}).(string)
	return id
}
