package oorm

import (
	"github.com/kwinH/go-oorm/schema"
	sqlBuilder "github.com/kwinH/go-sql-builder"
	"reflect"
	"strings"
)

func (d *DB) structToMap(args ...interface{}) ([]interface{}, []interface{}) {

	params := make([]interface{}, 0)
	structParams := make([]interface{}, 0)

	for _, arg := range args {
		isPtr := false
		fieldType := reflect.TypeOf(arg)
		if fieldType.Kind() == reflect.Ptr {
			isPtr = true
			fieldType = fieldType.Elem()
		}
		kind := fieldType.Kind()

		if kind == reflect.Struct {

			tableInfo := d.getTableInfo(arg)

			if d.b.GetTable() == "" {
				d.Table(tableInfo.TableName)
			}

			model := tableInfo.Value.Addr().Interface()
			if modelSetAttr, ok := model.(ISetAttr); ok {
				modelSetAttr.SetAttr()
			}

			params = append(params, tableInfo.RecordValues(d.omitEmpty, false))
			structParams = append(structParams, arg)
		} else if kind == reflect.Map {
			params = append(params, arg)
		} else if kind == reflect.Slice {
			ret := make([]interface{}, 0)
			v := reflect.ValueOf(arg)
			if isPtr {
				v = v.Elem()
			}

			for i := 0; i < v.Len(); i++ {
				ret = append(ret, v.Index(i).Addr().Interface())
			}

			params1, structParams1 := d.structToMap(ret...)
			params = append(params, params1...)
			structParams = append(structParams, structParams1...)
		}
	}

	return params, structParams
}

func (d *DB) Parse(value interface{}) *schema.Schema {
	s := *schema.Parse(value, d.dialector, d.TablePrefix)
	d.schema = &s
	return d.schema
}

func (d *DB) getTableInfo(value interface{}) *schema.Schema {
	db := d.getInstance()
	if db.schema != nil {
		return db.schema
	}
	schemaParse := db.Parse(value)

	field := db.getField()

	if len(field) > 0 {
		schemaParse.FieldNames = []interface{}{}
		schemaParse.Fields = []*schema.Field{}
		var name string
		var ok bool
		for _, v := range field {
			if name, ok = v.(string); !ok {
				name = string(v.(sqlBuilder.Raw))
			}
			schemaParse.FieldNames = append(schemaParse.FieldNames, name)

			aliasIndex := strings.Index(name, " ")
			if aliasIndex >= 0 {
				name = name[:aliasIndex]
			}

			fieldData := schemaParse.GetField(name)
			if fieldData == nil {
				fieldData = &schema.Field{
					FieldName: name,
				}
			}
			schemaParse.Fields = append(schemaParse.Fields, fieldData)
		}
	}

	db.schema = schemaParse
	return schemaParse
}

func (d *DB) getField() []interface{} {
	if len(d.b.GetField()) > 0 {
		return d.b.GetField()
	}

	if len(d.omitField) == 0 {
		return d.schema.FieldNames
	} else {
		fieldNames := make([]interface{}, 0)
		for _, field := range d.schema.FieldNames {
			fieldStr := ""
			if _, ok := field.(sqlBuilder.Raw); ok {
				fieldStr = string(field.(sqlBuilder.Raw))
			} else {
				fieldStr = field.(string)
			}

			if _, ok := d.omitField[fieldStr]; !ok {
				fieldNames = append(fieldNames, field)
			}
		}
		return fieldNames

	}
}
