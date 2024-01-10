package bind_etcd_cfg

import (
	"encoding/json"
	"errors"
	"gopkg.in/yaml.v3"
	"reflect"
	"strings"

	"sync/atomic"
)

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
