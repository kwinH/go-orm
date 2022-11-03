package schema

import (
	"reflect"
)

type With struct {
	ModelType     reflect.Type
	Type          WithType
	Schema        *Schema
	Name          string
	LocalKey      *Field
	ForeignKey    *Field
	Values        []interface{}
	Relationships map[interface{}][]reflect.Value
}

func MakeWith(p reflect.StructField, dialect IDialect, tagSettings map[string]string, schema1 *Schema) *With {

	pType := p.Type
	withType := One

	if pType.Kind() == reflect.Slice {
		withType = Many
		pType = pType.Elem()
	}

	if pType.Kind() == reflect.Struct {

		switch pType.String() {
		case "time.Time", "sql.NullTime":
			return nil
		}

		with := &With{
			Type:          withType,
			Name:          p.Name,
			ModelType:     pType,
			Relationships: make(map[interface{}][]reflect.Value, 0),
			Schema:        Parse(reflect.New(pType).Interface(), dialect, schema1.TablePrefix),
		}

		foreignKey, ok := tagSettings["foreignKey"]
		if !ok {
			if schema1.PrimaryKey != nil {
				foreignKey = schema1.TableName + "_" + schema1.PrimaryKey.FieldName
			}
		}

		foreignKeyField := with.Schema.GetField(foreignKey)
		if foreignKeyField != nil {
			with.ForeignKey = foreignKeyField
			localKey, ok := tagSettings["localKey"]

			if ok {
				with.LocalKey = schema1.GetField(localKey)
			}

			if with.LocalKey == nil {
				with.LocalKey = schema1.PrimaryKey
			}

			return with
		}
	}
	return nil
}
