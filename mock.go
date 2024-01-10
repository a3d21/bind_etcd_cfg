package bind_etcd_cfg

// Mock Supplier for dev
func Mock[T any](v T) Supplier[T] {
	return func() T {
		return v
	}
}
