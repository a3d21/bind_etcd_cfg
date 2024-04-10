package bind_etcd_cfg

// Mock Supplier for dev
func Mock[T any](v T) Supplier[T] {
	return func() T {
		return v
	}
}

func MockPrefix[T any](m map[string]T) PrefixSupplier[T] {
	return &mockPrefixSupplier[T]{data: m}
}

type mockPrefixSupplier[T any] struct {
	data map[string]T
}

func (m *mockPrefixSupplier[T]) Get(key string) (T, bool) {
	t, ok := m.data[key]
	return t, ok
}

func (m *mockPrefixSupplier[T]) ToMap() map[string]T {
	return m.data
}
