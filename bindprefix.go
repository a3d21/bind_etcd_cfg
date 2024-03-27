package bind_etcd_cfg

import (
	"context"
	clientv3 "go.etcd.io/etcd/client/v3"
	"strings"
)

// PrefixSupplier 提供key查val
type PrefixSupplier[T any] func(key string) (T, bool)

// PrefixListener 值更新时触发
type PrefixListener[T any] func(eventType EventType, key string, val T)

func LoadPrefix[T any](v3cli *clientv3.Client, prefix string, _ T) (map[string]T, error) {
	h := &PrefixHolder[T]{}

	ctx := context.Background()
	get, err := v3cli.Get(ctx, prefix, clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}

	for _, kv := range get.Kvs {
		k := strings.TrimPrefix(string(kv.Key), prefix)
		v := string(kv.Value)
		if k != "" && v != "" {
			if err2 := h.Refresh(EventTypePut, k, v); err2 != nil {
				return nil, err2
			}
		}
	}
	return h.ToMap(), nil
}
func BindPrefix[T any](v3cli *clientv3.Client, prefix string, _ T, lis ...PrefixListener[T]) (PrefixSupplier[T], error) {
	h := &PrefixHolder[T]{}

	ctx := context.Background()
	get, err := v3cli.Get(ctx, prefix, clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}

	for _, kv := range get.Kvs {
		k := strings.TrimPrefix(string(kv.Key), prefix)
		v := string(kv.Value)
		if k != "" && v != "" {
			if err2 := h.Refresh(EventTypePut, k, v); err2 != nil {
				return nil, err2
			}
		}
	}

	// lis
	for _, li := range lis {
		h.Range(li)
	}

	// watch
	ch := v3cli.Watch(ctx, prefix, clientv3.WithPrefix())
	go func() {
		for w := range ch {
			for _, evt := range w.Events {
				if k := strings.TrimPrefix(string(evt.Kv.Key), prefix); k != "" {
					v := string(evt.Kv.Value)
					if err2 := h.Refresh(evt.Type, k, v); err2 != nil {
						defaultLogger.Errorf("refresh fail, type: %v key: %v val: %v err: %v", evt.Type, k, v, err2)
					} else {
						var t T
						if evt.Type == EventTypePut {
							t, _ = h.GetByKey(k)
						}
						for _, li := range lis {
							li(evt.Type, k, t)
						}
					}
				}
			}
		}
	}()

	return h.GetByKey, nil
}
