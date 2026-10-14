package coffeemail

import (
	"context"
	"iter"
)

type pageFetcher[T any] func(ctx context.Context, cursor string) ([]T, *string, error)

func paginate[T any](ctx context.Context, cursor string, fetch pageFetcher[T]) iter.Seq2[T, error] {
	return func(yield func(T, error) bool) {
		next := cursor
		for {
			items, following, err := fetch(ctx, next)
			if err != nil {
				var zero T
				yield(zero, err)
				return
			}
			if !yieldAll(items, yield) || following == nil || *following == "" {
				return
			}
			next = *following
		}
	}
}

func yieldAll[T any](items []T, yield func(T, error) bool) bool {
	for _, item := range items {
		if !yield(item, nil) {
			return false
		}
	}
	return true
}
