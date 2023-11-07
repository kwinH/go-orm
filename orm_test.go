package oorm

import (
	"crypto/md5"
	"fmt"
	"github.com/kwinH/go-oorm/drive/mysql"
	"reflect"
	"strings"
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
	u.UserName = strings.Trim(u.UserName, "")
	//fmt.Printf("%#v", value)
}

func (u *User) SetAttr() {
	if u.Password != "" {
		u.Password = fmt.Sprintf("%x", md5.Sum([]byte(u.Password)))
	}
}

var orm *Orm

func init() {
	var err error
	orm, err = Open(mysql.Open("root:root@tcp(127.0.0.1:3306)/oorm_demo?parseTime=true"))

	if err != nil {
		fmt.Printf("%#v", err.Error())
		panic("连接数据库失败")
	}
}

func TestModel_get(t *testing.T) {
	var user User
	user = User{}

	s, _ := reflect.TypeOf(user).FieldByName("Model")
	fmt.Printf("===%#v", s.Anonymous)

	return
	var users []User

	err := orm.NewDB().Select("id,created_at,updated_at,deleted_at,user_name,password,nickname,status,avatar").Where("id", ">", 1).Get(&users)

	if err != nil {
		fmt.Printf("==err: %#v\n", err)
		return
	}

	for _, user := range users {
		fmt.Printf("user:%#v\n", user.UserName)
	}
}

func TestModel_Insert(t *testing.T) {
	//orm.NewDB().Table("user").Create(map[string]interface{}{
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

	_, err := orm.NewDB().Select("user_name,status").Create(&users)
	//_, err = orm.NewDB().Select("user_name,status").Create(&users)
	fmt.Printf("err:%#v\n", err)
	//fmt.Printf("user:%v", users[0].Id)
}

func TestDB_Migrate(t *testing.T) {
	value := User{}

	orm.NewDB().Migrate.Auto(value, true, true)
}
