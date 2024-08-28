package main

import (
	"fmt"
)

type UserModel[T any] struct {
	Id   T //可能是数字也可能是字符串
	Name string
}

func main() {
	u1 := &UserModel[int]{Id: 123, Name: "ls"} //初始化时才指定类型
	u2 := &UserModel[string]{Id: "one", Name: "zs"}
	fmt.Println(u1, u2)

	users := []string{"zhangSan", "lisi"}
	//	PrintArray2(users)
	PrintArray(users)
}

// NewUserModel 构造函数是需要指定类型
func NewUserModel[T any](id T) *UserModel[T] {
	return &UserModel[T]{Id: id}
}

func PrintArray[T any](arr []T) {
	for _, a := range arr {
		fmt.Println(a)
	}
}

func PrintArray2(arr []interface{}) {
	for _, a := range arr {
		fmt.Println(a)
	}
}
