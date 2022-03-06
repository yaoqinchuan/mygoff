# ORM结果处理

## 1、数据结构

查询结果的数据结构如下：

```go
type Value  = *gvar.Var              // 返回数据表记录值
type Record   map[string]Value       // 返回数据表记录键值对
type Result   []Record               // 返回数据表记录列表
```

1. Value/Record/Result用于ORM操作的结果数据类型。
2. Result表示数据表记录列表，Record表示一条数据表记录，Value表示记录中的一条键值数据。
3. Value是*gvar.Var类型的别名类型，方便于后续的数据类型转换。

## 2、Record数据记录

​		gdb为数据表记录操作提供了极高的灵活性和简便性，除了支持以map的形式访问/操作数据表记录以外，也支持将数据表记录转换为struct进行处理。我们以下使用一个简单的示例来演示该特性。

首先，我们的用户表结构是这样的（简单设计的示例表）：

```sql
CREATE TABLE `user` (
  `uid` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `name` varchar(30) NOT NULL DEFAULT '' COMMENT '昵称',
  `site` varchar(255) NOT NULL DEFAULT '' COMMENT '主页',
  PRIMARY KEY (`uid`)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8;
```

其次，我们的表数据如下：

```
uid  name   site
1    john   https://goframe.org
```

```go
package main

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gctx"
)

type User struct {
	Uid  int
	Name string
}

func main() {
	var (
		user *User
		ctx  = gctx.New()
	)
	err := g.DB().Model("user").Where("uid", 1).Scan(&user)
	if err != nil {
		g.Log().Header(false).Fatal(ctx, err)
	}
	if user != nil {
		g.Log().Header(false).Print(ctx, user)
	}
}
```

​		这里，我们自定义了一个struct，里面只包含了Uid和Name属性，可以看到它的属性并不和数据表的字段一致，这也是ORM灵活的特性之一：支持指定属性获取。

​		通过gdb.Model.Scan方法可以将查询到的数据记录转换为struct对象或者struct对象数组。由于这里传递的参数为&user即**User类型，那么将会转换为一个struct对象，如果传递为[]*User类型的参数，将会转换为数组结果。

**属性字段映射规则：**

​		需要注意的是，map中的键名为uid,name,site，而struct中的属性为Uid,Name，那么他们之间是如何执行映射的呢？主要是以下几点简单的规则：

- struct中需要匹配的属性必须为公开属性(首字母大写)；
- 记录结果中键名会自动按照  不区分大小写  且  忽略-/_/空格符号  的形式与struct属性进行匹配；
- 如果匹配成功，那么将键值赋值给属性，如果无法匹配，那么忽略该键值；
  

## 3、Result数据集合

​		Result/Record数据类型根据数据结果集操作的需要，往往需要根据记录中特定的字段作为键名进行数据检索，因此它包含多个用于转换Map/List的方法，同时也包含了常用数据结构JSON/XML的转换方法。

​		由于方法比较简单，这里便不再举例说明。需要注意的是两个方法Record.Map及Result.List，这两个方法也是使用比较频繁的方法，用以将ORM查询结果信息转换为可做展示的数据类型。由于结果集字段值底层为[]byte类型，虽然使用了新的Value类型做了封装，并且也提供了数十种常见的类型转换方法，但是大多数时候需要直接将结果Result或者Record直接作为json或者xml数据结构返回，就需要做转换才行。

```go
package main

import (
	"database/sql"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gctx"

)

type User struct {
	Uid  int
	Name string
	Site string
}

func main() {
	var (
		user []*User
		ctx  = gctx.New()
	)
	err := g.DB().Model("user").Where("uid", 1).Scan(&user)
	if err != nil && err != sql.ErrNoRows {
		g.Log().Header(false).Fatal(ctx, err)
	}
	if user != nil {
		g.Log().Header(false).Print(ctx, user)
	}
}
```

## 4、结果为空判断

​		使用gf ORM对返回结果为空判断非常简便，大部分场景下直接判断返回的数据是否为nil或者长度为0，或者使用IsEmpty/IsNil方法。

### 4.1、数据集合

```go
r, err := g.Model("order").Where("status", 1).All()
if err != nil {
	return err
}
if len(r) == 0 {
    // 结果为空
}
```

```go
r, err := g.Model("order").Where("status", 1).All()
if err != nil {
	return err
}
if r.IsEmpty() {
    // 结果为空
}
```

### 4.2、数据记录

```go
r, err := g.Model("order").Where("status", 1).One()
if err != nil {
    return err
}
if len(r) == 0 {
    // 结果为空
}
```

```go
r, err := g.Table("order").Where("status", 1).One()
if err != nil {
    return err
}
if r.IsEmpty() {
    // 结果为空
}


```

### 4.3、数据字段值

返回的是一个"泛型"变量，这个只能使用IsEmpty来判断是否为空了。

```go
r, err := g.Model("order").Where("status", 1).Value()
if err != nil {
	return err
}
if r.IsEmpty() {
    // 结果为空
}
```

### 4.4、struct对象

​		当传递的对象本身就是一个空指针时，如果查询到数据，那么会在内部自动创建这个对象；如果没有查询到数据，那么这个空指针仍旧是一个空指针，内部并不会做任何处理。

```go
var user *User
err := g.Model("order").Where("status", 1).Scan(&user)
if err != nil {
    return err
}
if user == nil {
    // 结果为空
}
```

​		当传递的对象本身已经是一个初始化的对象，如果查询到数据，那么会在内部将数据赋值给这个对象；如果没有查询到数据，那么此时就没办法将对象做nil判断空结果。因此ORM会返回一个sql.ErrNoRows错误，提醒开发者没有查询到任何数据并且对象没有做任何赋值，对象的所有属性还是给定的初始化数值，以便开发者可以做进一步的空结果判断。

```go
var user = new(User)
err := g.Model("order").Where("status", 1).Scan(&user)
if err != nil && err != sql.ErrNoRows {
    return err
}
if err == sql.ErrNoRows {
    // 结果为空
}
```

​		所以推荐不要传递一个初始化过后的对象给ORM，而是直接传递一个对象的指针的指针类型（**struct类型），ORM内部会根据查询结果智能地做自动初始化。

### 4.5、struct数组

​		当传递的对象数组本身是一个空数组（长度为0），如果查询到数据，那么会在内部自动赋值给数组；如果没有查询到数据，那么这个空数组仍旧是一个空数组，内部并不会做任何处理。

```go
var users []*User
err := g.Model("order").Where("status", 1).Scan(&users)
if err != nil {
    return err
}
if len(users) == 0 {
    // 结果为空
}
```

​		当传递的对象数组本身不是空数组，如果查询到数据，那么会在内部自动从索引0位置覆盖到数组上；如果没有查询到数据，那么此时就没办法将数组做长度为0判断空结果。因此ORM会返回一个sql.ErrNoRows错误，提醒开发者没有查询到任何数据并且数组没有做任何赋值，以便开发者可以做进一步的空结果判断。

```go
var users = make([]*User, 100)
err := g.Model("order").Where("status", 1).Scan(&users)
if err != nil {
    return err
} 
if err == sql.ErrNoRows {     
    // 结果为空
}
```

​		由于struct转换利用了Golang反射特性，执行性能会有一定的损耗。如果您涉及到大量查询结果数据的struct数组对象转换，并且需要提高转换性能，请参考自定义实现对应struct的UnmarshalValue方法

## 5、ORM接口开发

​		gdb模块使用了非常灵活且扩展性强的接口设计，接口设计允许开发者可以非常方便地自定义实现和替换接口定义中的任何方法。

### 5.1、DB接口

​		DB接口是数据库操作的核心接口，也是我们通过ORM操作数据库时最常用的接口，这里主要对接口的几个重要方法做说明：

- Open方法用于创建特定的数据库连接对象，返回的是标准库的*sql.DB通用数据库对象。*

- *Do*系列方法的第一个参数link为Link接口对象，该对象在master-slave模式下可能是一个主节点对象，也可能是从节点对象，因此如果在继承的驱动对象实现中使用该link参数时，注意当前的运行模式。slave节点在大部分的数据库主从模式中往往是不可写的。

- HandleSqlBeforeCommit方法将会在每一条SQL提交给数据库服务端执行时被调用做一些提交前的回调处理。

- 其他接口方法详见

  [接口文档]: https://pkg.go.dev/github.com/gogf/gf/v2/database/gdb#DB

### 5.2、Driver接口

开发者自定义的驱动需要实现以下接口：

```go
// Driver is the interface for integrating sql drivers into package gdb.
type Driver interface {
	// New creates and returns a database object for specified database server.
	New(core *Core, node *ConfigNode) (DB, error)
}
```

​		其中的New方法用于根据Core数据库基础对象以及ConfigNode配置对象创建驱动对应的数据库操作对象，需要注意的是，返回的数据库对象需要实现DB接口。而数据库基础对象Core已经实现了DB接口，因此开发者只需要”继承”Core对象，然后根据需要覆盖对应的接口实现方法即可。

## 6、上下文变量

​		ORM支持传递自定义的context上下文变量，用于异步IO控制、上下文信息传递（特别是链路跟踪信息的传递）、以及嵌套事务支持。

​		我们可以通过Ctx方法传递自定义的上下文变量给ORM对象，Ctx方法其实是一个链式操作方法，该上下文传递进去后仅对当前DB接口对象有效，方法定义如下：

```go
func Ctx(ctx context.Context) DB
```

### 6.1、请求超时控制

```go
ctx, cancel := context.WithTimeout(context.Background(), time.Second)
defer cancel()
_, err := db.Ctx(ctx).Query("SELECT SLEEP(10)")
fmt.Println(err)
```

该示例中执行会sleep 10秒中，因此必定会引发请求的超时。执行后，输出结果为：

```
context deadline exceeded, SELECT SLEEP(10)
```

