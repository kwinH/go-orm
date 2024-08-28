package orm

import (
	"fmt"
	sqlBuilder "github.com/kwinH/go-sql-builder"
)

func (d *DB) aggregate(sql string) (data int64, err error) {
	defer d.resetClone()
	db := d.getInstance()

	if len(db.b.GetGroup()) > 0 {
		db.sql, db.bindings = d.ClonePure(1).b.Select(sql).
			Table(func() *sqlBuilder.Builder {
				return &db.b
			}).ToSql()
	} else {
		db.sql, db.bindings = db.b.Select(sql).ToSql()
	}

	rows, err := db.Query(db.sql, db.bindings...)
	if err != nil {
		return
	}

	defer rows.Close()
	rows.Next()
	rows.Scan(&data)

	return
}

func (d *DB) Count() (int64, error) {
	return d.aggregate("COUNT(*)")
}

func (d *DB) Max(field string) (int64, error) {
	return d.aggregate(fmt.Sprintf("MAX(%s)", field))
}

func (d *DB) Min(field string) (int64, error) {
	return d.aggregate(fmt.Sprintf("MIN(%s)", field))
}

func (d *DB) Avg(field string) (int64, error) {
	return d.aggregate(fmt.Sprintf("AVG(%s)", field))
}

func (d *DB) Sum(field string) (int64, error) {
	return d.aggregate(fmt.Sprintf("SUM(%s)", field))
}
