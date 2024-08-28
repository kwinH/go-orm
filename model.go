package orm

import (
	"database/sql"
	"time"
)

type Model struct {
	Id        uint `db:"autoIncrement"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt sql.NullTime `db:"index"`
}
