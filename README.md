# bind_etcd_cfg

bind_etcd_cfg 是一个动态绑定配置工具包，实现etcd配置绑定。

支持配置数据

- json
- yaml

支持配置类型

- map[K]V
- []T
- struct
- *struct
- string
- int
- ...

```go
package main

import (
	"fmt"
	clientv3 "go.etcd.io/etcd/client/v3"
	"github.com/a3d21/bind_etcd_cfg"
	"github.com/sirupsen/logrus"
)

type ConfigStruct struct {
	Name string `json:"name" yaml:"name"`
	City string `json:"city" yaml:"city"`
}

func main() {
	var v3cli *clientv3.Client // TODO 初始化
	key := "/akey"
	bind_etcd_cfg.SetLogger(logrus.New()) // use logrus logger
	var GetConf = bind_etcd_cfg.MustBind(v3cli, key, &ConfigStruct{})
	// 也支持原生结构类型
	// var GetConf = bind_etcd_cfg.MustBind(cli, key, ConfigStruct{})
	// var GetConf = bind_etcd_cfg.MustBind(cli, key, map[string]string{})
	// var GetConf = bind_etcd_cfg.MustBind(cli, key, []string{})
	// 绑定并监听变更。考虑到存在需要监听配置做额外操作的场景，增加的可选Listner[T]参数
	// var GetConf = bind_etcd_cfg.MustBind(cli, key, &ConfigStruct{}, func(v *ConfigStruct) {
	// 	 fmt.Println(v)
	// })

	// c.(type) == *ConfigStruct
	c := GetConf() // 获取最新配置
	c = GetConf()
	c = GetConf() // 配置变更自动监听更新

	fmt.Println(c)
}

```