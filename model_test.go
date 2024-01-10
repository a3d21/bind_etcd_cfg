package bind_etcd_cfg

import (
	"reflect"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
)

type AStruct struct {
	Name string `json:"name" yaml:"name"`
	City string `json:"city" yaml:"city"`
}

func newHolder[T any](typ T) *Holder {
	h := &Holder{
		typ: reflect.TypeOf(typ),
		v:   &atomic.Value{},
	}
	// init empty
	var err error
	k := reflect.TypeOf(typ).Kind()
	if k == reflect.Slice {
		err = h.Refresh("[1,2,3]")
	} else if k == reflect.String {
		err = h.Refresh("foobar")
	} else if k == reflect.Int {
		err = h.Refresh("1024")
	} else {
		err = h.Refresh(`{"name":"foo","city":"SZ"}`)
	}
	if err != nil {
		panic(err)
	}

	return h
}

// TestHolderTypes 测试BindCfg支持的typ类型，覆盖一般配置结构类型
func TestHolderTypes(t *testing.T) {
	assert.Equal(t, newHolder(AStruct{}).Get(), AStruct{Name: "foo", City: "SZ"})
	assert.Equal(t, newHolder(&AStruct{}).Get(), &AStruct{Name: "foo", City: "SZ"})
	assert.Equal(t, newHolder(map[string]string{}).Get(), map[string]string{"name": "foo", "city": "SZ"})
	assert.Equal(t, newHolder([]int{}).Get(), []int{1, 2, 3})
	assert.Equal(t, newHolder("").Get(), "foobar")
	assert.Equal(t, newHolder(1).Get(), 1024)
}
