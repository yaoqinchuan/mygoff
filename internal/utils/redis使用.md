# Redis使用

## 1、常见的redis使用

### 1.1、配置文件（推荐）

​		绝大部分情况下推荐使用g.Redis单例方式来操作redis。因此同样推荐使用配置文件来管理Redis配置，在config.yaml中的配置示例如下：

**单实例配置**

```yml
# Redis 配置示例
redis:
  # 单实例配置示例1
  default:
    address: 127.0.0.1:6379
    db:      1

  # 单实例配置示例2
  cache:
    address:     127.0.0.1:6379
    db:          1
    pass:        123456
    idleTimeout: 600
```

**集群化配置**

```yml
# Redis 配置示例
redis:
   # 集群模式配置方法
  group:	
    address: 127.0.0.1:6379,127.0.0.1:6370
    db:      1
```

**配置项说明**

| 配置项名称        | 是否必须 | 默认值  | 说明                                                         |
| :---------------- | :------- | :------ | :----------------------------------------------------------- |
| `address`         | 是       | -       | 格式：`地址:端口`支持`Redis`单实例模式和集群模式配置，使用`,`分割多个地址。例如：`192.168.1.1:6379, 192.168.1.2:6379` |
| `db`              | 否       | `0`     | 数据库索引                                                   |
| `pass`            | 否       | `-`     | 访问授权密码                                                 |
| `minIdle`         | 否       | `0`     | 允许闲置的最小连接数                                         |
| `maxIdle`         | 否       | `10`    | 允许闲置的最大连接数(`0`表示不限制)                          |
| `maxActive`       | 否       | `100`   | 最大连接数量限制(`0`表示不限制)                              |
| `idleTimeout`     | 否       | `10`    | 连接最大空闲时间，使用时间字符串例如`30s/1m/1d`              |
| `maxConnLifetime` | 否       | `30`    | 连接最长存活时间，使用时间字符串例如`30s/1m/1d`              |
| `waitTimeout`     | 否       | `0`     | 等待连接池连接的超时时间，使用时间字符串例如`30s/1m/1d`      |
| `dialTimeout`     | 否       | `0`     | `TCP`连接的超时时间，使用时间字符串例如`30s/1m/1d`           |
| `readTimeout`     | 否       | `0`     | `TCP`的`Read`操作超时时间，使用时间字符串例如`30s/1m/1d`     |
| `writeTimeout`    | 否       | `0`     | `TCP`的`Write`操作超时时间，使用时间字符串例如`30s/1m/1d`    |
| `masterName`      | 否       | `-`     | 哨兵模式下使用, 设置`MasterName`                             |
| `tls`             | 否       | `false` | 是否使用`TLS`认证                                            |
| `tlsSkipVerify`   | 否       | `false` | 通过`TLS`连接时，是否禁用服务器名称验证                      |

**使用用例**

```go
func main() {
	var (
		ctx = context.Background()
	)
	conn, _ := g.Redis().Conn(ctx)
	conn.Do(ctx, "SET", "Key", "Value")
	v, _ := conn.Do(ctx, "GET", "Key")
	fmt.Println(v.String())
}
```

​		其中的 default 和 cache 分别表示配置分组名称，我们在程序中可以通过该名称获取对应配置的 redis 单例对象。不传递分组名称时，默认使用 redis.default 配置分组项)来获取对应配置的 redis 客户端单例对象

### 1.2、配置方法（高级）

​		由于GoFrame是模块化的框架，除了可以通过耦合且便捷的g模块来自动解析配置文件并获得单例对象之外，也支持有能力的开发者模块化使用gredis包。

​		gredis提供了全局的分组配置功能，相关配置管理方法如下：

```go
func SetConfig(config Config, name ...string)
func SetConfigByMap(m map[string]interface{}, name ...string) error
func GetConfig(name ...string) (config Config, ok bool)
func RemoveConfig(name ...string)
func ClearConfig()
```

​		其中name参数为分组名称，即为通过分组来对配置对象进行管理，我们可以为不同的配置对象设置不同的分组名称，随后我们可以通过Instance单例方法获取redis客户端操作对象单例。

```go
func Instance(name ...string) *Redis
```

使用示例：

```go
package main

import (
	"context"
	"fmt"
	"github.com/gogf/gf/v2/database/gredis"
	"github.com/gogf/gf/v2/util/gconv"
)

var (
	config = gredis.Config{
		Address: "192.168.1.2:6379, 192.168.1.3:6379",
		Db   : 1,
	}
	ctx = context.Background()
)

func main() {
	group := "test"
	gredis.SetConfig(&config, group)

	redis := gredis.Instance(group)
	defer redis.Close(ctx)
	
	_, err := redis.Do(ctx, "SET", "k", "v")
	if err != nil {
		panic(err)
	}
	
	r, err := redis.Do(ctx, "GET", "k")
	if err != nil {
		panic(err)
	}
	fmt.Println(gconv.String(r))

}
```

## 2、基本使用

### 2.1、Do方法

​		我们最常用的是Do方法，执行同步指令，通过向Redis Server发送对应的Redis API命令，来使用Redis Server的服务。<u>Do方法最大的特点是内部自行从连接池中获取连接对象，使用完毕后自动丢回连接池中，开发者无需手动调用Close方法</u>，方便使用。

```go
package main

import (
	"fmt"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gctx"
)

func main() {
	var (
		ctx = gctx.New()
	)
	v, _ := g.Redis().Do(ctx, "SET", "k", "v")
	fmt.Println(v.String())
}
```

### 2.2、HSET/HGETALL操作

```go
package main

import (
	"fmt"
	"github.com/gogf/gf/v2/container/gvar"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gctx"
)

func main() {
	var (
		err    error
		result *gvar.Var
		key    = "user"
		ctx = gctx.New()
	)
	_, err = g.Redis().Do(ctx, "HSET", key, "id", 10000)
	if err != nil {
		panic(err)
	}
	_, err = g.Redis().Do(ctx,"HSET", key, "name", "john")
	if err != nil {
		panic(err)
	}
	result, err = g.Redis().Do(ctx,"HGETALL", key)
	if err != nil {
		panic(err)
	}
	fmt.Println(result.Map())

	// May Output:
	// map[id:10000 name:john]

}
```

### 2.3、HSET/HGETALL操作

```go
package main

import (
	"fmt"
	"github.com/gogf/gf/v2/container/gvar"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gctx"
)

func main() {
	var (
		err    error
		result *gvar.Var
		key    = "user"
		ctx = gctx.New()
	)
	_, err = g.Redis().Do(ctx, "HSET", key, "id", 10000)
	if err != nil {
		panic(err)
	}
	_, err = g.Redis().Do(ctx,"HSET", key, "name", "john")
	if err != nil {
		panic(err)
	}
	result, err = g.Redis().Do(ctx,"HGETALL", key)
	if err != nil {
		panic(err)
	}
	fmt.Println(result.Map())

	// May Output:
	// map[id:10000 name:john]

}
```

### 2.4、HMSET/HMGET操作

我们可以通过map参数执行HMSET操作。

```go
package main

import (
	"fmt"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gctx"
	"github.com/gogf/gf/v2/util/gutil"
)

func main() {
	var (
		ctx = gctx.New()
		key  = "user_100"
		data = g.Map{
			"name":  "gf",
			"sex":   0,
			"score": 100,
		}
	)
	_, err := g.Redis().Do(ctx, "HMSET", append(g.Slice{key}, gutil.MapToSlice(data)...)...)
	if err != nil {
		g.Log().Fatal(ctx, err)
	}
	v, err := g.Redis().Do(ctx, "HMGET", key, "name")
	if err != nil {
		g.Log().Fatal(ctx, err)
	}
	fmt.Println(v.Slice())

	// May Output:
	// [gf]

}
```

我们可以通过 struct 参数执行HMSET操作。

```go
package main

import (
	"fmt"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gctx"
	"github.com/gogf/gf/v2/util/gutil"
)

func main() {
	type User struct {
		Name  string `json:"name"`
		Sex   int    `json:"sex"`
		Score int    `json:"score"`
	}
	var (
		ctx = gctx.New()
		key  = "user_100"
		data = &User{
			Name:  "gf",
			Sex:   0,
			Score: 100,
		}
	)
	_, err := g.Redis().Do(ctx,"HMSET", append(g.Slice{key}, gutil.StructToSlice(data)...)...)
	if err != nil {
		g.Log().Fatal(ctx, err)
	}
	v, err := g.Redis().Do(ctx,"HMGET", key, "name")
	if err != nil {
		g.Log().Fatal(ctx, err)
	}
	fmt.Println(v.Slice())

	// May Output:
	// [gf]

}
```

### **2.5、自动序列化/反序列化**

​		当给定的参数为map, slice, struct时，gredis内部支持自动对其使用json序列化，并且读取数据时可使用gvar.Var的转换功能实现反序列化。

**map存取**

```go
package main

import (
	"fmt"
	"github.com/gogf/gf/v2/container/gvar"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gctx"
)

func main() {
	var (
		ctx = gctx.New()
		err    error
		result *gvar.Var
		key    = "user"
		data   = g.Map{
			"id":   10000,
			"name": "john",
		}
	)
	_, err = g.Redis().Do(ctx, "SET", key, data)
	if err != nil {
		panic(err)
	}
	result, err = g.Redis().Do(ctx,"GET", key)
	if err != nil {
		panic(err)
	}
	fmt.Println(result.Map())
}
```

**struct存取**

```go
package main

import (
	"fmt"
	"github.com/gogf/gf/v2/container/gvar"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gctx"
)

func main() {
	type User struct {
		Id   int
		Name string
	}

	var (
		ctx = gctx.New()
		err    error
		result *gvar.Var
		key    = "user"
		user   = g.Map{
			"id":   10000,
			"name": "john",
		}
	)
	
	_, err = g.Redis().Do(ctx, "SET", key, user)
	if err != nil {
		panic(err)
	}
	result, err = g.Redis().Do(ctx, "GET", key)
	if err != nil {
		panic(err)
	}
	
	var user2 *User
	if err = result.Struct(&user2); err != nil {
		panic(err)
	}
	fmt.Println(user2.Id, user2.Name)

}
```

## 3、Conn对象

​		使用Do方法已经能够满足绝大部分的场景需要，如果需要更复杂的Redis操作（例如订阅发布），那么我们可以使用Conn方法从连接池中获取一个连接对象，随后使用该连接对象进行操作。<u>但需要注意的是，该连接对象不再使用时，应当显式调用Close方法进行关闭（丢回连接池）</u>。

### 3.1、基本使用

```go
package main

import (
	"fmt"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gctx"
	"github.com/gogf/gf/v2/util/gconv"
)

func main() {
	var (
		ctx = gctx.New()
	)
	conn, _ := g.Redis().Conn(ctx)
	defer conn.Close(ctx)
	conn.Do(ctx, "SET", "k", "v")
	v, _ := conn.Do(ctx,"GET", "k")
	fmt.Println(gconv.String(v))
}
```

### 3.2、订阅/发布

​		我们可以通过Redis的SUBSCRIBE/PUBLISH命令实现订阅/发布模式。

```go
package main

import (
	"fmt"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gctx"
	"github.com/gogf/gf/v2/util/gconv"
)

func main() {
	var (
		ctx = gctx.New()
	)
	conn, _ := g.Redis().Conn(ctx)
	defer conn.Close(ctx)
	_, err := conn.Do(ctx, "SUBSCRIBE", "channel")
	if err != nil {
		panic(err)
	}
	for {
		reply, err := conn.Receive(ctx)
		if err != nil {
			panic(err)
		}
		fmt.Println(gconv.Strings(reply))
	}
}
```

执行后，程序将阻塞等待获取数据。

另外打开一个终端通过redis-cli命令进入Redis Server，发布一条消息：$ redis-cli

```go
127.0.0.1:6379> publish channel test
(integer) 1
127.0.0.1:6379
```

 随后程序终端立即打印出从`Redis Server`获取的数据： 

[message channel test]

## 4、接口化设计

```go
// SetAdapter sets custom adapter for current redis client.
func (r *Redis) SetAdapter(adapter Adapter) 

// GetAdapter returns the adapter that is set in current redis client.
func (r *Redis) GetAdapter() Adapter
```

接口设计文档https://pkg.go.dev/github.com/gogf/gf/v2/database/gredis#Adapter