# 数据库ORM使用（一）

## 1、常见修改操作

### 1.1、数据结构

gdb数据库管理模块的内部配置管理数据结构如下：

```go
type Config      map[string]ConfigGroup // 数据库配置对象
type ConfigGroup []ConfigNode           // 数据库分组配置
// 数据库配置项(一个分组配置对应多个配置项)
type ConfigNode  struct {
    Host             string        // 地址
    Port             string        // 端口
    User             string        // 账号
    Pass             string        // 密码
    Name             string        // 数据库名称
    Type             string        // 数据库类型：mysql, sqlite, mssql, pgsql, oracle<br />    Link             string        // (可选)自定义链接信息，当该字段被设置值时，以上链接字段(Host,Port,User,Pass,Name)将失效(该字段是一个扩展功能)     Role             string        // (可选，默认为master)数据库的角色，用于主从操作分离，至少需要有一个master，参数值：master, slave
    Debug            bool          // (可选)开启调试模式
    Charset          string        // (可选，默认为 utf8)编码，默认为 utf8
    Prefix           string        // (可选)表名前缀
    Weight           int           // (可选)用于负载均衡的权重计算，当集群中只有一个节点时，权重没有任何意义
    MaxIdleConnCount int           // (可选)连接池最大闲置的连接数
    MaxOpenConnCount int           // (可选)连接池最大打开的连接数
    MaxConnLifetime  time.Duration // (可选，单位秒)连接对象可重复使用的时间长度
}
```

- ConfigNode用于存储一个数据库节点信息；
- ConfigGroup用于管理多个数据库节点组成的配置分组(一般一个分组对应一个业务数据库集群)；
- Config用于管理多个ConfigGroup配置分组。

配置管理特点：

- 支持多节点数据库集群管理；
- 每个节点可以单独配置连接属性；
- 采用单例模式管理数据库实例化对象；
- 支持对数据库集群分组管理，按照分组名称获取实例化的数据库操作对象；
- 支持多种关系型数据库管理，可通过ConfigNode.Type属性进行配置；
- 支持Master-Slave读写分离，可通过ConfigNode.Role属性进行配置；
- 支持客户端的负载均衡管理，可通过ConfigNode.Weight属性进行配置，值越大，优先级越高；
- 特别说明，gdb的配置管理最大的特点是，（同一进程中）所有的数据库集群信息都使用同一个配置管理模块进行统一维护，不同业务的数据库集群配置使用不同的分组名称进行配置和获取。

### 1.2、模型创建

​		Model方法用于创建基于数据表的Model对象。常见的，也可以使用g对象管理模块中的Model方法在默认的数据库配置上创建Model对象。

```go
g.Model("user")
// 或者
g.DB().Model("user")

m := g.DB("user-center").Model("user")
```

​		Raw方法用于创建一个基于原始SQL语句的Model对象。也可以使用g对象管理模块中的ModelRaw方法通过给定的SQL语句在默认的数据库配置上创建Model对象。

```go
s := "SELECT * FROM `user` WHERE `status` IN(?)"
m := g.ModelRaw(s, g.Slice{1,2,3}).WhereLT("age", 18).Limit(10).OrderAsc("id").All()
// SELECT * FROM `user` WHERE `status` IN(1,2,3) AND `age`<18 ORDER BY `id` ASC LIMIT 10
```

### 1.3、链式安全

​		链式安全只是模型操作的两种方式区别：一种会修改当前model对象（不安全，默认），一种不会（安全）但是模型属性修改/条件叠加需要使用赋值操作，仅此而已。

​		在默认情况下，gdb是非链式安全的，也就是说链式操作的每一个方法都将对当前操作的Model属性进行修改，因此该Model对象不可以重复使用。例如，当存在多个分开查询的条件时，我们可以这么来使用Model对象：

```go
user := g.Model("user")
user.Where("status IN(?)", g.Slice{1,2,3})
if vip {
    // 查询条件自动叠加，修改当前模型对象
    user.Where("money>=?", 1000000)
} else {
    // 查询条件自动叠加，修改当前模型对象
    user.Where("money<?",  1000000)
}
//  vip: SELECT * FROM user WHERE status IN(1,2,3) AND money >= 1000000
// !vip: SELECT * FROM user WHERE status IN(1,2,3) AND money < 1000000
r, err := user.All()
//  vip: SELECT COUNT(1) FROM user WHERE status IN(1,2,3) AND money >= 1000000
// !vip: SELECT COUNT(1) FROM user WHERE status IN(1,2,3) AND money < 1000000
n, err := user.Count()
```

​		可以看到，如果是分开执行链式操作，链式的每一个操作都会修改已有的Model对象，查询条件会自动叠加，因此user对象不可重复使用，否则条件会不停叠加。并且在这种使用方式中，每次我们需要操作user用户表，都得使用g.DB().Table("user")这样的语法创建一个新的user模型对象，相对来说会比较繁琐。

#### 1.3.1、Clone方法

​		此外，我们也可以手动调动Clone方法克隆当前模型，创建一个新的模型来实现链式安全，由于是新的模型对象，因此并不担心会修改已有的模型对象的问题。例如：

```go
// 定义一个用户模型单例
user := g.Model("user")
// 克隆一个新的用户模型
m := user.Clone()
m.Where("status IN(?)", g.Slice{1,2,3})
if vip {
    m.And("money>=?", 1000000)
} else {
    m.And("money<?",  1000000)
}
//  vip: SELECT * FROM user WHERE status IN(1,2,3) AND money >= 1000000
// !vip: SELECT * FROM user WHERE status IN(1,2,3) AND money < 1000000
r, err := m.All()
//  vip: SELECT COUNT(1) FROM user WHERE status IN(1,2,3) AND money >= 1000000
// !vip: SELECT COUNT(1) FROM user WHERE status IN(1,2,3) AND money < 1000000
n, err := m.Count()
```

#### 1.3.2、Safe方法

​		当然，我们可以通过Safe方法设置当前模型为链式安全的对象，后续的每一个链式操作都将返回一个新的Model对象，该Model对象可重复使用。但需要特别注意的是，模型属性的修改，或者操作条件的叠加，需要通过变量赋值的方式（m = m.xxx）覆盖原有的模型对象来实现。例如：

```go
// 定义一个用户模型单例
user := g.Model("user").Safe()
m := user.Where("status IN(?)", g.Slice{1,2,3})
if vip {
    // 查询条件通过赋值叠加
    m = m.And("money>=?", 1000000)
} else {
    // 查询条件通过赋值叠加
    m = m.And("money<?",  1000000)
}
//  vip: SELECT * FROM user WHERE status IN(1,2,3) AND money >= 1000000
// !vip: SELECT * FROM user WHERE status IN(1,2,3) AND money < 1000000
r, err := m.All()
//  vip: SELECT COUNT(1) FROM user WHERE status IN(1,2,3) AND money >= 1000000
// !vip: SELECT COUNT(1) FROM user WHERE status IN(1,2,3) AND money < 1000000
n, err := m.Count()
```

​		可以看到，示例中的用户模型单例对象user可以重复使用，而不用担心被“污染”的问题。在这种链式安全的方式下，我们可以创建一个用户单例对象user，并且可以重复使用到后续的各种查询中。但是存在多个查询条件时，条件的叠加需要通过模型赋值操作（m = m.xxx）来实现。

### 1.4、写入保存

​		这几个链式操作方法用于数据的写入，并且支持自动的单条或者批量的数据写入，区别如下：

- **Insert**

​		使用INSERT INTO语句进行数据库写入，<u>如果写入的数据中存在主键或者唯一索引时，返回失败，否则写入一  条新数据。</u>

- **Replace**

​		使用REPLACE INTO语句进行数据库写入，如果写入的数据中存在主键或者唯一索引时，<u>会删除原有的记录，必定会写入一条新记录。</u>

- **Save**

​		使用INSERT INTO语句进行数据库写入，<u>如果写入的数据中存在主键或者唯一索引时，更新原有数据，否则写入一条新数据。</u>

​		这几个方法往往需要结合Data方法使用，该方法用于传递数据参数，用于数据写入/更新等写操作，支持的参数为string/map/slice/struct/*struct。例如，在进行Insert操作时，开发者可以传递任意的map类型，如: map[string]string/map[string]interface{}/map[interface{}]interface{}等等，也可以传递任意的struct/*struct/[]struct/[]*struct类型。此外，这几个方法的参数列表也支持直接的data参数输入，该参数Data方法参数一致。

- **InsertIgnore**
  		从goframe v1.9.0版本开始，goframe的ORM提供了一个常用写入方法InsertIgnore，用于写入数据时如果写入的数据中存在主键或者唯一索引时，忽略错误继续执行写入。该方法定义如下：

```go
func (m *Model) InsertIgnore(data ...interface{}) (result sql.Result, err error)
```

- **InsertAndGetId**

​		从goframe v1.15.7版本开始，goframe的ORM同时也提供了一个常用写入方法InsertAndGetId，用于写入数据时并直接返回自增字段的ID。该方法定义如下：

```go
func (m *Model) InsertAndGetId(data ...interface{}) (lastInsertId int64, err error)
```

- **OnDuplicate/OnDuplicateEx**

​		OnDuplicate/OnDuplicateEx方法用于指定Save方法的更新/不更新字段，参数为字符串、字符串数组、Map。例如：

```go
OnDuplicate("nickname, age")
OnDuplicate("nickname", "age")
OnDuplicate(g.Map{
    "nickname": gdb.Raw("CONCAT('name_', VALUES(`nickname`))"),
})
OnDuplicate(g.Map{
    "nickname": "passport",
})
```

​	其中OnDuplicateEx用于排除指定忽略更新的字段，排除的字段需要在写入的数据集合中。

- **RawSQL语句嵌入**

​		gdb.Raw是字符串类型，该类型的参数将会直接作为SQL片段嵌入到提交到底层的SQL语句中，不会被自动

换为字符串参数类型、也不会被当做预处理参数。例如：

```go
// INSERT INTO `user`(`id`,`passport`,`password`,`nickname`,`create_time`) VALUES('id+2','john','123456','now()')
g.Model("user").Data(g.Map{
	"id":          "id+2",
	"passport":    "john",
	"password":    "123456",
	"nickname":    "JohnGuo",
	"create_time": "now()",
}).Insert()
// 执行报错：Error Code: 1136. Column count doesn't match value count at row 1
```

使用gdb.Raw改造后：

```go
// INSERT INTO `user`(`id`,`passport`,`password`,`nickname`,`create_time`) VALUES(id+2,'john','123456',now())
g.Model("user").Data(g.Map{
	"id":          gdb.Raw("id+2"),
	"passport":    "john",
	"password":    "123456",
	"nickname":    "JohnGuo",
	"create_time": gdb.Raw("now()"),
}).Insert()
```

### 1.5、更新删除

#### 1.5.1、Update

​		Update用于数据的更新，往往需要结合Data及Where方法共同使用。Data方法用于指定需要更新的数据，Where方法用于指定更新的条件范围。同时，Update方法也支持直接给定数据和条件参数。

使用示例：

```go
// UPDATE `user` SET `name`='john guo' WHERE name='john'
g.Model("user").Data(g.Map{"name" : "john guo"}).Where("name", "john").Update()
g.Model("user").Data("name='john guo'").Where("name", "john").Update()
// UPDATE `user` SET `status`=1 ORDER BY `login_time` asc LIMIT 10
g.Model("user").Data("status", 1).Order("login_time asc").Limit(10).Update()

// UPDATE `user` SET `status`=1 WHERE 1
g.Model("user").Data("status=1").Where(1).Update()
g.Model("user").Data("status", 1).Where(1).Update()
g.Model("user").Data(g.Map{"status" : 1}).Where(1).Update()
```

也可以直接给Update方法传递data及where参数：

```go
// UPDATE `user` SET `name`='john guo' WHERE name='john'
g.Model("user").Update(g.Map{"name" : "john guo"}, "name", "john")
g.Model("user").Update("name='john guo'", "name", "john")

// UPDATE `user` SET `status`=1 WHERE 1
g.Model("user").Update("status=1", 1)
g.Model("user").Update(g.Map{"status" : 1}, 1)
```

#### 1.5.2、Counter

可以使用Counter类型参数对特定的字段进行数值操作，例如：增加、减少操作。

Counter数据结构定义：

```go
// Counter  is the type for update count.
type Counter struct {
	Field string
	Value float64
}
```

字段自增：

```go
updateData := g.Map{
	"views": &gdb.Counter{ 
        Field: "views", 
        Value: 1,
    },
}
// UPDATE `article` SET `views`=`views`+1 WHERE `id`=1
result, err := db.Update("article", updateData, "id", 1)
```

Counter也可以实现非自身字段的自增，例如：

```go
updateData := g.Map{
	"views": &gdb.Counter{ 
        Field: "clicks", 
        Value: 1,
    },
}
// UPDATE `article` SET `views`=`clicks`+1 WHERE `id`=1
result, err := db.Update("article", updateData, "id", 1)
```

#### 1.5.3、Increment/Decrement

​		我们可以通过Increment和Decrement方法实现对指定字段的自增/自减常用操作。两个方法的定义如下：

```go
// Increment increments a column's value by a given amount.
func (m *Model) Increment(column string, amount float64) (sql.Result, error)

// Decrement decrements a column's value by a given amount.
func (m *Model) Decrement(column string, amount float64) (sql.Result, error)
```

```go
// UPDATE `article` SET `views`=`views`+10000 WHERE `id`=1
g.Model("article").Where("id", 1).Increment("views", 10000)
// UPDATE `article` SET `views`=`views`-10000 WHERE `id`=1
g.Model("article").Where("id", 1).Decrement("views", 10000)
```

#### 1.5.4、RawSQL语句嵌入

​		gdb.Raw是字符串类型，该类型的参数将会直接作为SQL片段嵌入到提交到底层的SQL语句中，不会被自动转换为字符串参数类型、也不会被当做预处理参数。例如：

```go
// UPDATE `user` SET login_count='login_count+1',update_time='now()' WHERE id=1
g.Model("user").Data(g.Map{
    "login_count": "login_count+1",
    "update_time": "now()",
}).Where("id", 1).Update()
// 执行报错：Error Code: 1136. Column count doesn't match value count at row 1
```

使用gdb.Raw改造后：

```go
// UPDATE `user` SET login_count=login_count+1,update_time=now() WHERE id=1
g.Model("user").Data(g.Map{
    "login_count": gdb.Raw("login_count+1"),
    "update_time": gdb.Raw("now()"),
}).Where("id", 1).Update()
```

#### 1.5.5、Delete

```go
// DELETE FROM `user` WHERE uid=10
g.Model("user").Where("uid", 10).Delete()
// DELETE FROM `user` ORDER BY `login_time` asc LIMIT 10
g.Model("user").Order("login_time asc").Limit(10).Delete()
```

### 1.6、时间维护

​		gdb模块支持对数据记录的写入、更新、删除时间自动填充，提高开发维护效率。为了便于时间字段名称、类型的统一维护，如果使用该特性，我们约定：

1. 字段应当设置允许值为null。
2. 字段的类型必须为时间类型，如:date,  datetime,  timestamp。不支持数字类型字段，如int。
3. 字段的名称不支持自定义设置，并且固定名称约定为：
   - created_at用于记录创建时更新，仅会写入一次。
   - updated_at用于记录修改时更新，每次记录变更时更新。
   - deleted_at用于记录的软删除特性，只有当记录删除时会写入一次。

​        字段名称其实不区分大小写，也会忽略特殊字符，例如CreatedAt,  UpdatedAt,  DeletedAt也是支持的。此外，时间字段名称可以通过配置文件进行自定义修改，并可使用TimeMaintainDisabled配置完整关闭该特性。

​		当数据表包含created_at、updated_at、deleted_at任意一个或多个字段时，该特性自动启用。

​		以下的示例中，我们默认示例中的数据表均包含了这3个字段。

#### 1.6.1、created_at

​		在执行Insert/InsertIgnore/BatchInsert/BatchInsertIgnore方法时自动写入该时间，随后保持不变。

```go
// INSERT INTO `user`(`name`,`created_at`,`updated_at`) VALUES('john', `2020-06-06 21:00:00`, `2020-06-06 21:00:00`)
db.Model("user").Data(g.Map{"name": "john"}).Insert()

// INSERT IGNORE INTO `user`(`uid`,`name`,`created_at`,`updated_at`) VALUES(10000,'john', `2020-06-06 21:00:00`, `2020-06-06 21:00:00`)
db.Model("user").Data(g.Map{"uid": 10000, "name": "john"}).InsertIgnore()

// REPLACE INTO `user`(`uid`,`name`,`created_at`,`updated_at`) VALUES(10000,'john', `2020-06-06 21:00:00`, `2020-06-06 21:00:00`)
db.Model("user").Data(g.Map{"uid": 10000, "name": "john"}).Replace()

// INSERT INTO `user`(`uid`,`name`,`created_at`,`updated_at`) VALUES(10001,'john', `2020-06-06 21:00:00`, `2020-06-06 21:00:00`) ON DUPLICATE KEY UPDATE `uid`=VALUES(`uid`),`name`=VALUES(`name`),`updated_at`=VALUES(`updated_at`)
db.Model("user").Data(g.Map{"uid": 10001, "name": "john"}).Save()
```

#### 1.6.2、updated_at

​		在执行Insert/InsertIgnore/BatchInsert/BatchInsertIgnore方法时自动写入该时间，在执行Save/Update时更新该时间（注意当写入数据存在时会更新updated_at时间，不会更新created_at时间）。

```go
// UPDATE `user` SET `name`='john guo',`updated_at`='2020-06-06 21:00:00' WHERE name='john'
db.Model("user").Data(g.Map{"name" : "john guo"}).Where("name", "john").Update()

// UPDATE `user` SET `status`=1,`updated_at`='2020-06-06 21:00:00' ORDER BY `login_time` asc LIMIT 10
db.Model("user").Data("status", 1).Order("login_time asc").Limit(10).Update()

// INSERT INTO `user`(`id`,`name`,`update_at`) VALUES(1,'john guo','2020-12-29 20:16:14') ON DUPLICATE KEY UPDATE `id`=VALUES(`id`),`name`=VALUES(`name`),`update_at`=VALUES(`update_at`)
db.Model("user").Data(g.Map{"id": 1, "name": "john guo"}).Save()
```

#### 1.6.3、deleted_at

​		软删除会稍微比较复杂一些，当软删除存在时，所有的查询语句都将会自动加上deleted_at的条件。

```go
// UPDATE `user` SET `deleted_at`='2020-06-06 21:00:00' WHERE uid=10
db.Model("user").Where("uid", 10).Delete()
```

​		查询的时候会发生一些变化，例如：

```go
// SELECT * FROM `user` WHERE uid>1 AND `deleted_at` IS NULL
db.Model("user").Where("uid>?", 1).All()
```

### 2、常见查询操作

### 2.1、查询条件（Where/WhereOr/WhereNot）

ORM组件提供了一些常用的条件查询方法，并且条件方法支持多种数据类型输入。

```go
func (m *Model) Where(where interface{}, args...interface{}) *Model
func (m *Model) Wheref(format string, args ...interface{}) *Model
func (m *Model) WherePri(where interface{}, args ...interface{}) *Model
func (m *Model) WhereBetween(column string, min, max interface{}) *Model
func (m *Model) WhereLike(column string, like interface{}) *Model
func (m *Model) WhereIn(column string, in interface{}) *Model
func (m *Model) WhereNull(columns ...string) *Model
func (m *Model) WhereLT(column string, value interface{}) *Model
func (m *Model) WhereLTE(column string, value interface{}) *Model
func (m *Model) WhereGT(column string, value interface{}) *Model
func (m *Model) WhereGTE(column string, value interface{}) *Model

func (m *Model) WhereNotBetween(column string, min, max interface{}) *Model
func (m *Model) WhereNotLike(column string, like interface{}) *Model
func (m *Model) WhereNotIn(column string, in interface{}) *Model
func (m *Model) WhereNotNull(columns ...string) *Model

func (m *Model) WhereOr(where interface{}, args ...interface{}) *Model
func (m *Model) WhereOrBetween(column string, min, max interface{}) *Model
func (m *Model) WhereOrLike(column string, like interface{}) *Model
func (m *Model) WhereOrIn(column string, in interface{}) *Model
func (m *Model) WhereOrNull(columns ...string) *Model
func (m *Model) WhereOrLT(column string, value interface{}) *Model 
func (m *Model) WhereOrLTE(column string, value interface{}) *Model 
func (m *Model) WhereOrGT(column string, value interface{}) *Model 
func (m *Model) WhereOrGTE(column string, value interface{}) *Model

func (m *Model) WhereOrNotBetween(column string, min, max interface{}) *Model
func (m *Model) WhereOrNotLike(column string, like interface{}) *Model
func (m *Model) WhereOrNotIn(column string, in interface{}) *Model
func (m *Model) WhereOrNotNull(columns ...string) *Model
```

​		Where条件参数推荐使用字符串的参数传递方式（并使用?占位符预处理），因为map/struct类型作为查询参数无法保证顺序性，且在部分情况下（数据库有时会帮助你自动进行查询索引优化），数据库的索引和你传递的查询条件顺序有一定关系。

​		当使用多个Where方法连接查询条件时，多个条件之间使用And进行连接。 此外，当存在多个查询条件时，gdb会默认将多个条件分别使用()符号进行包含，这种设计可以非常友好地支持查询条件分组。

​		使用示例：

```go
// WHERE `uid`=1
Where("uid=1")
Where("uid", 1)
Where("uid=?", 1)
Where(g.Map{"uid" : 1})
// WHERE `uid` <= 1000 AND `age` >= 18
Where(g.Map{
    "uid <=" : 1000,
    "age >=" : 18,
})

// WHERE (`uid` <= 1000) AND (`age` >= 18)
Where("uid <=?", 1000).Where("age >=?", 18)

// WHERE `level`=1 OR `money`>=1000000
Where("level=? OR money >=?", 1, 1000000)

// WHERE (`level`=1) OR (`money`>=1000000)
Where("level", 1).WhereOr("money >=", 1000000)

// WHERE `uid` IN(1,2,3)
Where("uid IN(?)", g.Slice{1,2,3})
```

使用struct参数的示例，其中orm的tag用于指定struct属性与表字段的映射关系：

```go
type Condition struct{
    Sex int `orm:"sex"`
    Age int `orm:"age"`
}
Where(Condition{1, 18})
// WHERE `sex`=1 AND `age`=18
```

#### 2.1.1、Where + string，条件参数使用字符串和预处理

```go
// 查询多条记录并使用Limit分页
// SELECT * FROM user WHERE uid>1 LIMIT 0,10
g.Model("user").Where("uid > ?", 1).Limit(0, 10).All()

// 使用Fields方法查询指定字段
// 未使用Fields方法指定查询字段时，默认查询为*
// SELECT uid,name FROM user WHERE uid>1 LIMIT 0,10
g.Model("user").Fileds("uid,name").Where("uid > ?", 1).Limit(0, 10).All()

// 支持多种Where条件参数类型
// SELECT * FROM user WHERE uid=1 LIMIT 1
g.Model("user").Where("uid=1",).One()
g.Model("user").Where("uid", 1).One()
g.Model("user").Where("uid=?", 1).One()

// SELECT * FROM user WHERE (uid=1) AND (name='john') LIMIT 1
g.Model("user").Where("uid", 1).Where("name", "john").One()
g.Model("user").Where("uid=?", 1).And("name=?", "john").One()

// SELECT * FROM user WHERE (uid=1) OR (name='john') LIMIT 1
g.Model("user").Where("uid=?", 1).Or("name=?", "john").One()
```

#### 2.1.2、Where + slice，预处理参数可直接通过slice参数给定

```go
// SELECT * FROM user WHERE age>18 AND name like '%john%'
g.Model("user").Where("age>? AND name like ?", g.Slice{18, "%john%"}).All()

// SELECT * FROM user WHERE status=1
g.Model("user").Where("status=?", g.Slice{1}).All()
```

#### 2.1.3、Where + map，条件参数使用任意map类型传递

```go
// SELECT * FROM user WHERE uid=1 AND name='john' LIMIT 1
g.Model("user").Where(g.Map{"uid" : 1, "name" : "john"}).One()

// SELECT * FROM user WHERE uid=1 AND age>18 LIMIT 1
g.Model("user").Where(g.Map{"uid" : 1, "age>" : 18}).One()
```

```go
condition := g.Map{
    "title like ?"         : "%九寨%",
    "online"               : 1,
    "hits between ? and ?" : g.Slice{1, 10},
    "exp > 0"              : nil,
    "category"             : g.Slice{100, 200},
}
// SELECT * FROM article WHERE title like '%九寨%' AND online=1 AND hits between 1 and 10 AND exp > 0 AND category IN(100,200)
g.Model("article").Where(condition).All()
```



#### 2.1.4、Where + struct/*struct，struct标签支持 orm/json，映射属性到字段名称关系

```go
type User struct {
    Id       int    `json:"uid"`
    UserName string `orm:"name"`
}
// SELECT * FROM user WHERE uid =1 AND name='john' LIMIT 1
g.Model("user").Where(User{ Id : 1, UserName : "john"}).One()

// SELECT * FROM user WHERE uid =1 LIMIT 1
g.Model("user").Where(&User{ Id : 1}).One()
```

#### 2.1.5、Wheref格式化条件字符串

​		在某些场景中，在输入带有字符串的条件语句时，往往需要结合fmt.Sprintf来格式化条件（当然，注意在字符串中使用占位符代替变量的输入而不是直接将变量格式化），因此提供了Where+fmt.Sprintf结合的便捷方法Wheref。使用示例：

```go
// WHERE score > 100 and status in('succeeded','completed')
Wheref(`score > ? and status in (?)`, 100, g.Slice{"succeeded", "completed"})
```

#### 2.1.6、WherePri支持主键的查询条件

​		WherePri方法的功能同Where，但提供了对表主键的智能识别，常用于根据主键的便捷数据查询。假如user表的主键为uid，我们来看一下Where与WherePri的区别：

```go
// WHERE `uid`=1
Where("uid", 1)
WherePri(1)

// WHERE `uid` IN(1,2,3)
Where("uid", g.Slice{1,2,3})
WherePri(g.Slice{1,2,3})
```

​		可以看到，当使用WherePri方法且给定参数为单一的参数基本类型或者slice类型时，将会被识别为主键的查询条件值。

### 2.2、查询结束（All/One/Array/Value/Count）

```go
func (m *Model) All(where ...interface{} (Result, error)
func (m *Model) One(where ...interface{}) (Record, error)
func (m *Model) Array(fieldsAndWhere ...interface{}) ([]Value, error)
func (m *Model) Value(fieldsAndWhere ...interface{}) (Value, error)
func (m *Model) Count(where ...interface{}) (int, error)
func (m *Model) CountColumn(column string) (int, error)
```

**简要说明：**

- All  用于查询并返回多条记录的列表/数组。
- One  用于查询并返回单条记录。
- Array  用于查询指定字段列的数据，返回数组。
- Value  用于查询并返回一个字段值，往往需要结合Fields方法使用。
- Count  用于查询并返回记录数。

​       此外，也可以看得到这四个方法定义中也支持条件参数的直接输入，参数类型与Where方法一致。但需要注意，其中Array和Value方法的参数中至少应该输入字段参数。

```go
// SELECT * FROM `user` WHERE `score`>60
Model("user").Where("score>?", 60).All()

// SELECT * FROM `user` WHERE `score`>60 LIMIT 1
Model("user").Where("score>?", 60).One()

// SELECT `name` FROM `user` WHERE `score`>60
Model("user").Fields("name").Where("score>?", 60).Array()

// SELECT `name` FROM `user` WHERE `uid`=1 LIMIT 1
Model("user").Fields("name").Where("uid", 1).Value()

// SELECT COUNT(1) FROM `user` WHERE `status` IN(1,2,3)
Model("user").Where("status", g.Slice{1,2,3}).Count()
```

#### 2.2.1、Find

```go
func (m *Model) FindAll(where ...interface{}) (Result, error)
func (m *Model) FindOne(where ...interface{}) (Record, error)
func (m *Model) FindArray(fieldsAndWhere ...interface{}) (Value, error)
func (m *Model) FindValue(fieldsAndWhere ...interface{}) (Value, error)
func (m *Model) FindCount(where ...interface{}) (int, error)
func (m *Model) FindScan(pointer interface{}, where ...interface{}) error
```

​		Find*方法包含：FindAll/FindOne/FineValue/FindCount/FindScan，这些方法与All/One/Array/Value/Count/Scan方法的区别在于，当方法直接给定条件参数时，前者的效果与WherePri方法一致；而后者的效果与Where方法一致。也就是说Find*方法的条件参数支持智能主键识别特性。



```go
// SELECT * FROM `scores` WHERE `id`=1
Model("scores").FindAll(1)

// SELECT * FROM `scores` WHERE `id`=1 LIMIT 1
Model("scores").FindOne(1)

// SELECT `name` FROM `scores` WHERE `id`=1
Model("scores").FindArray("name", 1)

// SELECT `name` FROM `scores` WHERE `id`=1 LIMIT 1
Model("user").FindValue("name", 1)

// SELECT COUNT(1) FROM `user`  WHERE `id`=1 
Model("user").FindCount(1)
```

#### 2.2.2、Scan

​		方法支持将查询结果转换为结构体或者结构体数组，Scan方法将会根据给定的参数类型自动识别执行的转换类型。

**struct对象**
		Scan支持将查询结果转换为一个struct对象，查询结果应当是特定的一条记录，并且pointer参数应当为struct对象的指针地址（*struct或者**struct），使用方式例如：

```go
type User struct {
    Id         int
    Passport   string
    Password   string
    NickName   string
    CreateTime *gtime.Time
}
user := User{}
g.Model("user").Where("id", 1).Scan(&user)
```

**struct数组**
		Scan支持将多条查询结果集转换为一个[]struct/[]*struct数组，查询结果应当是多条记录组成的结果集，并且pointer应当为数组的指针地址，使用方式例如：

```go
var users []User
g.Model("user").Scan(&users)
```

### 2.3、分组与排序

Group方法用于查询分组，Order方法用于查询排序

```go
// SELECT COUNT(*) total,age FROM `user` GROUP BY age
g.Model("user").Fields("COUNT(*) total,age").Group("age").All()

// SELECT * FROM `student` ORDER BY class asc,course asc,score desc
g.Model("student").Order("class asc,course asc,score desc").All()

```

```go
// 按照指定字段递增排序
func (m *Model) OrderAsc(column string) *Model
// 按照指定字段递减排序
func (m *Model) OrderDesc(column string) *Model
// 随机排序
func (m *Model) OrderRandom() *Model
```

使用用例

```go
// SELECT `id`,`title` FROM `article` ORDER BY `created_at` ASC
g.Model("article").Fields("id,title").OrderAsc("created_at").All()

// SELECT `id`,`title` FROM `article` ORDER BY `views` DESC
g.Model("article").Fields("id,title").OrderDesc("views").All()

// SELECT `id`,`title` FROM `article` ORDER BY RAND()
g.Model("article").Fields("id,title").OrderRandom().All()
```

**HAVING操作**

```go
// SELECT COUNT(*) total,age FROM `user` GROUP BY age HAVING total>100
g.Model("user").Fields("COUNT(*) total,age").Group("age").Having("total>100").All()

// SELECT * FROM `student` ORDER BY class HAVING score>60
g.Model("student").Order("class").Having("score>?", 60).All()
```

### 2.4、子查询特性

ORM组件目前支持常见的三种语法的子查询：Where子查询、Having子查询及From子查询。

#### **2.4.1、Where子查询**

```go
// 获取默认配置的数据库对象(配置名称为"default")
db := g.DB()

db.Model("orders").Where("amount > ?", db.Model("orders").Fields("AVG(amount)")).Scan(&orders)
// SELECT * FROM "orders" WHERE amount > (SELECT AVG(amount) FROM "orders")
```

#### 2.4.2、Having子查询

```go
subQuery := db.Model("users").Fields("AVG(age)").WhereLike("name", "name%")
db.Model("users").Fields("AVG(age) as avgage").Group("name").Having("AVG(age) > ?", subQuery).Scan(&results)
// SELECT AVG(age) as avgage FROM `users` GROUP BY `name` HAVING AVG(age) > (SELECT AVG(age) FROM `users` WHERE name LIKE "name%")
```

#### 2.4.3、From子查询

```go
db.Model("? as u", db.Model("users").Fields("name", "age")).Where("age", 18).Scan(&users)
// SELECT * FROM (SELECT `name`,`age` FROM `users`) as u WHERE `age` = 18

subQuery1 := db.Model("users").Fields("name")
subQuery2 := db.Model("pets").Fields("name")
db.Model("? as u, ? as p", subQuery1, subQuery2).Scan(&users)
// SELECT * FROM (SELECT `name` FROM `users`) as u, (SELECT `name` FROM `pets`) as p
```

## 3、其它查询

### 3.1、In

```go
// SELECT * FROM user WHERE uid IN(100,10000,90000)
db.Model("user").Where("uid IN(?,?,?)", 100, 10000, 90000).All()
db.Model("user").Where("uid", g.Slice{100, 10000, 90000}).All()

// SELECT * FROM user WHERE gender=1 AND uid IN(100,10000,90000)
db.Model("user").Where("gender=? AND uid IN(?)", 1, g.Slice{100, 10000, 90000}).All()

// SELECT COUNT(*) FROM user WHERE age in(18,50)
db.Model("user").Where("age IN(?,?)", 18, 50).Count()
db.Model("user").Where("age", g.Slice{18, 50}).Count()
```

```go


// SELECT * FROM `user` WHERE `gender`=1 AND `type` IN(1,2,3)
db.Model("user").Where("gender", 1).WhereIn("type", g.Slice{1,2,3}).All()

// SELECT * FROM `user` WHERE `gender`=1 AND `type` NOT IN(1,2,3)
db.Model("user").Where("gender", 1).WhereNotIn("type", g.Slice{1,2,3}).All()

// SELECT * FROM `user` WHERE `gender`=1 OR `type` IN(1,2,3)
db.Model("user").Where("gender", 1).WhereOrIn("type", g.Slice{1,2,3}).All()

// SELECT * FROM `user` WHERE `gender`=1 OR `type` NOT IN(1,2,3)
db.Model("user").Where("gender", 1).WhereOrNotIn("type", g.Slice{1,2,3}).All()
```

### 3.2、like

```go
// SELECT * FROM `user` WHERE name like '%john%'
db.Model("user").Where("name like ?", "%john%").All()
// SELECT * FROM `user` WHERE birthday like '1990-%'
db.Model("user").Where("birthday like ?", "1990-%").All()
```

也可以这样使用

```go
func (m *Model) WhereLike(column string, like interface{}) *Model
func (m *Model) WhereNotLike(column string, like interface{}) *Model
func (m *Model) WhereOrLike(column string, like interface{}) *Model
func (m *Model) WhereOrNotLike(column string, like interface{}) *Model
```

```go
// SELECT * FROM `user` WHERE `gender`=1 AND `name` LIKE 'john%'
db.Model("user").Where("gender", 1).WhereLike("name", "john%").All()

// SELECT * FROM `user` WHERE `gender`=1 AND `name` NOT LIKE 'john%'
db.Model("user").Where("gender", 1).WhereNotLike("name", "john%").All()

// SELECT * FROM `user` WHERE `gender`=1 OR `name` LIKE 'john%'
db.Model("user").Where("gender", 1).WhereOrLike("name", "john%").All()

// SELECT * FROM `user` WHERE `gender`=1 OR `name` NOT LIKE 'john%'
db.Model("user").Where("gender", 1).WhereOrNotLike("name", "john%").All()
```

### 3.3、min/max/avg/sum

可以直接使用的filed字段上：

```go
// SELECT MIN(score) FROM `user` WHERE `uid`=1
db.Model("user").Fields("MIN(score)").Where("uid", 1).Value()

// SELECT MAX(score) FROM `user` WHERE `uid`=1
db.Model("user").Fields("MAX(score)").Where("uid", 1).Value()

// SELECT AVG(score) FROM `user` WHERE `uid`=1
db.Model("user").Fields("AVG(score)").Where("uid", 1).Value()

// SELECT SUM(score) FROM `user` WHERE `uid`=1 
db.Model("user").Fields("SUM(score)").Where("uid", 1).Value()
```

​		从goframe v1.16版本开始，goframe的ORM同时也提供了常用统计方法Min/Max/Avg/Sum方法，用于常用的字段统计查询。方法定义如下：

```go
func (m *Model) Min(column string) (float64, error)
func (m *Model) Max(column string) (float64, error)
func (m *Model) Avg(column string) (float64, error)
func (m *Model) Sum(column string) (float64, error)
```

```go
// SELECT MIN(`score`) FROM `user` WHERE `uid`=1
db.Model("user").Where("uid", 1).Min("score")

// SELECT MAX(`score`) FROM `user` WHERE `uid`=1
db.Model("user").Where("uid", 1).Max("score")

// SELECT AVG(`score`) FROM `user` WHERE `uid`=1
db.Model("user").Where("uid", 1).Avg("score")

// SELECT SUM(`score`) FROM `user` WHERE `uid`=1
db.Model("user").Where("uid", 1).Sum("score")
```

### 3.4、count查询

```go
// SELECT COUNT(1) FROM `user` WHERE `birthday`='1990-10-01'
db.Model("user").Where("birthday", "1990-10-01").Count()
// SELECT COUNT(uid) FROM `user` WHERE `birthday`='1990-10-01'
db.Model("user").Fields("uid").Where("birthday", "1990-10-01").Count()
db.Model("user").Where("birthday", "1990-10-01").CountColumn("uid")
```

### 3.5、distinct查询

```go
// SELECT DISTINCT uid,name FROM `user`
db.Model("user").Fields("DISTINCT uid,name").All()
// SELECT COUNT(DISTINCT uid,name) FROM `user`
db.Model("user").Fields("DISTINCT uid,name").Count()
```

```go
// SELECT COUNT(DISTINCT `name`) FROM `user`
db.Model("user").Distinct().CountColumn("name")

// SELECT COUNT(DISTINCT uid,name) FROM `user`
db.Model("user").Distinct().CountColumn("uid,name")
```

### 3.6、between查询

```go
// SELECT * FROM `user ` WHERE age between 18 and 20
db.Model("user").Where("age between ? and ?", 18, 20).All()
```

goframe1.6提供了额外的能力

```go
func (m *Model) WhereBetween(column string, min, max interface{}) *Model
func (m *Model) WhereNotBetween(column string, min, max interface{}) *Model
func (m *Model) WhereOrBetween(column string, min, max interface{}) *Model
func (m *Model) WhereOrNotBetween(column string, min, max interface{}) *Model
```

```go
// SELECT * FROM `user` WHERE `gender`=0 AND `age` BETWEEN 16 AND 20
db.Model("user").Where("gender", 0).WhereBetween("age", 16, 20).All()

// SELECT * FROM `user` WHERE `gender`=0 AND `age` NOT BETWEEN 16 AND 20
db.Model("user").Where("gender", 0).WhereNotBetween("age", 16, 20).All()

// SELECT * FROM `user` WHERE `gender`=0 OR `age` BETWEEN 16 AND 20
db.Model("user").Where("gender", 0).WhereOrBetween("age", 16, 20).All()

// SELECT * FROM `user` WHERE `gender`=0 OR `age` NOT BETWEEN 16 AND 20
db.Model("user").Where("gender", 0).WhereOrNotBetween("age", 16, 20).All()
```

### 3.7、null查询

```go
func (m *Model) WhereNull(columns ...string) *Model
func (m *Model) WhereNotNull(columns ...string) *Model
func (m *Model) WhereOrNull(columns ...string) *Model
func (m *Model) WhereOrNotNull(columns ...string) *Model
```

```go
// SELECT * FROM `user` WHERE `created_at` > '2021-05-01 00:00:00' AND `inviter` IS NULL
db.Model("user").Where("created_at>?", gtime.New("2021-05-01")).WhereNull("inviter").All()

// SELECT * FROM `user` WHERE `created_at` > '2021-05-01 00:00:00' AND `inviter` IS NOT NULL
db.Model("user").Where("created_at>?", gtime.New("2021-05-01")).WhereNotNull("inviter").All()

// SELECT * FROM `user` WHERE `created_at` > '2021-05-01 00:00:00' OR `inviter` IS NULL
db.Model("user").Where("created_at>?", gtime.New("2021-05-01")).WhereOrNull("inviter").All()

// SELECT * FROM `user` WHERE `created_at` > '2021-05-01 00:00:00' OR `inviter` IS NOT NULL
db.Model("user").Where("created_at>?", gtime.New("2021-05-01")).WhereOrNotNull("inviter").All()
```

## 4、数据读取(Scanlist)

```go
// 用户表
type EntityUser struct {
    Uid  int    `orm:"uid"`
    Name string `orm:"name"`
}
// 用户详情
type EntityUserDetail struct {
    Uid     int    `orm:"uid"`
    Address string `orm:"address"`
}
// 用户学分
type EntityUserScores struct {
    Id     int    `orm:"id"`
    Uid    int    `orm:"uid"`
    Score  int    `orm:"score"`
    Course string `orm:"course"`
}
// 组合模型，用户信息
type Entity struct {
    User       *EntityUser
    UserDetail *EntityUserDetail
    UserScores []*EntityUserScores
}
```

​		其中，EntityUser, EntityUserDetail, EntityUserScores分别对应的是用户表、用户详情、用户学分数据表的数据模型。Entity是一个组合模型，对应的是一个用户的所有详细信息。

### 4.1、数据写入

​	写入数据时涉及到简单的数据库事务即可。

```go
err := db.Transaction(func(tx *gdb.TX) error {
    r, err := tx.Table("user").Save(EntityUser{
        Name: "john",
    })
    if err != nil {
        return err
    }
    uid, err := r.LastInsertId()
    if err != nil {
        return err
    }
    _, err = tx.Table("user_detail").Save(EntityUserDetail{
        Uid:     int(uid),
        Address: "Beijing DongZhiMen #66",
    })
    if err != nil {
        return err
    }
    _, err = tx.Table("user_scores").Save(g.Slice{
        EntityUserScores{Uid: int(uid), Score: 100, Course: "math"},
        EntityUserScores{Uid: int(uid), Score: 99, Course: "physics"},
    })
    return err
})
```

### 4.2、数据查询

#### 4.2.1、单条数据记录

​		查询单条模型数据比较简单，直接使用Scan方法即可，该方法会自动识别绑定查询结果到单个对象属性还是数组对象属性中。例如：

```go
// 定义用户列表
var user Entity
// 查询用户基础数据
// SELECT * FROM `user` WHERE `name`='john'
err := db.Table("user").Scan(&user.User, "name", "john")
if err != nil {
    return err
}
// 查询用户详情数据
// SELECT * FROM `user_detail` WHERE `uid`=1
err := db.Table("user_detail").Scan(&user.UserDetail, "uid", user.User.Uid)
// 查询用户学分数据
// SELECT * FROM `user_scores` WHERE `uid`=1
err := db.Table("user_scores").Scan(&user.UserScores, "uid", user.User.Uid)
```

#### 4.2.2、多条数据记录

​		查询多条数据记录并绑定数据到数据模型数组中，需要使用到ScanList方法，该方法会需要用户指定结果字段与模型属性的关系，随后底层会遍历数组并自动执行数据绑定。例如：

```go
// 定义用户列表
var users []Entity
// 查询用户基础数据
// SELECT * FROM `user`
err := db.Table("user").ScanList(&users, "User")
// 查询用户详情数据
// SELECT * FROM `user_detail` WHERE `uid` IN(1,2)
err := db.Table("user_detail").
       Where("uid", gdb.ListItemValuesUnique(users, "User", "Uid")).
       ScanList(&users, "UserDetail", "User", "uid:Uid")
// 查询用户学分数据
// SELECT * FROM `user_scores` WHERE `uid` IN(1,2)
err := db.Table("user_scores").
       Where("uid", gdb.ListItemValuesUnique(users, "User", "Uid")).
       ScanList(&users, "UserScores", "User", "uid:Uid")
```

这其中涉及到两个比较重要的方法：

- **ScanList**

方法定义：

```go
// ScanList converts <r> to struct slice which contains other complex struct attributes.
// Note that the parameter <listPointer> should be type of *[]struct/*[]*struct.
// Usage example:
//
// type Entity struct {
// 	   User       *EntityUser
// 	   UserDetail *EntityUserDetail
//	   UserScores []*EntityUserScores
// }
// var users []*Entity
// or
// var users []Entity
//
// ScanList(&users, "User")
// ScanList(&users, "UserDetail", "User", "uid:Uid")
// ScanList(&users, "UserScores", "User", "uid:Uid")
// The parameters "User"/"UserDetail"/"UserScores" in the example codes specify the target attribute struct
// that current result will be bound to.
// The "uid" in the example codes is the table field name of the result, and the "Uid" is the relational
// struct attribute name. It automatically calculates the HasOne/HasMany relationship with given <relation>
// parameter.
// See the example or unit testing cases for clear understanding for this function.
func (m *Model) ScanList(listPointer interface{}, attributeName string, relation ...string) (err error)
```

该方法用于将查询到的数组数据绑定到指定的列表上，例如：

**ScanList(&users, "User")**
		表示将查询到的用户信息数组数据绑定到users列表中每一项的User属性上。

**ScanList(&users, "UserDetail", "User", "uid:Uid")**
		表示将查询到用户详情数组数据绑定到users列表中每一项的UserDetail属性上，并且和另一个User对象属性通过uid:Uid的字段:属性关联，内部将会根据这一关联关系自动进行数据绑定。其中uid:Uid前面的uid表示查询结果字段中的uid字段，后面的Uid表示目标关联对象中的Uid属性。

**ScanList(&users, "UserScores", "User", "uid:Uid")**
		表示将查询到用户详情数组数据绑定到users列表中每一项的UserScores属性上，并且和另一个User对象属性通过uid:Uid的字段:属性关联，内部将会根据这一关联关系自动进行数据绑定。由于UserScores是一个数组类型[]*EntityUserScores，因此该方法内部可以自动识别到User到UserScores其实是1:N的关系，自动完成数据绑定。

需要提醒的是，如果关联数据中对应的关联属性数据不存在，那么该属性不会被初始化并将保持nil。

- **ListItemValues/ListItemValuesUnique**

```go
// ListItemValues retrieves and returns the elements of all item struct/map with key <key>.
// Note that the parameter <list> should be type of slice which contains elements of map or struct,
// or else it returns an empty slice.
//
// The parameter <list> supports types like:
// []map[string]interface{}
// []map[string]sub-map
// []struct
// []struct:sub-struct
// Note that the sub-map/sub-struct makes sense only if the optional parameter <subKey> is given.
func ListItemValues(list interface{}, key interface{}, subKey ...interface{}) (values []interface{})  

// ListItemValuesUnique retrieves and returns the unique elements of all struct/map with key <key>.
// Note that the parameter <list> should be type of slice which contains elements of map or struct,
// or else it returns an empty slice.
// See gutil.ListItemValuesUnique.
func ListItemValuesUnique(list interface{}, key string, subKey ...interface{}) []interface{}
```



​		ListItemValuesUnique与ListItemValues方法的区别在于过滤重复的返回值，保证返回的列表数据中不带有重复值。这两个方法都会在当给定的列表中包含struct/map数据项时，用于获取指定属性/键名的数据值，构造成数组[]interface{}返回。示例：

1. gdb.ListItemValuesUnique(users, "Uid")用于获取users数组中，每一个Uid属性，构造成[]interface{}数组返回。这里以便根据uid构造成SELECT...IN...查询。
2. gdb.ListItemValuesUnique(users, "User", "Uid")用于获取users数组中，每一个User属性项中的Uid属性，构造成[]interface{}数组返回。这里以便根据uid构造成SELECT...IN...查询。

## 5、数据读取（With）

可以把With特性看做ScanList与模型关联关系维护的一种结合和改进。

```sql
# 用户表

CREATE TABLE `user` (
  id int(10) unsigned NOT NULL AUTO_INCREMENT,
  name varchar(45) NOT NULL,
  PRIMARY KEY (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

# 用户详情

CREATE TABLE `user_detail` (
  uid  int(10) unsigned NOT NULL AUTO_INCREMENT,
  address varchar(45) NOT NULL,
  PRIMARY KEY (uid)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

# 用户学分

CREATE TABLE `user_scores` (
  id int(10) unsigned NOT NULL AUTO_INCREMENT,
  uid int(10) unsigned NOT NULL,
  score int(10) unsigned NOT NULL,
  PRIMARY KEY (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
```

根据表定义，我们可以得知：

- 用户表与用户详情是1:1关系。
- 用户表与用户学分是1:N关系。

​     这里并没有演示N:N的关系，因为相比较于1:N的查询只是多了一次关联、或者一次查询，最终处理方式和1:N类似。

```go
// 用户详情
type UserDetail struct {
	gmeta.Meta `orm:"table:user_detail"`
	Uid        int    `json:"uid"`
	Address    string `json:"address"`
}
// 用户学分
type UserScores struct {
	gmeta.Meta `orm:"table:user_scores"`
	Id         int `json:"id"`
	Uid        int `json:"uid"`
	Score      int `json:"score"`
}
// 用户信息
type User struct {
	gmeta.Meta `orm:"table:user"`
	Id         int           `json:"id"`
	Name       string        `json:"name"`
	UserDetail *UserDetail   `orm:"with:uid=id"`
	UserScores []*UserScores `orm:"with:uid=id"`
}
```



```go
db.Transaction(func(tx *gdb.TX) error {
	for i := 1; i <= 5; i++ {
		// User.
		user := User{
			Name: fmt.Sprintf(`name_%d`, i),
		}
		lastInsertId, err := db.Model(user).Data(user).OmitEmpty().InsertAndGetId()
		if err != nil {
			return err
		}
		// Detail.
		userDetail := UserDetail{
			Uid:     int(lastInsertId),
			Address: fmt.Sprintf(`address_%d`, lastInsertId),
		}
		_, err = db.Model(userDetail).Data(userDetail).OmitEmpty().Insert()
		if err != nil {
			return err
		}
		// Scores.
		for j := 1; j <= 5; j++ {
			userScore := UserScores{
				Uid:   int(lastInsertId),
				Score: j,
			}
			_, err = db.Model(userScore).Data(userScore).OmitEmpty().Insert()
			if err != nil {
				return err
			}
		}
	}
	return nil
})
```

```
执行结束后

mysql> show tables;
+----------------+
| Tables_in_test |
+----------------+
| user           |
| user_detail    |
| user_score     |
+----------------+
3 rows in set (0.01 sec)

mysql> select * from `user`;
+----+--------+
| id | name   |
+----+--------+
|  1 | name_1 |
|  2 | name_2 |
|  3 | name_3 |
|  4 | name_4 |
|  5 | name_5 |
+----+--------+
5 rows in set (0.01 sec)

mysql> select * from `user_detail`;
+-----+-----------+
| uid | address   |
+-----+-----------+
|   1 | address_1 |
|   2 | address_2 |
|   3 | address_3 |
|   4 | address_4 |
|   5 | address_5 |
+-----+-----------+
5 rows in set (0.00 sec)

mysql> select * from `user_score`;
+----+-----+-------+
| id | uid | score |
+----+-----+-------+
|  1 |   1 |     1 |
|  2 |   1 |     2 |
|  3 |   1 |     3 |
|  4 |   1 |     4 |
|  5 |   1 |     5 |
|  6 |   2 |     1 |
|  7 |   2 |     2 |
|  8 |   2 |     3 |
|  9 |   2 |     4 |
| 10 |   2 |     5 |
| 11 |   3 |     1 |
| 12 |   3 |     2 |
| 13 |   3 |     3 |
| 14 |   3 |     4 |
| 15 |   3 |     5 |
| 16 |   4 |     1 |
| 17 |   4 |     2 |
| 18 |   4 |     3 |
| 19 |   4 |     4 |
| 20 |   4 |     5 |
| 21 |   5 |     1 |
| 22 |   5 |     2 |
| 23 |   5 |     3 |
| 24 |   5 |     4 |
| 25 |   5 |     5 |
+----+-----+-------+
25 rows in set (0.00 sec)
```

### 5.1、数据查询

新的With特性下，数据查询相当简便，例如，我们查询一条数据：