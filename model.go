package bind_etcd_cfg

import (
	"encoding/json"
	"errors"
	"go.etcd.io/etcd/api/v3/mvccpb"
	"gopkg.in/yaml.v3"
	"reflect"
	"strings"
	"sync"

	"sync/atomic"
)

type EventType = mvccpb.Event_EventType

const (
	EventTypePut    EventType = mvccpb.PUT
	EventTypeDelete EventType = mvccpb.DELETE
)

type PrefixHolder[T any] struct {
	data sync.Map
}

func (h *PrefixHolder[T]) Refresh(eventType EventType, key, raw string) error {
	if eventType == EventTypeDelete {
		h.data.Delete(key)
		return nil
	}
	if eventType == EventTypePut {
		t := new(T)
		if err := unmarshal(raw, t); err != nil {
			return err
		}
		h.data.Store(key, *t)
		return nil
	}
	return nil
}

func (h *PrefixHolder[T]) Get(key string) (T, bool) {
	if v, ok := h.data.Load(key); ok {
		if vv, ok2 := v.(T); ok2 {
			return vv, true
		}
	}
	var empty T
	return empty, false
}

func (h *PrefixHolder[T]) Range(li PrefixListener[T]) {
	h.data.Range(func(key, value any) bool {
		k := key.(string)
		v := value.(T)
		li(EventTypePut, k, v)
		return true
	})
}

func (h *PrefixHolder[T]) ToMap() map[string]T {
	m := map[string]T{}
	h.data.Range(func(key, value any) bool {
		k := key.(string)
		v := value.(T)
		m[k] = v
		return true
	})
	return m
}

// Holder is a Dynamic Config Holder with `Refresh`
type Holder struct {
	typ reflect.Type
	v   *atomic.Value
}

func (h *Holder) Refresh(raw string) error {
	if raw == "" {
		return errors.New("empty raw")
	}

	ttyp := h.typ
	isPtr := ttyp.Kind() == reflect.Ptr
	if isPtr {
		ttyp = ttyp.Elem()
	}

	vv := reflect.New(ttyp)
	v := vv.Interface()
	err := unmarshal(raw, v)
	if err != nil {
		return err
	}

	if isPtr {
		h.v.Store(v)
	} else {
		h.v.Store(vv.Elem().Interface())
	}
	return nil
}

func (h *Holder) Get() interface{} {
	return h.v.Load()
}

// unmarshal json or yaml
func unmarshal(raw string, out interface{}) error {
	// is json?
	if strings.HasPrefix(raw, "{") || strings.HasPrefix(raw, "[") {
		return json.Unmarshal([]byte(raw), out)
	} else {
		return yaml.Unmarshal([]byte(raw), out)
	}
}
