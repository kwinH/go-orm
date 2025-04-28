package orm

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/kwinh/go-orm/schema"
)

func (d *DB) Alias(tableAlias string) *DB {
	db := d.getInstance()

	db.tableAlias = tableAlias
	return db
}

func (d *DB) Raw(sql string, param ...any) *DB {
	db := d.getInstance()
	db.sql = sql
	db.bindings = param
	return db
}

func (d *DB) WithDelete() *DB {
	db := d.getInstance()
	db.withDel = true
	return db
}

func (d *DB) Model(value any) *DB {
	db := d.getInstance()
	tableInfo := db.getTableInfo(value)
	return db.setTableName(tableInfo)
}

func (d *DB) setTableName(tableInfo *schema.Schema) *DB {
	db := d.getInstance()
	tableName := tableInfo.TableName

	if db.tableAlias != "" {
		tableName = tableName + " as " + db.tableAlias
	}
	db.b.Table(tableName)
	return db
}

func (d *DB) rowsBuildMap(rows *sql.Rows, tableInfo *schema.Schema) error {
	cols, err := rows.Columns()
	if err != nil {
		return err
	}

	values := make([][]byte, len(cols))
	scans := make([]any, len(cols))
	for i := range values {
		scans[i] = &values[i]
	}

	if tableInfo.FirstType.Kind() == reflect.Map {
		if rows.Next() {
			if err := rows.Scan(scans...); err != nil {
				return err
			}
			for k, v := range values {
				key := cols[k]
				tableInfo.Value.SetMapIndex(reflect.ValueOf(key), reflect.ValueOf(v))
			}
		}
	} else if tableInfo.FirstType.Kind() == reflect.Slice {
		index := 0
		for rows.Next() {
			if err := rows.Scan(scans...); err != nil {
				return err
			}
			newMap := make(map[string]any)
			tableInfo.Value.Set(reflect.Append(tableInfo.Value, reflect.ValueOf(newMap)))
			mapValue := tableInfo.Value.Index(index)
			for k, v := range values {
				key := cols[k]
				mapValue.SetMapIndex(reflect.ValueOf(key), reflect.ValueOf(v))
			}
			index++
		}
	}

	return nil
}

func (d *DB) Get(value any) (err error) {
	defer d.resetClone()

	db := d.getInstance()

	tableInfo := db.getTableInfo(value)

	if db.sql == "" {
		if model, ok := tableInfo.Model.(IBeforeQuery); ok {
			err := model.BeforeQuery(db)
			if err != nil {
				return err
			}
		}

		if len(db.b.GetField()) == 0 {
			db.b.Select(db.getField()...)
		}

		if db.b.GetTable() == "" {
			db = db.setTableName(tableInfo)
		}

		if db.withDel == false && tableInfo.GetField("DeletedAt") != nil {
			db.WhereNull(db.b.TableAlias + ".deleted_at")
		}

		db.sql, db.bindings = db.b.ToSql()
	}

	rows, err := db.Query(db.sql, db.bindings...)
	if err != nil {
		return err
	}

	defer func() {
		if err2 := rows.Close(); err2 != nil {
			if err != nil {
				err = fmt.Errorf("%w；%v", err, err2)
			} else {
				err = err2
			}
		}
	}()

	if tableInfo.Type.Kind() == reflect.Map {
		return db.rowsBuildMap(rows, tableInfo)
	}

	withs := db.makeWiths(tableInfo)

	var dests []reflect.Value

	dest := reflect.New(tableInfo.Type).Elem()

	var getAttrFunc func()
	var afterQueryFunc func(*DB) error
	if model, ok := dest.Addr().Interface().(IGetAttr); ok {
		getAttrFunc = model.GetAttr
	}

	if model, ok := dest.Addr().Interface().(IAfterQuery); ok {
		afterQueryFunc = model.AfterQuery
	}

	var values = make([]any, len(tableInfo.Fields))
	var jsonFields = map[int]string{}

	for i, field := range tableInfo.Fields {
		if field.IsJson {
			var str []byte
			jsonFields[i] = field.Name
			values[i] = &str
		} else {
			values[i] = dest.FieldByName(field.Name).Addr().Interface()
		}
	}

	found := false
	for rows.Next() {
		found = true
		err = db.rowHandle(rows, dest, values, jsonFields, getAttrFunc, afterQueryFunc)
		if err == nil {
			dests = append(dests, dest)
			db.getWiths(withs, dest)
		} else {
			db.AddError(err)
		}
	}

	if !found {
		return ErrNotFind
	}

	if len(dests) == 0 {
		if db.Error != nil {
			return db.Error
		}
		return ErrNotFind
	}

	db.relationships(withs)

	db.setDestRelationships(dests, withs, tableInfo.Value)

	if db.Error != nil {
		return db.Error
	}

	return nil
}

func (d *DB) Find(value any, id int64) error {
	db := d.getInstance()

	tableInfo := db.getTableInfo(value)

	if err := db.Where(tableInfo.PrimaryKey.FieldName, id).Limit(1).Get(value); err != nil {
		return err
	}
	return nil
}

func (d *DB) First(value any) error {
	db := d.getInstance()

	if err := db.Limit(1).Get(value); err != nil {
		return err
	}

	return nil
}

func (d *DB) rowHandle(rows *sql.Rows, dest reflect.Value, values []any, jsonFields map[int]string, getAttrFunc func(), afterQueryFunc func(*DB) error) (err error) {

	if err = rows.Scan(values...); err != nil {
		return
	}

	for i, fieldName := range jsonFields {
		data := values[i].(*[]byte)
		if len(*data) > 0 {
			if err = json.Unmarshal(*data, dest.FieldByName(fieldName).Addr().Interface()); err != nil {
				return
			}
		}
	}

	if getAttrFunc != nil {
		getAttrFunc()
	}

	if afterQueryFunc != nil {
		if err = afterQueryFunc(d); err != nil {
			return
		}
	}

	return
}

func (d *DB) Value(field string, value any) (err error) {
	defer d.resetClone()
	db := d.getInstance()

	db.sql, db.bindings = db.b.Select(field).Limit(1).ToSql()

	rows, err := db.Query(db.sql, db.bindings...)
	if err != nil {
		return
	}

	defer rows.Close()
	rows.Next()
	rows.Scan(value)

	return
}
