package oorm

import (
	"database/sql"
	"github.com/kwinH/go-oorm/schema"
	"reflect"
	"sync"
	"time"
)

func (d *DB) OmitEmpty() *DB {
	db := d.getInstance()
	db.omitEmpty = true
	return db
}

func (d *DB) insertReplace(mode string, args ...interface{}) (result int64, err error) {
	defer d.resetClone()
	db := d.getInstance()

	fieldType := reflect.TypeOf(args[0])
	if fieldType.Kind() == reflect.Ptr || fieldType.Kind() == reflect.Slice || fieldType.Kind() == reflect.Array {
		fieldType = fieldType.Elem()
	}

	if fieldType.Kind() == reflect.Struct {
		tableInfo := db.getTableInfo(args[0])

		if model, ok := tableInfo.Value.Addr().Interface().(IBeforeCreate); ok {
			err = model.BeforeCreate(db)
			if err != nil {
				return
			}
		}

		if len(db.withs) > 0 {
			withs := db.makeWiths(tableInfo)
			if db.tx == nil {
				db.Transaction(func(query *DB) error {
					result, err = query.Select(db.getField()...).withCreateGroup(withs, args...)
					return err
				})
				return
			} else {
				result, err = db.withCreateGroup(withs, args...)
				return
			}
		}
	}

	argsMap, structParams := db.structToMap(args...)

	var sql string
	var params []interface{}
	if mode == "REPLACE" {
		sql, params = db.b.Replace(argsMap...)
	} else {
		sql, params = db.b.Insert(argsMap...)
	}

	res, err := db.Exec(sql, params...)

	if err != nil {
		return
	}

	if len(structParams) > 0 {

		tableInfo := db.schema
		if tableInfo.PrimaryKey != nil &&
			(tableInfo.PrimaryKey.DataType == schema.Int ||
				tableInfo.PrimaryKey.DataType == schema.Uint) {

			var id int64
			id, err = res.LastInsertId()
			if err != nil {
				return
			}

			for i, arg := range structParams {
				argValue := reflect.ValueOf(arg).Elem()
				id1 := id + int64(i)
				if tableInfo.PrimaryKey.DataType == schema.Int {
					argValue.FieldByName(tableInfo.PrimaryKey.Name).SetInt(id1)
				} else if tableInfo.PrimaryKey.DataType == schema.Uint {
					argValue.FieldByName(tableInfo.PrimaryKey.Name).SetUint(uint64(id1))
				}

				if model, ok := argValue.Addr().Interface().(IAfterCreate); ok {
					err = model.AfterCreate(db)
					if err != nil {
						return
					}
				}

			}
		}
	}

	return res.RowsAffected()
}

func (d *DB) Create(args ...interface{}) (result int64, err error) {
	return d.insertReplace("INSERT", args...)
}

func (d *DB) Insert(args ...interface{}) (result int64, err error) {
	return d.insertReplace("INSERT", args...)
}

func (d *DB) Replace(args ...interface{}) (result int64, err error) {
	return d.insertReplace("REPLACE", args...)
}

func (d *DB) withCreateGroup(withs []*With, args ...interface{}) (rowsAffected int64, err error) {
	wg := &sync.WaitGroup{}
	field := d.getField()

	rowsAffects := make(chan int64, len(args))
	for _, arg := range args {
		wg.Add(1)
		go d.createModel(wg, withs, rowsAffects, field, arg)
	}

	wg.Wait()
	close(rowsAffects)

	if d.Error != nil {
		err = d.Error
		return
	}

	for {
		if affected, ok := <-rowsAffects; ok {
			rowsAffected += affected
		} else {
			break
		}
	}

	return
}

func (d *DB) createModel(wg *sync.WaitGroup, withs []*With, rowsAffected chan int64, field []interface{}, arg interface{}) {
	defer wg.Done()

	if d.Error != nil {
		return
	}

	affected, err := d.ClonePure().Select(field...).Create(arg)

	if err != nil {
		d.AddError(err)
		return
	}

	rowsAffected <- affected

	wg1 := &sync.WaitGroup{}

	for _, with := range withs {
		wg1.Add(1)
		go d.withCreate(arg, with, wg1)
	}
	wg1.Wait()
}

func (d *DB) withCreate(arg interface{}, with *With, wg *sync.WaitGroup) {
	defer wg.Done()

	argValue := reflect.ValueOf(arg).Elem()
	withModel := argValue.FieldByName(with.Name)

	if withModel.IsZero() {
		return
	}

	if withModel.Kind() == reflect.Slice || withModel.Kind() == reflect.Array {
		for i := 0; i < withModel.Len(); i++ {
			foreignKey := withModel.Index(i).FieldByName(with.ForeignKey.Name)
			foreignKey.Set(argValue.FieldByName(with.LocalKey.Name))
		}
	} else {
		foreignKey := withModel.FieldByName(with.ForeignKey.Name)
		foreignKey.Set(argValue.FieldByName(with.LocalKey.Name))
	}

	db := d.ClonePure(1)

	for modelName, funcList := range d.childWiths {
		db.With(modelName, funcList...)
	}

	with.Callback(db)

	_, err := db.Create(withModel.Addr().Interface())

	if err != nil {
		d.AddError(err)
	}
}

func (d *DB) Delete(value interface{}, force ...bool) (affected int64, err error) {
	defer d.resetClone()
	db := d.getInstance()
	tableInfo := db.getTableInfo(value)

	if model, ok := tableInfo.Value.Addr().Interface().(IBeforeDelete); ok {
		if err = model.BeforeDelete(db); err != nil {
			return
		}
	}

	if db.b.GetTable() == "" {
		db.b.Table(tableInfo.TableName)
	}

	if len(db.b.GetWhere()) == 0 &&
		tableInfo.PrimaryKey != nil &&
		tableInfo.RecordValue(tableInfo.PrimaryKey, db.omitEmpty, true) == true {
		db.Where(tableInfo.PrimaryKey.FieldName, tableInfo.PrimaryKey.Value)
	}

	if tableInfo.GetField("DeletedAt") != nil &&
		(len(force) == 0 || force[0] == false) {
		affected, err = db.softDelete()
	} else {
		if len(db.b.GetWhere()) == 0 {
			return 0, ErrMissingCondition
		}

		var result sql.Result

		sql, params := db.b.Delete()

		if result, err = db.Exec(sql, params...); err != nil {
			return
		}

		affected, err = result.RowsAffected()

		return
	}

	if err != nil {
		db.AddError(err)
	}

	if model, ok := tableInfo.Value.Addr().Interface().(IAfterDelete); ok {
		if err1 := model.AfterDelete(db); err1 != nil {
			db.AddError(err1)
		}
	}

	err = db.Error
	return
}

func (d *DB) softDelete() (int64, error) {
	return d.Update(map[string]interface{}{
		"deleted_at": time.Now().Format("2006-01-02 15:04:05.000"),
	})
}

func (d *DB) Update(arg interface{}) (affected int64, err error) {
	defer d.resetClone()
	db := d.getInstance()

	argType := reflect.TypeOf(arg)
	if argType.Kind() == reflect.Ptr {
		argType = argType.Elem()
	}

	var tableInfo *schema.Schema
	kind := argType.Kind()
	if kind == reflect.Struct {
		tableInfo = db.getTableInfo(arg)
		if model, ok := tableInfo.Value.Addr().Interface().(IBeforeUpdate); ok {
			err = model.BeforeUpdate(db)
			if err != nil {
				return
			}
		}

		if len(db.withs) > 0 {
			withs := db.makeWiths(db.Parse(arg))
			if db.tx == nil {
				db.Transaction(func(query *DB) error {
					query = query.ClonePure(1)
					query.b = db.b
					affected, err = query.withUpdates(withs, arg)
					return err
				})
				return
			} else {
				affected, err = db.withUpdates(withs, arg)
				return
			}
		}
	}

	var argToMap map[string]interface{}

	switch kind {
	case reflect.Struct:
		if db.b.GetTable() == "" {
			db.b.Table(tableInfo.TableName)
		}

		argToMap = tableInfo.RecordValues(db.omitEmpty, true)

		if len(d.b.GetWhere()) == 0 {
			if tableInfo.PrimaryKey == nil ||
				tableInfo.PrimaryKey.Value == nil ||
				reflect.ValueOf(tableInfo.PrimaryKey.Value).IsZero() {
				return 0, ErrMissingCondition
			} else {
				db.Where(tableInfo.PrimaryKey.FieldName, tableInfo.PrimaryKey.Value)
			}
		}
	case reflect.Map:
		ok := false
		if argToMap, ok = arg.(map[string]interface{}); ok {
			if len(d.b.GetWhere()) == 0 {
				return 0, ErrMissingCondition
			}

			if d.b.GetTable() == "" {
				return 0, ErrMissingTableName
			}
		}
	}

	if len(argToMap) == 0 {
		return 0, ErrParam
	}

	sql, params := db.b.Update(argToMap)

	result, err := db.Exec(sql, params...)
	if err != nil {
		return 0, err
	}

	if tableInfo != nil {
		if model, ok := tableInfo.Value.Addr().Interface().(IAfterUpdate); ok {
			err = model.AfterUpdate(db)
			if err != nil {
				return
			}
		}
	}

	return result.RowsAffected()
}

func (d *DB) withUpdates(withs []*With, arg interface{}) (i int64, err error) {
	wg := &sync.WaitGroup{}

	argValue := reflect.ValueOf(arg).Elem()
	for _, with := range withs {
		wg.Add(1)
		go d.withUpdate(with, argValue, wg)
	}

	d.withs = make(map[string]WithFunc, 0)
	wg.Add(1)
	ch := make(chan int64, 1)

	go func() {
		defer wg.Done()
		res, err := d.Update(arg)
		ch <- res
		if err != nil {
			d.AddError(err)
		}
	}()

	wg.Wait()

	return <-ch, d.Error
}

func (d *DB) withUpdate(with *With, argValue reflect.Value, wg *sync.WaitGroup) {
	defer wg.Done()

	db := d.ClonePure(1)

	for modelName, funcList := range d.childWiths {
		db.With(modelName, funcList...)
	}

	with.Callback(db)

	withModel := argValue.FieldByName(with.Name)

	if withModel.IsZero() {
		return
	}

	if with.Type == schema.One {
		val := argValue.FieldByName(with.LocalKey.Name)
		if val.IsZero() {
			return
		}

		_, err := db.Where(with.ForeignKey.FieldName, val.Interface()).
			Update(withModel.Addr().Interface())

		if err != nil {
			d.AddError(err)
		}
	}
}
