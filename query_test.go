package oorm

import (
	sqlBuilder "github.com/kwinH/go-sql-builder"
	"testing"
)

func TestDB_Find(t *testing.T) {
	var u User
	err := orm.Where("id", ">", 0).Where("user_name", "kwin").First(&u)

	if err != nil {
		t.Error(err)
	}
	t.Log(u)
}

func TestDB_Get(t *testing.T) {
	var u []User
	err := orm.Where("id", ">", 0).Select("user_name", "status as c").Get(&u)

	if err != nil {
		t.Error(err)
	}
	t.Log(u)
}

func TestDB_GetMap(t *testing.T) {
	var u = make([]map[string]any, 0)
	err := orm.Table("user").Where("id", ">", 0).Select("user_name", "status as c").Get(&u)

	if err != nil {
		t.Error(err)
	}
	t.Logf("%#v", u)
}

func TestDB_Value(t *testing.T) {

	var userName string
	err := orm.Table("user").Where("id", "=", 1).Value("user_name", &userName)

	if err != nil {
		t.Error(err)
	}
	t.Logf("%#v", userName)
}

func TestDB_Max(t *testing.T) {
	val, err := orm.Table("user").Max("id")

	if err != nil {
		t.Error(err)
	}
	t.Logf("%#v", val)
}

func TestDB_Mix(t *testing.T) {
	val, err := orm.Model(User{}).Min("id")

	if err != nil {
		t.Error(err)
	}
	t.Logf("%#v", val)
}

func TestDB_Sum(t *testing.T) {
	val, err := orm.Model(User{}).Sum("id")

	if err != nil {
		t.Error(err)
	}
	t.Logf("%#v", val)
}

func TestDB_Avg(t *testing.T) {
	val, err := orm.Model(User{}).Avg("id")

	if err != nil {
		t.Error(err)
	}
	t.Logf("%#v", val)
}

func TestDB_Count(t *testing.T) {

	//SELECT COUNT(*) FROM `user`
	val, err := orm.Table("user").Count()

	if err != nil {
		t.Error(err)
	}
	t.Logf("%#v", val)

	type User2 struct {
		Status int8
		C      int `db:"raw:count(*)"`
	}

	var t2 []User2

	//SELECT `status`,count(*) FROM `user` WHERE `user`.`deleted_at` IS NULL GROUP BY `status`
	err = orm.Table("user").Group("status").Get(&t2)

	if err != nil {
		t.Error(err)
	}
	t.Logf("%#v", t2)

	//SELECT COUNT(*) FROM (SELECT * FROM `user` GROUP BY `status`) as `tmp1`
	val, err = orm.Table("user").Group("status").Count()

	if err != nil {
		t.Error(err)
	}
	t.Logf("%#v", val)
}

func TestDB_SubQuery(t *testing.T) {
	type User2 struct {
		Status int8
		C      int
	}
	var users []User2
	err := orm.Table(func(m *sqlBuilder.Builder) {
		m.Table("user").Select("status", "count(*) as c").Group("status").Having("c", ">", 1)
	}).Where("status", ">", 1).Get(&users)
	if err != nil {
		t.Error(err)
	}
	t.Logf("%#v", users)
}
