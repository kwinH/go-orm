package orm

import (
	"fmt"
	"github.com/kwinH/go-oorm/drive/sqlite3"
	"testing"
)

func TestDB_Fin1(t *testing.T) {
	orm, err := Open(sqlite3.Open("max.db"))

	if err != nil {
		fmt.Println("error")
		fmt.Println(err)
	} else {
		fmt.Println("ok")
		fmt.Println(orm)
	}

}
