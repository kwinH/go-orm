package oorm

import (
	"crypto/md5"
	"fmt"
	sqlBuilder "github.com/kwinH/go-sql-builder"
	"oorm/drive/mysql"
	"testing"
)

type Test struct {
	Id      uint
	OrderId uint
	Test    string
}

type Order struct {
	Id      uint `db:"primaryKey"`
	UserId  uint
	OrderNo string
	Ex      map[string]interface{} `db:"json"`
	Test    Test
}
type User struct {
	Model
	UserName string
	Password string
	Nickname string
	Status   int8
	Avatar   string
	//	Contact  Contact `db:"foreignKey:UserId"`
	Order []Order
}

type Contact struct {
	Model
	UserId uint
	Mobile string
	Email  string
	User   User `db:"localKey:UserId;foreignKey:Id"`
}

var _ IGetAttr = (*User)(nil)
var _ ISetAttr = (*User)(nil)

func (u *User) GetAttr() {
	//user := value.(*User)
	u.UserName = "Aaa"
	//fmt.Printf("%#v", value)
}

func (u *User) SetAttr() {
	if u.Password != "" {
		u.Password = fmt.Sprintf("%x", md5.Sum([]byte(u.Password)))
	}
}

var baseDB *DB

func init() {
	var err error
	baseDB, err = Open(mysql.Open("root:root@tcp(127.0.0.1:3306)/gin_demo?parseTime=true"))

	if err != nil {
		fmt.Printf("%#v", err.Error())
		panic("连接数据库失败")
	}
}

func TestModel_get(t *testing.T) {
	s := sqlBuilder.Raw("aaa")

	fmt.Printf("%#v", s == "aaa")

	return
	var users []User

	db := baseDB.Select("id,user_name")
	db.Clone().Where("id", ">", 1).Get(&users)
	err := db.Clone().Get(&users)
	//.With("Order.Test", func(db *DB) {
	//	fmt.Printf("call db %p\n", db)
	//	db.Where("order_no", "<>", "sss")
	//}, func(db *DB) {
	//	fmt.Printf("call db %p\n", db)
	//	db.Where("test", "<>", "sss")
	//}).Get(&users)

	if err != nil {
		fmt.Printf("==err: %#v\n", err)
		return
	}

	//for _, user := range users {
	//	fmt.Printf("user:%v\n", user.UserName)
	//}
}

func TestModel_Insert(t *testing.T) {

	type User20 struct {
		Model
		Name     string `db:"index:us|1"`
		Password string
		Status   int8
		Age      int8
		Sex      int8
		Balance  float64 `db:"decimal:10,2"`
	}

	var user2 User20
	baseDB.Omit("age", "sex").Get(&user2)

}

func TestModel_Insert1(t *testing.T) {
	//baseDB.Table("user").Create(map[string]interface{}{
	//	"user_name": "aaa",
	//})
	//return
	users := []User{
		{
			UserName: "kwin",
		},
		{
			UserName: "kwin2",
		},
	}

	_, err := baseDB.Select("user_name,status").Create(&users)
	//_, err = baseDB.Select("user_name,status").Create(&users)
	fmt.Printf("err:%#v\n", err)
	//fmt.Printf("user:%v", users[0].Id)
}

func TestDB_Migrate(t *testing.T) {
	type User struct {
		Model
		CreatedAt int
		UpdatedAt int64
		UserName  string `db:"index:us|1"`
		Password  string
		Nickname  string
		Status    int8
		Avatar    string
		Balance   float64 `db:"decimal:10,2"`
	}

	value := User{}

	baseDB.Migrate.Auto(value, true, true)
}
