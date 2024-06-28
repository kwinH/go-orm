package oorm

import (
	"crypto/md5"
	"fmt"
	"github.com/kwinH/go-oorm/drive/mysql"
	"strings"
	"testing"
)

type User struct {
	Model
	UserName string
	Password string
	Nickname string
	Status   int8
	Avatar   string
}

var _ IGetAttr = (*User)(nil)
var _ ISetAttr = (*User)(nil)

func (u *User) GetAttr() {
	u.UserName = strings.Trim(u.UserName, "")
}

func (u *User) SetAttr() {
	if u.Password != "" {
		u.Password = fmt.Sprintf("%x", md5.Sum([]byte(u.Password)))
	}
}

var orm *DB

func init() {
	var err error
	orm, err = Open(mysql.Open("root:root@tcp(127.0.0.1:3306)/oorm_demo?parseTime=true"))

	if err != nil {
		fmt.Printf("%#v", err.Error())
		panic("连接数据库失败")
	}
}

func TestDB_Migrate(t *testing.T) {
	value := User{}

	orm.Migrate.Auto(value, true, true)
}
