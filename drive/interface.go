package drive

import (
	"database/sql"
)

type IConnPool interface {
	Prepare(query string) (*sql.Stmt, error)
	Exec(query string, args ...any) (sql.Result, error)
	Query(query string, args ...any) (*sql.Rows, error)
}
