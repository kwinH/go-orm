package orm

import "github.com/kwinH/go-sql-builder"

func (d *DB) Select(args ...any) *DB {
	db := d.getInstance()
	db.b.Select(args...)
	return db
}

func (d *DB) Omit(field ...string) *DB {
	db := d.getInstance()
	d.omitField = make(map[string]bool)
	for _, s := range field {
		d.omitField[s] = true
	}
	return db
}

func (d *DB) Table(table any) *DB {
	db := d.getInstance()
	db.b.Table(table)
	return db
}

func (d *DB) Where(args ...any) *DB {
	db := d.getInstance()
	db.b.Where(args...)
	return db
}

func (d *DB) OrWhere(args ...any) *DB {
	db := d.getInstance()
	db.b.OrWhere(args...)
	return db
}

func (d *DB) WhereExists(where func(*sqlBuilder.Builder)) *DB {
	db := d.getInstance()
	db.b.WhereExists(where)
	return db
}

func (d *DB) WhereNotExists(where func(*sqlBuilder.Builder)) *DB {
	db := d.getInstance()
	db.b.WhereNotExists(where)
	return db
}

func (d *DB) OrWhereExists(where func(*sqlBuilder.Builder)) *DB {
	db := d.getInstance()
	db.b.OrWhereExists(where)
	return db
}

func (d *DB) OrWhereNotExists(where func(*sqlBuilder.Builder)) *DB {
	db := d.getInstance()
	db.b.OrWhereNotExists(where)
	return db
}

func (d *DB) WhereIn(field string, value ...any) *DB {
	db := d.getInstance()
	db.b.WhereIn(field, value...)
	return db
}

func (d *DB) WhereNotIn(field string, value ...any) *DB {
	db := d.getInstance()
	db.b.WhereNotIn(field, value...)
	return db
}

func (d *DB) OrWhereIn(field string, value ...any) *DB {
	db := d.getInstance()
	db.b.OrWhereIn(field, value...)
	return db
}

func (d *DB) OrWhereNotIn(field string, value ...any) *DB {
	db := d.getInstance()
	db.b.OrWhereNotIn(field, value...)
	return db
}

func (d *DB) WhereNull(field string) *DB {
	db := d.getInstance()
	db.b.WhereNull(field)
	return db
}

func (d *DB) WhereNotNull(field string) *DB {
	db := d.getInstance()
	db.b.WhereNotNull(field)
	return db
}

func (d *DB) OrWhereNull(field string) *DB {
	db := d.getInstance()
	db.b.OrWhereNull(field)
	return db
}

func (d *DB) OrWhereNotNull(field string) *DB {
	db := d.getInstance()
	db.b.OrWhereNotNull(field)
	return db
}

func (d *DB) WhereBetween(field string, value ...any) *DB {
	db := d.getInstance()
	db.b.WhereBetween(field, value...)
	return db
}

func (d *DB) OrWhereBetween(field string, value ...any) *DB {
	db := d.getInstance()
	db.b.OrWhereBetween(field, value...)
	return db
}

func (d *DB) WhereNotBetween(field string, value ...any) *DB {
	db := d.getInstance()
	db.b.WhereNotBetween(field, value...)
	return db
}

func (d *DB) OrWhereNotBetween(field string, value ...any) *DB {
	db := d.getInstance()
	db.b.OrWhereNotBetween(field, value...)
	return db
}

func (d *DB) Group(group ...string) *DB {
	db := d.getInstance()
	db.b.Group(group...)
	return db
}

func (d *DB) Having(args ...any) *DB {
	db := d.getInstance()
	db.b.Having(args...)
	return db
}

func (d *DB) OrHaving(args ...any) *DB {
	db := d.getInstance()
	db.b.OrHaving(args...)
	return db
}

func (d *DB) Order(args ...any) *DB {
	db := d.getInstance()
	db.b.Order(args...)
	return db
}

func (d *DB) Limit(args ...int64) *DB {
	db := d.getInstance()
	db.b.Limit(args...)
	return db
}

func (d *DB) Page(page int64, listRows int64) *DB {
	db := d.getInstance()
	db.b.Page(page, listRows)
	return db
}

func (d *DB) LefJoin(table any, condition string, params ...any) *DB {
	db := d.getInstance()
	db.b.Joins(table, condition, "LEFT", params...)
	return db
}

func (d *DB) RightJoin(table any, condition string, params ...any) *DB {
	db := d.getInstance()
	db.b.Joins(table, condition, "RIGHT", params...)
	return db
}

func (d *DB) Join(table any, condition string, params ...any) *DB {
	db := d.getInstance()
	db.b.Joins(table, condition, "INNER", params...)
	return db
}

func (d *DB) ToSql() (string, []any) {
	db := d.getInstance()
	return db.b.ToSql()
}

func (d *DB) DuplicateKey(duplicateKey map[string]any) *DB {
	db := d.getInstance()
	db.b.DuplicateKey(duplicateKey)
	return d
}
