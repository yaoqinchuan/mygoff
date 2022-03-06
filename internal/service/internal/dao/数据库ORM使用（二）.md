# 数据库ORM使用（二）

## 1、对象输入

​		Data/Where/WherePri/And/Or方法支持任意的string/map/slice/struct/*struct数据类型参数，该特性为gdb提供了很高的灵活性。当使用struct/*struct对象作为输入参数时，将会被自动解析为map类型，只有struct的公开属性能够被转换，并且支持 orm/gconv/json 标签，用于定义转换后的键名，即与表字段的映射关系。

```go
type User struct {
    Uid      int    `orm:"user_id"`
    Name     string `orm:"user_name"`
    NickName string `orm:"nick_name"`
}
// 或者
type User struct {
    Uid      int    `gconv:"user_id"`
    Name     string `gconv:"user_name"`
    NickName string `gconv:"nick_name"`
}
// 或者
type User struct {
    Uid      int    `json:"user_id"`
    Name     string `json:"user_name"`
    NickName string `json:"nick_name"`
}
```

​		其中，struct的属性应该是公开属性（首字母大写），orm标签对应的是数据表的字段名称。表字段的对应关系标签既可以使用orm，也可以用gconv，还可以使用传统的json标签，但是当三种标签都存在时，orm标签的优先级更高。为避免将struct对象转换为JSON数据格式返回时与JSON编码标签冲突，推荐使用orm标签来实现数据库ORM的映射关系。

## 2、字段过滤

### 2.1、Fields字段过滤

假如user表有4个字段uid, nickname, passport, password。

- 查询字段过滤

```go
 // SELECT `uid`,`nickname` FROM `user` ORDER BY `uid` asc
 db.Table("user").Fields("uid, nickname").Order("uid asc").All()
```

- 写入字段过滤

```GO
m := g.Map{
     "uid"      : 10000,
     "nickname" : "John Guo",
     "passport" : "john",
     "password" : "123456",
 }
 db.Table(table).Fields("nickname,passport,password").Data(m).Insert()
 // INSERT INTO `user`(`nickname`,`passport`,`password`) VALUES('John Guo','john','123456')
```

### 2.2、FieldsEx字段过滤

假如user表有4个字段uid, nickname, passport, password。

- 查询字段排除

```go
 // SELECT `uid`,`nickname` FROM `user`
 db.Table("user").FieldsEx("passport, password").All()
```

- 写入字段排除

```go
 m := g.Map{
     "uid"      : 10000,
     "nickname" : "John Guo",
     "passport" : "john",
     "password" : "123456",
 }
 db.Table(table).FieldsEx("uid").Data(m).Insert()
 // INSERT INTO `user`(`nickname`,`passport`,`password`) VALUES('John Guo','john','123456')
```

### 2.3、OmitEmpty空值过滤

​		当 map/struct 中存在空值如 nil,"",0 时，默认情况下，gdb将会将其当做正常的输入参数，因此这些参数也会被更新到数据表。OmitEmpty特性可以在将数据写入到数据库之前过滤空值数据的字段。

```go
func (m *Model) OmitEmpty() *Model
func (m *Model) OmitEmptyWhere() *Model
func (m *Model) OmitEmptyData() *Model 
```

​		OmitEmpty方法会同时过滤Where及Data中的空值数据，而通过OmitEmptyWhere/OmitEmptyData方法可以执行特定的字段过滤。

#### 2.3.1、写入/更新操作

```go
// UPDATE `user` SET `name`='john',update_time=null WHERE `id`=1
db.Table("user").Data(g.Map{
    "name"        : "john",
    "update_time" : nil,
}).Where("id", 1).Update()
```

针对空值情况，我们可以通过OmitEmpty方法来过滤掉这些空值。例如，以上示例可以修改为：

```go
// UPDATE `user` SET `name`='john' WHERE `id`=1
db.Table("user").OmitEmpty().Data(g.Map{
    "name"        : "john",
    "update_time" : nil,
}).Where("id", 1).Update()
```

也可以这样

```go
type User struct {
    Id         int    `orm:"id"`
    Passport   string `orm:"passport"`
    Password   string `orm:"password"`
    NickName   string `orm:"nickname"`
    CreateTime string `orm:"create_time"`
    UpdateTime string `orm:"update_time"`
}
user := User{
    Id        : 1,
    NickName  : "john",
    UpdateTime: gtime.Now().String(),
}
db.Table("user").OmitEmpty().Data(user).Insert()
// INSERT INTO `user`(`id`,`nickname`,`update_time`) VALUES(1,'john','2019-10-01 12:00:00')
```

#### 2.3.2、omitempty标签与OmitEmpty方法

​        针对于struct的空值过滤大家会想到omitempty的标签。该标签常用于json转换的空值过滤，也在某一些第三方的ORM库中用作struct到数据表字段的空值过滤，即当属性为空值时不做转换。
​		omitempty标签与OmitEmpty方法所达到的效果是一样的。在ORM操作中，我们不建议对struct使用omitempty的标签来控制字段的空值过滤，而建议使用OmitEmpty方法来做控制。因为该标签一旦加上之后便绑定到了struct上，没有办法做灵活控制；而通过OmitEmpty方法使得开发者可以选择性地、根据业务场景对struct做空值过滤，操作更加灵活。

#### 2.3.3、数据查询操作

​     空值也会影响数据查询操作，主要是影响where条件参数。我们可以通过OmitEmpty方法过滤条件参数中的空值。

```go
// SELECT * FROM `user` WHERE `passport`='john' LIMIT 1
r, err := db.Table("user").Where(g.Map{
    "nickname" : "",
    "passport" : "john",
}).OmitEmpty().One()
```

```go
type User struct {
    Id         int    `orm:"id"`
    Passport   string `orm:"passport"`
    Password   string `orm:"password"`
    NickName   string `orm:"nickname"`
    CreateTime string `orm:"create_time"`
    UpdateTime string `orm:"update_time"`
}
user := User{
    Passport : "john",
}
r, err := db.Table("user").OmitEmpty().Where(user).One()
// SELECT * FROM `user` WHERE `passport`='john' LIMIT 1
```

#### 2.3.4、OmitNil空值过滤

​		当 map/struct 中存在空值如 nil时，默认情况下，gdb将会将其当做正常的输入参数，因此这些参数也会被更新到数据表。OmitNil特性可以在将数据写入到数据库之前过滤空值数据的字段。与OmitEmpty特性的区别在于，OmitNil只会过滤值为nil的空值字段，其他空值如"",0并不会被过滤。

```go
func (m *Model) OmitNil() *Model
func (m *Model) OmitNilWhere() *Model
func (m *Model) OmitNilData() *Model 
```

​		OmitEmpty方法会同时过滤Where及Data中的空值数据，而通过OmitEmptyWhere/OmitEmptyData方法可以执行特定的字段过滤。

## 3、字段获取

### 3.2、FieldsStr字段获取

FieldsStr 用于获取指定表的字段，并可给定字段前缀，字段之间使用","符号连接成字符串返回；

假如user表有4个字段uid, nickname, passport, password。
**查询字段**

```GO
 // uid,nickname,passport,password
 db.Table("user").FieldsStr()
```

**查询字段给指定前缀**

```GO
 // gf_uid,gf_nickname,gf_passport,gf_password
 db.Table("user").FieldsStr("gf_")
```



### 3.3、FieldsExStr字段获取

FieldsExStr 用于获取指定表中例外的字段，并可给定字段前缀，字段之间使用","符号连接成字符串返回；

```GO
 // uid,nickname
 db.Table("user").FieldsExStr("passport, password")
```

```GO
 // gf_uid,gf_nickname
 db.Table("user").FieldsExStr("passport, password", "gf_")
```



## 4、事务处理

​		Model对象也可以通过TX事务对象创建，通过事务对象创建的Model对象与通过DB数据库对象创建的Model对象功能是一样的，只不过前者的所有操作都是基于事务，而当事务提交或者回滚后，对应的Model对象不能被继续使用，否则会返回错误。因为该TX对象不能被继续使用，一个事务对象仅对应于一个事务流程，Commit/Rollback后即结束。

### 4.1、通过Transaction

```go
func (db DB) Transaction(ctx context.Context, f func(ctx context.Context, tx *TX) error) (err error)
```

​		当给定的闭包方法返回的error为nil时，那么闭包执行结束后当前事务自动执行Commit提交操作；否则自动执行Rollback回滚操作。

```go
func Register() error {
	return db.Transaction(ctx, func(ctx context.Context, tx *gdb.TX) error {
		var (
			result sql.Result
			err    error
		)
		// 写入用户基础数据
		result, err = tx.Table("user").Insert(g.Map{
			"name":  "john",
			"score": 100,
			//...
		})
		if err != nil {
			return err
		}
		// 写入用户详情数据，需要用到上一次写入得到的用户uid
		result, err = tx.Table("user_detail").Insert(g.Map{
			"uid":   result.LastInsertId(),
			"phone": "18010576258",
			//...
		})
		return err
	})
}
```

### 4.2、通过TX链式操作

​		我们也可以在链式操作中通过TX方法切换绑定的事务对象。多次链式操作可以绑定同一个事务对象，在该事务对象中执行对应的链式操作。

```go
func Register() error {
	var (
		uid int64
		err error
	)
	tx, err := g.DB().Begin()
	if err != nil {
		return err
	}
	// 方法退出时检验返回值，
	// 如果结果成功则执行tx.Commit()提交,
	// 否则执行tx.Rollback()回滚操作。
	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()
	// 写入用户基础数据
	uid, err = AddUserInfo(tx, g.Map{
		"name":  "john",
		"score": 100,
		//...
	})
	if err != nil {
		return err
	}
	// 写入用户详情数据，需要用到上一次写入得到的用户uid
	err = AddUserDetail(tx, g.Map{
		"uid":   uid,
		"phone": "18010576259",
		//...
	})
	return err
}

func AddUserInfo(tx *gdb.TX, data g.Map) (int64, error) {
	result, err := g.Table("user").TX(tx).Data(data).Insert()
	if err != nil {
		return 0, err
	}
	uid, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	return uid, nil
}

func AddUserDetail(tx *gdb.TX, data g.Map) error {
	_, err := g.Table("user_detail").TX(tx).Data(data).Insert()
	return err
}
```

## 5、主从切换

​		以下是一个简单的主从配置，包含一主一从,在配置了主从的情况下，如果master挂了则写失败，slave挂了则读失败。

```yaml
database:
  default:

    type: "mysql"
    link: "root:12345678@tcp(192.168.1.1:3306)/test"
    role: "master"

    type: "mysql"
    link: "root:12345678@tcp(192.168.1.2:3306)/test"
    role: "slave"
```

​        在大部分的场景中，我们的写入请求是到Master主节点，而读取请求是到Slave从节点，这样的好处是能够对数据库的请求进行压力分摊，并提高数据库的可用性。但在某些场景中，我们期望读取操作在Master节点上执行，特别是一些对于即时性要求比较高的场景（因为主从节点之间的数据同步是有延迟的）。开发者可以通过Master和Slave方法自定义决定当前链式操作执行在哪个节点上。

​        在订单创建的时候，没有必要指定操作的节点，因为写入操作默认是在主节点上执行的。为简化示例，我们这里仅展示关键的代码：

```go
 db.Model("order").Data(g.Map{
     "uid"   : 1000,
     "price" : 99.99,
     // ...
 }).Insert()
```

在订单列表页面查询时，我们需要使用Master方法指定查询操作是在主节点上进行，以避免读取延迟。

```go
 db.Model("order").Master().Where("uid", 1000).All()
```

## 6、查询缓存

​		gdb支持对查询结果的缓存处理，常用于多读少写的查询缓存场景，并支持手动的缓存清理。需要注意的是，查询缓存仅支持链式操作，且在事务操作下不可用。

```go
type CacheOption struct {
	// Duration is the TTL for the cache.
	// If the parameter `Duration` < 0, which means it clear the cache with given `Name`.
	// If the parameter `Duration` = 0, which means it never expires.
	// If the parameter `Duration` > 0, which means it expires after `Duration`.
	Duration time.Duration

	// Name is an optional unique name for the cache.
	// The Name is used to bind a name to the cache, which means you can later control the cache
	// like changing the `duration` or clearing the cache with specified Name.
	Name string
	
	// Force caches the query result whatever the result is nil or not.
	// It is used to avoid Cache Penetration.
	Force bool

}

// Cache sets the cache feature for the model. It caches the result of the sql, which means
// if there's another same sql request, it just reads and returns the result from cache, it
// but not committed and executed into the database.
//
// Note that, the cache feature is disabled if the model is performing select statement
// on a transaction.
func (m *Model) Cache(option CacheOption) *Model
```

### 6.1、缓存对象

​		ORM对象默认情况下提供了缓存管理对象，该缓存对象类型为*gcache.Cache，也就是说同时也支持*gcache.Cache的所有特性。可以通过GetCache() *gcache.Cache 接口方法获得该缓存对象，并通过返回的对象实现自定义的各种缓存操作，例如：g.DB().GetCache().Keys()。

​		默认情况下ORM的*gcache.Cache缓存对象提供的是单进程内存缓存，虽然性能非常高效，但是只能在单进程内使用。如果服务如果采用多节点部署，多节点之间的缓存可能会产生数据不一致的情况，因此大多数场景下我们都是通过Redis服务器来实现对数据库查询数据的缓存。*gcache.Cache对象采用了适配器设计模式，可以轻松实现从单进程内存缓存切换为分布式的Redis缓存。

```go
CREATE TABLE `user` (
  `uid` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `name` varchar(30) NOT NULL DEFAULT '' COMMENT '昵称',
  `site` varchar(255) NOT NULL DEFAULT '' COMMENT '主页',
  PRIMARY KEY (`uid`)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8;
```



```go
package main

import (
	"time"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gctx"

)

func main() {
	var (
		db  = g.DB()
		ctx = gctx.New()
	)

	// 开启调试模式，以便于记录所有执行的SQL
	db.SetDebug(true)
	
	// 写入测试数据
	_, err := db.Model("user").Ctx(ctx).Data(g.Map{
		"name": "john",
		"site": "https://goframe.org",
	}).Insert()
	
	// 执行2次查询并将查询结果缓存1小时，并可执行缓存名称(可选)
	for i := 0; i < 2; i++ {
		r, _ := db.Model("user").Ctx(ctx).Cache(gdb.CacheOption{
			Duration: time.Hour,
			Name:     "vip-user",
			Force:    false,
		}).Where("uid", 1).One()
		g.Log().Debug(ctx, r.Map())
	}
	
	// 执行更新操作，并清理指定名称的查询缓存
	_, err = db.Model("user").Ctx(ctx).Cache(gdb.CacheOption{
		Duration: -1,
		Name:     "vip-user",
		Force:    false,
	}).Data(gdb.Map{"name": "smith"}).Where("uid", 1).Update()
	if err != nil {
		g.Log().Fatal(ctx, err)
	}
	
	// 再次执行查询，启用查询缓存特性
	r, _ := db.Model("user").Ctx(ctx).Cache(gdb.CacheOption{
		Duration: time.Hour,
		Name:     "vip-user",
		Force:    false,
	}).Where("uid", 1).One()
	g.Log().Debug(ctx, r.Map())

}
```

执行后输出结果为（测试表数据结构仅供示例参考）：

```
2022-02-08 17:36:19.817 [DEBU] {c0424c75f1c5d116d0df0f7197379412} {"name":"john","site":"https://goframe.org","uid":1} 
2022-02-08 17:36:19.817 [DEBU] {c0424c75f1c5d116d0df0f7197379412} {"name":"john","site":"https://goframe.org","uid":1} 
2022-02-08 17:36:19.817 [DEBU] {c0424c75f1c5d116d0df0f7197379412} [  0 ms] [default] [rows:1  ] UPDATE `user` SET `name`='smith' WHERE `uid`=1 
2022-02-08 17:36:19.818 [DEBU] {c0424c75f1c5d116d0df0f7197379412} [  1 ms] [default] [rows:1  ] SELECT * FROM `user` WHERE `uid`=1 LIMIT 1 
2022-02-08 17:36:19.818 [DEBU] {c0424c75f1c5d116d0df0f7197379412} {"name":"smith","site":"https://goframe.org","uid":1}
```

可以看到：

- 为了方便展示缓存效果，这里开启了数据debug特性，当有任何的SQL操作时将会输出到终端。
- 执行两次One方法数据查询，第一次走了SQL查询，第二次直接使用到了缓存，SQL没有提交到数据库执行，因此这里只打印了一条查询SQL，并且两次查询的结果也是一致的。
- 注意这里为该查询的缓存设置了一个自定义的名称vip-user，以便于后续清空更新缓存。如果缓存不需要清理，那么可以不用设置缓存名称。
- 当执行Update更新操作时，同时根据名称清空指定的缓存。
- 随后再执行One方法数据查询，这时重新缓存新的数据。

## 7、数据库切换

​		我们知道数据库的配置中有支持对默认数据库的配置，因此DB对象及Model对象在初始化的时候已经绑定到了特定的数据库上。运行时切换数据库有几种方案（假如我们的数据库有user用户数据库和order订单数据库）：

- 通过不同的配置分组来实现。这需要在配置文件中配置不同的分组配置，随后在程序中可以通过g.DB("分组名称")来获取特定数据库的单例对象。
- 通过运行时DB.SetSchema方法切换单例对象的数据库，需要注意的是由于修改的是单例对象的数据库配置，因此影响是全局的：

```go
 g.DB().SetSchema("user-schema")
 g.DB().SetSchema("order-schema")
```

- 通过链式操作Schema方法创建Schema数据库对象，并通过该数据库对象创建模型对象并执行后续链式操作：

```go
db.Schema("user-schema").Model("user").All()
db.Schema("order-schema").Model("order").All()
```

- 此外，假如当前数据库操作配置的用户有权限，那么可以直接通过表名中带数据库名称实现跨域操作，甚至跨域关联查询：

```go
 // SELECT * FROM `order`.`order` o LEFT JOIN `user`.`user` u ON (o.uid=u.id) WHERE u.id=1 LIMIT 1
 db.Model("order.order o").LeftJoin("user.user u", "o.uid=u.id").Where("u.id", 1).One()
```

## 8、Handler特性

Handler特性允许您轻松地复用常见的逻辑。

### 8.1、查询

```go
func AmountGreaterThan1000(m *gdb.Model) *gdb.Model {
	return m.WhereGT("amount", 1000)
}

func PaidWithCreditCard(m *gdb.Model) *gdb.Model {
	return m.Where("pay_mode_sign", "credit_card")
}

func PaidWithCod(m *gdb.Model) *gdb.Model {
	return m.Where("pay_mode_sign", "cod")
}

func OrderStatus(statuses []string) func(m *gdb.Model) *gdb.Model {
	return func(m *gdb.Model) *gdb.Model {
		return m.Where("status", statuses)
	}
}

var (
	m = g.Model("product_order")
)

m.Handler(AmountGreaterThan1000, PaidWithCreditCard).Scan(&orders)
// SELECT * FROM `product_order` WHERE `amount`>1000 AND `pay_mode_sign`='credit_card'
// 查找所有金额大于 1000 的信用卡订单

m.Handler(AmountGreaterThan1000, PaidWithCod).Scan(&orders)
// SELECT * FROM `product_order` WHERE `amount`>1000 AND `pay_mode_sign`='cod'
// 查找所有金额大于 1000 的 COD 订单

m.Handler(AmountGreaterThan1000, OrderStatus([]string{"paid", "shipped"})).Scan(&orders)
// SELECT * FROM `product_order` WHERE `amount`>1000 AND `status` IN('paid','shipped')
// 查找所有金额大于1000 的已付款或已发货订单
```

### 8.2、分页

```go
func Paginate(r *ghttp.Request) func(m *gdb.Model) *gdb.Model {
	return func(m *gdb.Model) *gdb.Model {
		type Pagination struct {
			Page int
			Size int
		}
		var pagination Pagination
		_ = r.Parse(&pagination)
		switch {
		case pagination.Size > 100:
			pagination.Size = 100

		case pagination.Size <= 0:
			pagination.Size = 10
		}
		return m.Page(pagination.Page, pagination.Size)
	}

}

m.Handler(Paginate(r)).Scan(&users)
m.Handler(Paginate(r)).Scan(&articles)
```

## 9、悲观锁 & 乐观锁

- 悲观锁（Pessimistic Lock），顾名思义，就是很悲观，每次去拿数据的时候都认为别人会修改，所以每次在拿数据的时候都会上锁，这样别人想拿这个数据就会阻塞直到它拿到锁。传统的关系型数据库里边就用到了很多这种锁机制，比如行锁、表锁、读锁、写锁等，都是在做操作之前先上锁。


- 乐观锁（Optimistic Lock），顾名思义，就是很乐观，每次去拿数据的时候都认为别人不会修改，所以不会上锁，但是在更新的时候会判断一下在此期间别人有没有去更新这个数据，可以使用版本号等机制实现。乐观锁适用于多读的应用类型，这样可以提高吞吐量。

### 9.1、悲观锁使用

```go
func (m *Model) LockUpdate() *Model
func (m *Model) LockShared() *Model
```

​		gdb模块的链式操作提供了两个方法在SQL语句中实现“悲观锁”。可以在查询中使用LockShared方法从而在运行语句时带一把”共享锁“。共享锁可以避免被选择的行被修改直到事务提交：

```go
db.Model("users").Ctx(ctx).Where("votes>?", 100).LockShared().All();
```

上面这个查询等价于下面这条 SQL 语句：

```sql
SELECT * FROM `users` WHERE `votes` > 100 LOCK IN SHARE MODE
```

此外还可以使用LockUpdate方法。该方法用于创建FOR UPDATE锁，避免选择行被其它共享锁修改或删除：

```go
db.Model("users").Ctx(ctx).Where("votes>?", 100).LockUpdate().All();
```

上面这个查询等价于下面这条 SQL 语句：

```go
SELECT * FROM `users` WHERE `votes` > 100 FOR UPDATE
```

​        FOR UPDATE 与 LOCK IN SHARE MODE 都是用于确保被选中的记录值不能被其它事务更新（上锁），两者的区别在于 LOCK IN SHARE MODE 不会阻塞其它事务读取被锁定行记录的值，而 FOR UPDATE会阻塞其他锁定性读对锁定行的读取（非锁定性读仍然可以读取这些记录，LOCK IN SHARE MODE 和 FOR UPDATE都是锁定性读）。

### 9.2、乐观锁使用

​		乐观锁，大多是基于数据版本 （ Version ）记录机制实现。何谓数据版本？即为数据增加一个版本标识，在基于数据库表的版本解决方案中，一般是通过为数据库表增加一个 "version" 字段来实现。

​		读取出数据时，将此版本号一同读出，之后更新时，对此版本号加一。此时，将提交数据的版本数据与数据库表对应记录的当前版本信息进行比对，如果提交的数据版本号大于数据库表当前版本号，则予以更新，否则认为是过期数据。

### 9.3、锁机制总结

​		两种锁各有优缺点，不可认为一种好于另一种，像乐观锁适用于写比较少的情况下，即冲突真的很少发生的时候，这样可以省去了锁的开销，加大了系统的整个吞吐量。但如果经常产生冲突，上层应用会不断的进行重试，这样反倒是降低了性能，所以这种情况下用悲观锁就比较合适。