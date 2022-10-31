package oorm

import (
	"testing"
)

func TestDB_Find(t *testing.T) {
	var u User
	err := baseDB.Where("id", ">", 0).Where("user_name", "kwin").First(&u)

	if err != nil {
		t.Error(err)
	}
	t.Log(u)
}

func TestDB_Get(t *testing.T) {
	var u []User
	err := baseDB.Where("id", ">", 0).Where("user_name", "kwin").Get(&u)

	if err != nil {
		t.Error(err)
	}
	t.Log(u)
}
