package ctxutil

import "context"

// key is a unique type that we can use as a key in a context
type key[T any] struct{}

// WithValue returns a new context with a given value set. Only one value of each type can be set in a context,
// setting a value of the same type will overwrite the previous value
func WithValue[T any](ctx context.Context, value T) context.Context {
	return context.WithValue(ctx, key[T]{}, value)
}

// Value returns the value of type T in the given context, or false if the context does not contain a value of type T.
func Value[T any](ctx context.Context) (T, bool) {
	value, ok := ctx.Value(key[T]{}).(T)
	return value, ok
}
