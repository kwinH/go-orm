# 简介

> `oorm`是一款功能全面的数据库操作工具，提供一个漂亮、简洁的链式调用方式来实现与数据库的交互。


# 模型定义
```go
	type User struct {
        oorm.Model
        Name      string `db:"index:us|1"`
        Password  string
        Status    int8
        Age       int8
        Sex       int8
        Balance   float64 `db:"decimal:10,2"`
}
```

# 连接数据库
## mysql
```go
    dsn := "user:pass@tcp(127.0.0.1:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local"
	baseDB, err = oorm.Open(mysql.Open(dsn), &oorm.Config{})

	if err != nil {
		fmt.Printf("%#v", err.Error())
		panic("连接数据库失败")
	}
```

## 连接池
OORM 使用 database/sql 维护连接池
```go
    sqlDB, err := baseDB.DB()
    
    // SetMaxIdleConns 设置空闲连接池中连接的最大数量
    sqlDB.SetMaxIdleConns(10)
    
    // SetMaxOpenConns 设置打开数据库连接的最大数量。
    sqlDB.SetMaxOpenConns(100)
    
    // SetConnMaxLifetime 设置了连接可复用的最大时间。
    sqlDB.SetConnMaxLifetime(time.Hour)
```



# 约定
`oorm` 倾向于约定优于配置 默认情况下，`oorm` 使用 ID 作为主键，使用结构体名的 `蛇形` 作为表名，字段名的 `蛇形` 作为列名，并使用 `CreatedAt`、`UpdatedAt` 字段追踪创建、更新时间

### oorm.Model
`oorm` 定义一个 `oorm.Model` 结构体，其包括字段 `ID`、`CreatedAt`、`UpdatedAt`、`DeletedAt`
> 您可以将它嵌入到您的结构体中，以包含这几个字段
```go
    type Model struct {
        Id        uint `db:"autoIncrement"`
        CreatedAt time.Time
        UpdatedAt time.Time
        DeletedAt sql.NullTime `db:"index"`
    }
```

### 创建/更新时间追踪（微秒、毫秒、秒、Time）
> GORM 约定使用 `CreatedAt`、`UpdatedAt` 追踪创建/更新时间。如果您定义了这种字段，GORM 在创建、更新时会自动填充 当前时间

如果您想要保存 UNIX（毫/微）秒时间戳，而不是 time，您只需简单地将 time.Time 修改为 `整数类型` 即可，以下即对应关系

>int/uint/int32/uint32 类型对应`秒`
> <br/>
>int64 类型对应`毫秒`
> <br/>
>int64 类型对应`微秒`

```go
type User struct {
  CreatedAt int  // 在创建时，如果该字段值为零值，则使用当前时间戳秒填充
  UpdatedAt int64       // 在创建时该字段值为零值或者在更新时，使用当前时间戳毫秒数填充
}
```

### 字段标签
> 多个标签用`;`分割，例如：`db:"decimal:10,2;index:us|1"`

| 标签名           | 说明                            |
|---------------|-------------------------------|
| field         | 指定 db 列名                      |
| size	         | 定义列数据类型的大小或长度，例如 size: 256    |
| default       | 定义列的默认值  (可以定义为NULL)          |
| primaryKey    | 将列定义为主键                       |
| autoIncrement | 指定列为自动增长(同是也会指定为主键)           |
| decimal       | 精度 例 `db:"decimal:10,2"`      |
| comment       | 	迁移时为字段添加注释                   |
| raw           | 原生表达式  例 `db:"raw:count(*)"`  |
| json          | 用于自动解析装载json                  |
| index         | 根据参数创建普通索引，多个字段使用相同的名称则创建复合索引 |
| unique        | 根据参数创建唯一索引，多个字段使用相同的名称则创建复合索引 |
| full          | 根据参数创建全文索引，多个字段使用相同的名称则创建复合索引 |


# 迁移

## Auto
AutoMigrate 用于自动迁移您的 schema，保持您的 schema 是最新的。

>注意： Auto 会创建表、缺失的外键、列和索引。 出于保护您数据的目的，它可以传第二个参数选择是否更新、传第三个参数选择是否删除（包括字段、属性、注释、索引）。

> 复合索引需要写成：`index:索引名|优先级` ，优先级以小到大排序
```go

	type User struct {
		oorm.Model
		Name      string `db:"index:us|1"`
		Password  string
		Status    int8
		Age       int8
		Sex       int8
		Balance   float64 `db:"decimal:10,2"`
	}


	baseDB.Migrate.Auto(User{}, true, true)


```

## Migrator 接口
```go
type IMigrator interface {
	TableExist(tableName string) bool
	Create(schema1 *Schema) error
	TableInfo(tableName string) TableInfo

	AddField(TableName string, field *Field) error
	ModifyField(TableName string, field *Field) error
	DropField(TableName string, FiledName string) error

	AddIndex(tableName string, indexType IndexType, indexFields IndexList) error
	DropIndex(indexKey, tableName string) error
	UpdateIndex(schema1 *Schema, schemaKeys IndexList, keys map[string][]string, modify bool, indexType IndexType) (err error)

	Auto(value interface{}, modify, drop bool) error
}
```

# 查询构造器
## 声明
```go
	type User struct {
        oorm.Model
        Name      string `db:"index:us|1"`
        Password  string
        Status    int8
        Age       int8
        Sex       int8
        Balance   float64 `db:"decimal:10,2"`
}

    var users []User
	var user User
```

## 获取所有结果
> 你可以在查询上链式调用更多的约束，最后使用 `Get` 方法获取结果

```go
baseDB.Get(&users)
```

## 查询单个结果

```go
baseDB.Where("user_name","kwinwong").First(&user)
```

## 根据主键检索

```go
baseDB.Where("user_name","kwinwong").Find(&user,1)
```

## 聚合查询
> 查询构造器还提供了各种聚合方法，比如 `count`，`max`，`min`，`avg`，还有 `sum`。你可以在构造查询后调用任何方法：
```go
// SELECT max(`id`)  FROM `user` []
res,err := baseDB.Table("user").Max("id")

// SELECT min(`id`) FROM `user` []
res,err := baseDB.Table("user").Mix("id")

// SELECT count(*) as `c` FROM `user` []
res,err := baseDB.Table("user").Count()
```

> 当`Group`分组查询的时候，会被当做子查询
```go
// SELECT COUNT(*) FROM (SELECT `id` FROM `user` GROUP BY `status`) as `tmp1`
res,err := baseDB.Table("user").Group("status").Select("id").Count()
```

> 如果需要分组查询：
```go

    type User struct {
    Status int8
    C      int `db:"raw:count(*)"`
    }
    
    var users []User
    // SELECT `status`,count(*) FROM `user` GROUP BY `status`
    baseDB.Group("status").Get(&users)
```

## Select

> 查询字段 默认查询所有字段
>
> 支持多个参数，单个参数，切片方式传值

```go

// SELECT `id`,`name` as `n` FROM `user` []
err := baseDB.Select("id", "name as n").Get(&users)

err := baseDB.Select("id,name n").Get(&users)

err := baseDB.Select([]string{"id", "name n"}).Get(&users)
```

## Omit 
> 忽略字段
```go
// SELECT `id`,`created_at`,`updated_at`,`deleted_at`,`name`,`password`,`status`,`age`,`sex`,`balance` FROM `user` WHERE `user`.`deleted_at` IS NULL []
err := baseDB.Omit("age", "sex").Get(&users)
```

## 原生表达式

> 有时候你可能需要在查询中使用原生表达式。你可以使用 `sqlBuilder.Raw` 创建一个原生表达式：

### 原生查询
```go
	type User struct {
		Sex int8
		C      int
	}

	var u []User
	err := baseDB.Raw("SELECT sex,count(*) c from `user` group by sex having c>?", 100).Get(&u)
```

### 原生执行
```go
res, err := baseDB.Exec("UPDATE `user` SET `status`=1 WHERE id=?", 1)
```

### 原生字段

```go
// SELECT DISTINCT mobile FROM `user` []
err := baseDB.Select(sqlBuilder.Raw("DISTINCT mobile")).Get(&users)
```

### 原生条件

```go 
// SELECT * FROM `user` WHERE price > IF(state = 'TX', 200, 100) []
err := baseDB.Where(sqlBuilder.Raw("price > IF(state = 'TX', 200, 100)")).Get(&users)
```

## Table

> 指定查询表名

```go
//SELECT * FROM `users`
baseDB.Table("users").Get(&users)
```

### 子查询

```go
// SELECT * FROM (SELECT * FROM (SELECT `sex`,count(*) as `c` FROM m_users GROUP BY `sex`) as `tmp2`) as `tmp1` []
err := baseDB.Table(func (m *sqlBuilder.Builder) {
m.Table(func (m *sqlBuilder.Builder) {
m.Table("m_users").Select("sex", "count(*) as c").Group("sex")
}).Get(&users)
```

## Where

### 简单where语句

> 在构造 where 查询实例中，你可以使用 where 方法。调用 where 最基本的方式是需要传递三个参数：第一个参数是列名，第二个参数是任意一个数据库系统支持的运算符，第三个是该列要比较的值。

```go
// SELECT `id`,`name` FROM `users` WHERE  `id` < ?  [100]
err := baseDB.Where("id", "<", 100).Get(&users)
```

> 为了方便，如果你只是简单比较列值和给定数值是否相等，可以将数值直接作为 where 方法的第二个参数：

```go
// SELECT `id`,`name` FROM `users` WHERE  `id` = ? [1]
err := baseDB.Where("id", 1).Get(&users)
```

### OrWhere语句

> orWhere 方法和 where 方法接收的参数一样：

```go
// SELECT * FROM `user` WHERE  `id` = ? OR  `name` like ? [1 %q%]
err := baseDB.Where("id", 1).OrWhere("name", "like", "%q%").Get(&users)
```

### WhereBetween / WhereNotIn / WhereNotBetween / OrWhereNotBetween

> `WhereBetween` 方法验证字段值是否在给定的两个值之间：

> 可以传一个数组，也可以传2个值

```go
// SELECT * FROM `user` WHERE `sex` = ? AND `attribute` BETWEEN ? AND ? [1 2 3]
err := baseDB.Where("sex", 1).WhereBetween("attribute", 2, 3).Get(&users)

// SELECT * FROM `user` WHERE `sex` = ? OR `attribute` BETWEEN ? AND ? [1 2 3]
err := baseDB.Where("sex", 1).OrWhereBetween("attribute", []int{2, 3}).Get(&users)
```

> `WhereNotBetween` 方法用于验证字段值是否在给定的两个值之外：

> 可以传一个数组，也可以传2个值

```go
// SELECT * FROM `user` WHERE `sex` = ? AND `attribute` NOT BETWEEN ? AND ? [1 2 3]
err := baseDB.Where("sex", 1).WhereNotBetween("attribute", 2, 3).Get(&users)

// SELECT * FROM `user` WHERE `sex` = ? OR `attribute` NOT BETWEEN ? AND ? [1 2 3]
err := baseDB.Where("sex", 1).OrWhereNotBetween("attribute", []int{2, 3}).Get(&users)
```

### WhereIn / WhereNotIn / OrWhereIn / OrWhereNotIn

> `WhereIn` 方法验证给定列的值是否包含在给定数组中：

> 可以传一个数组，也可以传多个值

```go
// SELECT * FROM `user` WHERE `sex` = ? AND `id` IN (?,?) [1 100 200]
err := baseDB.Where("sex", 1).WhereIn("id", 100, 200).Get(&users)

// SELECT * FROM `user` WHERE `sex` = ? OR `id` IN (?,?) [1 100 200]
err := baseDB.Where("sex", 1).OrWhereIn("id", []int{100, 200}).Get(&users)
```

> `WhereNotIn` 方法验证给定列的值是否`不存在`给定的数组中：

```go
// SELECT * FROM `user` WHERE `sex` = ? AND `id` NOT IN (?,?) [1 100 200]
err := baseDB.Where("sex", 1).WhereNotIn("id", []int{100, 200}).Get(&users)

// SELECT * FROM `user` WHERE `sex` = ? OR `id` NOT IN (?,?) [1 100 200]
err := baseDB.Where("sex", 1).OrWhereNotIn("id", []int{100, 200}).Get(&users)
```

### ＷhereNull / ＷhereNotNull / ＯrWhereNull / ＯrWhereNotNull

＞`ＷhereNull` 方法验证指定的字段`必须是 NULL`:

```go
// SELECT * FROM `user` WHERE `sex` = ? AND `deleted_at` IS NULL [1]
err := baseDB.Where("sex", 1).WhereNull("deleted_at").Get(&users)

// SELECT * FROM `user` WHERE `sex` = ? OR `deleted_at` IS NULL [1]
err := baseDB.Where("sex", 1).OrWhereNull("deleted_at").Get(&users)

```

> `WhereNotNull` 方法验证指定的字段`肯定不是 NULL`:

```go
// SELECT * FROM `user` WHERE `sex` = ? AND `deleted_at` IS NOT NULL [1]
err := baseDB.Where("sex", 1).WhereNotNull("deleted_at").Get(&users)

// SELECT * FROM `user` WHERE `sex` = ? OR `deleted_at` IS NOT NULL [1]
err := baseDB.Where("sex", 1).OrWhereNotNull("deleted_at").Get(&users)

```

### 分组查询

> 如果需要在括号内对 or 条件进行分组，将闭包作为 orWhere 方法的第一个参数也是可以的：

```go
// SELECT `id` FROM `user` WHERE  `id` <> ? OR  ( `age` > ? AND  `name` like ?) [1 18 %q%]
sql, bindings := baseDB.Where("id", "<>", 1).OrWhere(func (m *sqlBuilder.Builder) {
m.Where("age", ">", 18).
Where("name", "like", "%q%")
}).Get(&users)
```

### 子查询 Where 语句

```go
// SELECT * FROM `user` WHERE  `id` <> ? AND  `id` in (SELECT `id` FROM `user_old` WHERE  `age` > ? AND  `name` like ?) [1 18 %q%]
err := baseDB.Where("id", "<>", 1).WhereIn("id", func (m *sqlBuilder.Builder) {
m.Select("id").
Table("user_old").
Where("age", ">", 18).
Where("name", "like", "%q%")
}).Get(&users)
```

## Order

> `Order`方法允许你通过给定字段对结果集进行排序。 `order`
> 的第一个参数应该是你希望排序的字段，第二个参数控制排序的方向，可以是 `asc` 或 `desc`,也可以省略，默认是`desc`

```go
// SELECT `id`,`name` FROM `user` ORDER BY `id` DESC []
err := baseDB.Select("id", "name").Order("id", "desc").Get(&users)
```

> 如果你需要使用多个字段进行排序，你可以多次调用 `Order`

```go
// SELECT `id`,`name` FROM `user` ORDER BY `id` DESC,`age` ASC []
err := baseDB.Select("id", "name").Order("id").Order("age", "asc").Get(&users)

```

## groupBy / Having

> groupBy 和 having 方法用于将结果分组。 having 方法的使用与 where 方法十分相似：

```go
// SELECT `age`,count( * ) as `c` FROM `user` GROUP BY `age` HAVING  `c` > ? [10]
err := baseDB.Select("age", "count(*) as c").Group("age").Having("c", ">", 10).Get(&users)

// SELECT `age`,`sex`,count( * ) as `c` FROM `user` GROUP BY `age`,`sex` HAVING  `c` > ? [10]
err := baseDB.Select("age", 'sex', "count(*) as c").Group("age", "sex").Having("c", ">", 10).Get(&users)
```

## Limit

```go
// SELECT `id`,`name` FROM `user` LIMIT 10 []
err := baseDB.Select("id", "name").Limit(10).Get(&users)

// SELECT `id`,`name` FROM `user` LIMIT 1,10 []
err := baseDB.Select("id", "name").Limit(1, 10).Get(&users)

```

## Page

```go
//SELECT `id`,`name` FROM `user` LIMIT 0,10 []
err := baseDB.Select("id", "name").Page(1, 10).Get(&users)
```

## Joins

### Inner Join 语句

> 查询构造器也可以编写 join 方法。若要执行基本的「内链接」，你可以在查询构造器实例上使用 Join 方法。传递给 Join
> 方法的第一个参数是你需要连接的表的名称，第二个参数是指定连接的字段约束，而其他的则是绑定参数。你还可以在单个查询中连接多个数据表：

```go
// SELECT `id`,`name` FROM `user` INNER JOIN `order` as `o` o.user_id=u.user_id and o.type=? INNER JOIN `contacts` as `c` c.user_id=u.user_id [1]
err := baseDB.Table("user u").Select("id", "name").
Join("order o", "o.user_id=u.user_id and o.type=?", 1).
Join("contacts c", "c.user_id=u.user_id").
Get(&users)
```

### Left Join / Right Join 语句

> 如果你想使用 「左连接」或者 「右连接」代替「内连接」 ，可以使用 ＬeftJoin 或者 ＲightJoin 方法。这两个方法与 Join 方法用法相同：

```go
// SELECT `id`,`name` FROM `user` RIGHT JOIN `contacts` as `c` c.user_id=u.user_id []
err := baseDB.Table("user u").Select("id", "name").
LeftJoin("contacts c", "c.user_id=u.user_id").
Get(&users)

// SELECT `id`,`name` FROM `user` LEFT JOIN `contacts` as `c` c.user_id=u.user_id []
err := baseDB.Table("user u").Select("id", "name").
RightJoin("contacts c", "c.user_id=u.user_id").
Get(&users)
```

### 关联子查询

```go
// SELECT `id`,`name` FROM `user` as `u` INNER JOIN (SELECT * FROM `contacts` WHERE `id` > ?) as `tmp1` tmp1.user_id=u.user_id [100]
err := baseDB.Table("user u").Select("id", "name").
Join(func(b *sqlBuilder.Builder) {
b.Table("contacts").Where("id", ">", 100)
}, "tmp1.user_id=u.user_id").
Get(&users)
```

# 模型关联
> 可以使用 `with` 方法指定想要预加载的关联

## 一对一

> 一对一是最基本的关联关系。例如，一个 `User` 模型可能关联一个 `Contact` 模型。为了定义这个关联，我们要在 User 模型中定义一个 `Contact` 模型。

### 声明
```go
type Contact struct{
	oorm.Model
	UserId uint
	Mobile string
	Email string
}

type User struct {
	oorm.Model
	UserName string
	Password string
	Nickname string
	Status   string
	Avatar  string
	Contact  Contact
}
```
### 检索
```go
func GetUser(db *oorm.DB) (*User, error) {
	var user = &User{}

// SELECT `id`,`created_at`,`updated_at`,`deleted_at`,`user_name`,`password`,`nickname`,`status`,`avatar` FROM `user` WHERE `id` = "1" LIMIT 1
// SELECT `id`,`created_at`,`updated_at`,`deleted_at`,`user_id`,`uid`,`mobile`,`email` FROM `contact` WHERE `user_id` in ("1")
	err := db.With("Contact").Find(user, 1)

	if err != nil {
		return nil, err
	}
	return user, err
}
```
> oorm 会基于模型名决定外键名称。在这种情况下，会自动假设 `Contact` 模型有一个 `UserId` 外键。如果你想覆盖这个约定，标签 foreignKey 来更改它：

```go
type User struct {
    oorm.Model
    UserName string
    Password string
    Nickname string
    Status   string
    Avatar   string
    Contact  Contact `db:"foreignKey:Uid"`
}

type Contact struct {
    oorm.Model
    Uid    uint
    Mobile string
    Email  string
}
```

```go
func GetUser(db *oorm.DB) (*User, error) {
	var user = &User{}

// SELECT `id`,`created_at`,`updated_at`,`deleted_at`,`user_name`,`password`,`nickname`,`status`,`avatar` FROM `user` WHERE `id` = "1" LIMIT 1
// SELECT `id`,`created_at`,`updated_at`,`deleted_at`,`user_id`,`uid`,`mobile`,`email` FROM `contact` WHERE `uid` in ("1")
	err := db.With("Contact").Find(user, 1)

	if err != nil {
		return nil, err
	}
	return user, err
}
```

## 一对一（反向）
> 我们已经能从 `User` 模型访问到 `Contact` 模型了。现在，让我们再在 `Contact` 模型上定义一个关联，这个关联能让我们访问到拥有该`Contact`的 `User` 模型

> 以下声明表示拿`Contact`模型的`UserId`字段去查询`User`模型下的`Id`字段
### 声明
```go
type Contact struct {
    oorm.Model
    UserId uint
    Mobile string
    Email  string
    User   User `db:"localKey:UserId;foreignKey:Id"`
}

type User struct {
    oorm.Model
    UserName string
    Password string
    Nickname string
    Status   string
    Avatar   string
}
```

### 检索
```go
func GetUser(db *oorm.DB) (*Contact, error) {
    var contact = &Contact{}

// SELECT `id`,`created_at`,`updated_at`,`deleted_at`,`user_id`,`mobile`,`email` FROM `contact` WHERE `id` = "2" LIMIT 1
// SELECT `id`,`created_at`,`updated_at`,`deleted_at`,`user_name`,`password`,`nickname`,`status`,`avatar` FROM `user` WHERE `id` in ("1")
	err := db.With("Contact").Find(contact, 2)

	if err != nil {
		return nil, err
	}
	return contact, err
}
```

## 插入 & 更新关联模型

> 用`With`指定需要更新的关联模型
> <br/><br/>
> 如果不在事务中，则系统会创建一个事务,统一提交

### 声明

```go
	type Contact struct {
		oorm.Model
		UserId uint
		Mobile string
	}

	type User struct {
		oorm.Model
		UserName string
		Contact  []Contact
	}

```
### 新增
```go
func CreatreUser(db *oorm.DB) (*User, error) {
	var user = &User{
		UserName: "kwinwong",
		Contact: []Contact{{
			Mobile: "13758665977",
		}, {
			Mobile: "13589217699",
		},
		},
	}

// INSERT INTO `user` (`user_name`,`created_at`,`updated_at`) VALUES("kwinwong","2022-10-24 17:25:47.806","2022-10-24 17:25:47.806")
// INSERT INTO `contact` (`user_id`,`mobile`,`created_at`,`updated_at`) VALUES("1","13758665977","2022-10-24 17:25:47.811","2022-10-24 17:25:47.811"),("1","13589217699","2022-10-24 17:25:47.811","2022-10-24 17:25:47.811")		
    res, err := db.With("Contact").Create(user)
    if err != nil {
        return nil, err
    }
	
    return user, err
}
```


# 新增

> 查询构造器还提供了 `Create` 方法用于插新增记录到数据库中。

## 创建Model
```go
	user := User{
		UserName: "kwin",
		Status:   1,
	}

	res, err := baseDB.Create(&user)
```

## 用指定的字段创建记录

### 创建传递的选定字段
```go
	res, err := baseDB.Select("user_name,status").Create(&users)
```

### 创建忽略传递的选定字段
```go
	res, err := baseDB.Omit("user_name,status").Create(&users)
```

### 创建忽略传递的空值字段
```go
	res, err := baseDB.OmitEmpty().Create(&users)
```

## 批量插入
>可以传递给`Create`一个`slice`,或则传递多个`Model`
 
```go
	users := []User{
		{
			UserName: "kwin",
		},
		{
			UserName: "kwin2",
		},
	}

	_, err := baseDB.Create(&users)
``` 

```go
	user1 := User{
		UserName: "kwin",
		Status:   1,
	}

    user2 := User{
        UserName: "kwin2",
        Status:   1,
    }
	_, err := baseDB.Create(&user1,&user2)
```

## 根据Map创建
```go
// INSERT INTO `user` (`name`,`age`) VALUES(?,?) [张三 18]
sql, bindings, err = baseDB.Create(map[string]interface{}{
"name": "张三",
"age":  18,
})
```


> 你甚至可以传递多个map给 `Create` 方法，依次将多个记录插入到表中：

> 注意：多个map参数要一致，以第一个为准，否则会省略后面不一致的map

```go
// INSERT INTO `user` (`name`,`age`) VALUES(?,?),(?,?) [张三 18 李四 30]
sql, bindings, err = baseDB.Create(map[string]interface{}{
"name": "张三",
"age":  18,
}, map[string]interface{}{
"name": "李四",
"age":  30,
})
```

# 更新

## 根据主键更新
> 当没有其他条件时，以主键为条件,当指定了其他条件，则主键不参与条件
```go
	user := User{
		Model: Model{
			Id: 2,
		},
		UserName: "kwin",
		Password: "123456",
		Nickname: "kwin",
		Status:   1,
	}
    
	// UPDATE `user` SET `password`="123456",`nickname`="kwin",`status`=1,`avatar`="",`updated_at`="2022-10-28 11:17:53.303",`user_name`="kwin" WHERE `id` = 2
	res, err := baseDB.Update(&user)
```

## 根据条件更新
```go
	user := User{
		UserName: "kwin",
		Password: "123456",
		Nickname: "kwin",
		Status:   1,
	}
    
	// UPDATE `user` SET `password`="123456",`nickname`="kwin",`status`=1,`avatar`="",`updated_at`="2022-10-28 11:17:53.303",`user_name`="kwin" WHERE `user_name` = "kwin"
	res, err := baseDB.Where("user_name","kwin").Update(&user)
```

## 用指定的字段跟新记录

### 更新传递的选定字段
```go
	res, err := baseDB.Select("user_name,status").Update(&user)
```

### 更新忽略传递的选定字段
```go
	res, err := baseDB.Omit("user_name,status").Update(&user)
```

### 更新忽略传递的空值字段
```go
	res, err := baseDB.OmitEmpty().Update(&user)
```


## 根据Map更新
```go
// UPDATE `user` SET `name`=?,`age`=? WHERE `id` = ? [test 18 1]
err := baseDB.Table("user").Where("id", 1).Update(map[string]interface{}{
"name": "test",
"age":  18,
})
```

# 删除
## 根据主键记录

```go
// user.Id=1 不带条件 则主键不能为空
// UPDATE `user` SET `deleted_at`=? WHERE `id` = ? ["2022-10-26 17:19:04.981",1]
affected,err := baseDB.Delete(&user)
```
## 条件删除
> 当带有其他条件，则主键不引用
```go
// UPDATE `user` SET `deleted_at`=? WHERE `id` <> ? ["2022-10-26 17:19:04.981",100]
affected,err := baseDB.Where("id","<",100).Delete(&user)

```

## 软删除
>如果您的模型包含了一个 `oorm.DeletedAt` 字段（`oorm.Model` 已经包含了该字段)，它将自动获得软删除的能力！

>拥有软删除能力的模型调用 Delete 时，记录不会从数据库中被真正删除。但 `oorm` 会将 DeletedAt 置为当前时间， 并且你不能再通过普通的查询方法找到该记录。

### 查找被软删除的记录
> 您可以使用 `WithDelete` 找到被软删除的记录
```go

baseDB.WithDelete().Get(&users)
```

## 强制删除
> 当调用`Delete`方法传了第二个参数为`true`时，则强制删除

```go
// delete from `user` WHERE `id` < ? [100]
affected,err := baseDB.Where("id","<",100).Delete(&user,true)
```

# 数据库事务
想要在数据库事务中运行一系列操作，你可以使用 `oorm` 的 `Transaction` 方法。如果在事务的闭包中出现了异常，事务将会自动回滚。如果闭包执行成功，事务将会自动提交。在使用 `Transaction` 方法时不需要手动回滚或提交：

```go
	baseDB.Transaction(func(db *DB) (err error) {
        user := User{
        UserName: "test",
        }
        
        _, err = d.Select("user_name").Create(&user)
        
        if err != nil {
            return err
        }
        
        _, err = baseDB.Where("user_name", "test").Delete(&user)
        
        if err != nil {
            return err
        }
        
        return nil
	})
```

## 手动执行事务

>如果你想要手动处理事务并完全控制回滚和提交，可以使用 `oorm` 提供的 `Begin` 方法： 

>在事务中所有的操作都要用`Begin`返回的`*oorm.DB`对象去操作，包括回滚和提交

```go
db, err := d.Begin()
```
你可以通过 `RollBack` 方法回滚事务：

```go
 db.Rollback()
```

最后，你可以通过 `Commit` 方法提交事务： 

```go
db.Commit()
```

# 钩子

## 访问器 / 修改器

### 定义一个访问器
> 若要定义一个访问器， 则需在模型上创建一个 `GetAttr` 方法
```go
var _ IGetAttr = (*User)(nil)

func (u *User) GetAttr() {
	u.UserName = strings.ToUpper(u.UserName)
}
```

### 定义一个修改器
> 若要定义一个访问器， 则需在模型上创建一个 `GetAttr` 方法
```go
var _ ISetAttr = (*User)(nil)

func (u *User) SetAttr() {
    if u.Password != "" {
        u.Password = fmt.Sprintf("%x", md5.Sum([]byte(u.Password)))
    }
}
```

### JSON自动转化
> 当标签中有 `json`,则系统会自动解析、装载
```go
type User struct {
	oorm.Model
	Extend   map[string]interface{} `db:"json"`
}	
```
## 查询模型
```go
// IBeforeQuery 查询前钩子
type IBeforeQuery interface {
	BeforeQuery(*DB) error
}

// IAfterQuery 查询后钩子
type IAfterQuery interface {
	AfterQuery(*DB) error
}
```

## 创建模型
```go
// IBeforeCreate 创建前钩子
type IBeforeCreate interface {
	BeforeCreate(*DB) error
}

// IAfterCreate 查询后钩子
type IAfterCreate interface {
	AfterCreate(*DB) error
}

```

## 修改模型
```go
// IBeforeUpdate 修改前钩子
type IBeforeUpdate interface {
	BeforeUpdate(*DB) error
}

// IAfterUpdate 修改后钩子
type IAfterUpdate interface {
	AfterUpdate(*DB) error
}
```

## 删除模型
```go
// IBeforeDelete 删除前钩子
type IBeforeDelete interface {
	BeforeDelete(*DB) error
}

// IAfterDelete 删除后钩子
type IAfterDelete interface {
	AfterDelete(*DB) error
}
```
