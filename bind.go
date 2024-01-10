package bind_etcd_cfg

import (
	"context"
	"fmt"
	clientv3 "go.etcd.io/etcd/client/v3"
	"reflect"
	"sync/atomic"
)

// Supplier 提供最新配置
type Supplier[T any] func() T

// Listener 监听配置变更, optional
type Listener[T any] func(T)

func (sup Supplier[T]) Get() T { return sup() }

// MustLoad panic when fail
func MustLoad[T any](v3cli *clientv3.Client, key string, typ T) T {
	res, err := Load(v3cli, key, typ)
	if err != nil {
		panic(fmt.Errorf("load cfg fail, err: %v", err))
	}
	return res
}

// Load nacos config typed
func Load[T any](v3cli *clientv3.Client, key string, typ T) (T, error) {
	var empty T

	h := &Holder{
		typ: reflect.TypeOf(typ),
		v:   &atomic.Value{},
	}

	ctx := context.Background()
	get, err := v3cli.Get(ctx, key)
	if err != nil {
		return empty, err
	}

	raw := ""
	if len(get.Kvs) > 0 {
		raw = string(get.Kvs[0].Value)
	}
	err = h.Refresh(raw)
	if err != nil {
		return empty, err
	}

	return h.Get().(T), nil
}

// MustBind panic when fail
func MustBind[T any](v3cli *clientv3.Client, key string, typ T, lis ...Listener[T]) Supplier[T] {
	sup, err := Bind(v3cli, key, typ, lis...)
	if err != nil {
		panic(fmt.Errorf("bind cfg fail, err: %v", err))
	}
	return sup
}

// Bind dynamic bind config with typ, return `Supplier[T]` getting the latest config
// lis  optional, listen config change
func Bind[T any](v3cli *clientv3.Client, key string, typ T, lis ...Listener[T]) (Supplier[T], error) {
	h := &Holder{
		typ: reflect.TypeOf(typ),
		v:   &atomic.Value{},
	}

	ctx := context.Background()
	get, err := v3cli.Get(ctx, key)
	if err != nil {
		return nil, err
	}

	raw := ""
	if len(get.Kvs) > 0 {
		raw = string(get.Kvs[0].Value)
	}

	err = h.Refresh(raw)
	if err != nil {
		return nil, err
	}

	// lis
	for _, li := range lis {
		li(h.Get().(T))
	}

	// watch
	ch := v3cli.Watch(ctx, key)
	go func() {
		for w := range ch {
			for _, evt := range w.Events {
				if evt.Type == clientv3.EventTypePut {
					if data := string(evt.Kv.Value); data != "" {
						err2 := h.Refresh(data)
						if err2 != nil {
							defaultLogger.Errorf("refresh fail, raw: %v err: %v", data, err2)
							continue
						}

						for _, li := range lis {
							li(h.Get().(T))
						}
					}

				}
			}
		}
	}()

	if err != nil {
		return nil, err
	}

	return func() T { return h.Get().(T) }, nil
}
