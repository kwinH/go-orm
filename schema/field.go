package schema

import (
	"database/sql"
	sqlBuilder "github.com/kwinH/go-sql-builder"
	"go/ast"
	"reflect"
	"strconv"
	"strings"
	"time"
)

type DataType string

const DefaultNull = "NULL"

// Field represents a column of database
type Field struct {
	StructField   reflect.StructField
	FieldName     string
	Name          string
	Type          string
	PrimaryKey    bool
	AutoIncrement bool
	DefaultValue  string
	Tag           string
	TagSettings   map[string]string
	DataType      DataType
	Size          int
	Comment       string
	Raw           bool
	Decimal       string
	IsJson        bool
	Value         interface{}
}

const (
	Bool   DataType = "bool"
	Int    DataType = "int"
	Uint   DataType = "uint"
	Float  DataType = "float"
	String DataType = "string"
	Time   DataType = "time"
	Json   DataType = "json"
)

func parseField(p reflect.StructField, dialect IDialect, schema *Schema, isAnonymous bool) {
	if !p.Anonymous && ast.IsExported(p.Name) {
		fieldName := SnakeString(p.Name)

		var tag string
		var tagSettings map[string]string
		if v, ok := p.Tag.Lookup("db"); ok {
			tag = v
			tagSettings = ParseTagSetting(v, ";")
		}

		field := &Field{
			StructField: p,
			FieldName:   fieldName,
			Name:        p.Name,
			Tag:         tag,
			TagSettings: tagSettings,
		}

		parseTag(field, schema)

		if !field.IsJson {
			with := MakeWith(p, dialect, tagSettings, schema)
			if with != nil {
				schema.Withs[p.Name] = with
				return
			}
		}

		setDataType(field)
		setSize(field)

		if field.Raw {
			schema.FieldNames = append(schema.FieldNames, sqlBuilder.Raw(field.FieldName))
		} else {
			schema.FieldNames = append(schema.FieldNames, field.FieldName)
		}

		field.Type = dialect.DataTypeOf(field)

		if schema.fieldMap[field.FieldName] != nil {
			if isAnonymous == false {
				delete(schema.fieldMap, field.FieldName)
				delete(schema.fieldMap, p.Name)
				for i, f := range schema.Fields {
					if f.Name == field.Name {
						schema.Fields[i] = field
					}
				}

				schema.fieldMap[field.FieldName] = field
				schema.fieldMap[p.Name] = field
			}
		} else {
			schema.Fields = append(schema.Fields, field)
			schema.fieldMap[field.FieldName] = field
			schema.fieldMap[p.Name] = field
		}
	}

}

func parseTag(field *Field, schema *Schema) {
	var err error
	if customFieldName, ok := field.TagSettings["field"]; ok {
		field.FieldName = customFieldName
	}

	if _, ok := field.TagSettings["json"]; ok {
		field.IsJson = true
	}

	if _, ok := field.TagSettings["primaryKey"]; ok {
		field.PrimaryKey = true
		schema.PrimaryKey = field
	}

	if _, ok := field.TagSettings["autoIncrement"]; ok {
		field.AutoIncrement = true
		field.PrimaryKey = true
		schema.PrimaryKey = field
	}

	if v, ok := field.TagSettings["default"]; ok {
		field.DefaultValue = v
	}

	if num, ok := field.TagSettings["size"]; ok {
		if field.Size, err = strconv.Atoi(num); err != nil {
			field.Size = -1
		}
	}

	if decimal, ok := field.TagSettings["decimal"]; ok {
		field.Decimal = decimal
	}

	if val, ok := field.TagSettings["comment"]; ok {
		field.Comment = val
	}

	if raw, ok := field.TagSettings["raw"]; ok {
		field.Raw = true
		field.FieldName = raw
	}

	if index, ok := field.TagSettings["index"]; ok {
		setIndex(index, field, schema, INDEXKEY)
	}

	if index, ok := field.TagSettings["unique"]; ok {
		setIndex(index, field, schema, UNIQUEKEY)
	}

	if index, ok := field.TagSettings["full"]; ok {
		setIndex(index, field, schema, FULLTEXTKEY)
	}
}

func setIndex(indexKey string, field *Field, schema *Schema, indexType IndexType) {
	if indexKey == "" {
		indexKey = field.FieldName
	}

	indexKeyVal := strings.Split(indexKey, ".")

	priority := 0
	if len(indexKeyVal) >= 2 {
		indexKey = indexKeyVal[0]
		priority, _ = strconv.Atoi(indexKeyVal[1])
	}

	var indexKeys IndexList
	if indexType == INDEXKEY {
		indexKey += "_key"
		indexKeys = schema.IndexKeys
	} else if indexType == UNIQUEKEY {
		indexKeys = schema.UniqueKeys
		indexKey += "_uni"
	} else if indexType == FULLTEXTKEY {
		indexKey += "_full"
		indexKeys = schema.FullKeys
	}

	if _, ok := indexKeys[indexKey]; ok {
		indexKeys[indexKey] = append(indexKeys[indexKey], Index{
			Priority: priority,
			Field:    field,
		})
	} else {
		indexList := make([]Index, 1)
		indexList[0] = Index{
			Priority: priority,
			Field:    field,
		}

		indexKeys[indexKey] = indexList
	}
}

func setDataType(field *Field) {
	switch field.StructField.Type.Kind() {
	case reflect.Bool:
		field.DataType = Bool
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		field.DataType = Int
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		field.DataType = Uint
	case reflect.Float32, reflect.Float64:
		field.DataType = Float
	case reflect.String:
		field.DataType = String
	case reflect.Struct:
		fieldValue := reflect.New(field.StructField.Type)
		if _, ok := fieldValue.Interface().(*time.Time); ok {
			field.DataType = Time
		} else if _, ok := fieldValue.Interface().(*sql.NullTime); ok {
			field.DefaultValue = DefaultNull
			field.DataType = Time
		} else if fieldValue.Type().ConvertibleTo(reflect.TypeOf(time.Time{})) ||
			fieldValue.Type().ConvertibleTo(reflect.TypeOf(&time.Time{})) {
			field.DataType = Time
		} else {
			field.DataType = Json
		}
	case reflect.Array, reflect.Slice, reflect.Map:
		field.DataType = Json
	}
}

func setSize(field *Field) {
	if field.Size == 0 {
		switch field.StructField.Type.Kind() {
		case reflect.Int64, reflect.Uint, reflect.Uint64, reflect.Float64:
			field.Size = 64
		case reflect.Int8, reflect.Uint8:
			field.Size = 8
		case reflect.Int16, reflect.Uint16:
			field.Size = 16
		case reflect.Int, reflect.Int32, reflect.Uint32, reflect.Float32:
			field.Size = 32
		}
	}
}
