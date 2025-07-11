package registry

type Registry[T any] struct {
	items map[string]T
}

func NewRegistry[T any]() *Registry[T] {
	return &Registry[T]{items: make(map[string]T)}
}

func (r *Registry[T]) Register(name string, value T) {
	r.items[name] = value
}

func (r *Registry[T]) Get(name string) (T, bool) {
	item, ok := r.items[name]
	return item, ok
}
