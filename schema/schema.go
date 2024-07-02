package schema

import (
	"database/sql"
	"encoding/json"
	"reflect"
	"strings"
	"time"
)

type WithType uint

const (
	One WithType = iota
	Many
)

type IndexType string

const (
	PrimaryKey  IndexType = "PRIMARY KEY"
	UNIQUEKEY   IndexType = "UNIQUE KEY"
	INDEXKEY    IndexType = "KEY"
	FULLTEXTKEY IndexType = "FULLTEXT KEY"
)

type TimeType int64

var TimeReflectType = reflect.TypeOf(time.Time{})

var schemas = make(map[string]*Schema)

type Index struct {
	Priority int
	Field    *Field
}

type IndexList map[string][]Index

// Schema represents a table of database
type Schema struct {
	TablePrefix string
	Model       any
	Value       reflect.Value
	FirstType   reflect.Type
	Type        reflect.Type
	Name        string
	TableName   string
	FieldNames  []any
	Fields      []*Field
	fieldMap    map[string]*Field
	Withs       map[string]*With
	PrimaryKey  *Field
	IndexKeys   IndexList
	UniqueKeys  IndexList
	FullKeys    IndexList
	ExtendModel bool
}

// GetField returns field by name
func (schema *Schema) GetField(fieldName string) *Field {
	return schema.fieldMap[fieldName]
}

// RecordValues Values return the values of dest's member variables
func (schema *Schema) RecordValues(omitEmpty, isUpdate bool) map[string]any {
	fieldValues := make(map[string]any)
	haveCreatedAt := false
	haveUpdatedAt := false
	for _, field := range schema.Fields {
		if schema.ExtendModel {
			if haveCreatedAt == false && field.Name == "CreatedAt" {
				haveCreatedAt = true
			} else if haveUpdatedAt == false && field.Name == "UpdatedAt" {
				haveUpdatedAt = true
			}
		}

		if schema.RecordValue(field, omitEmpty, isUpdate) {
			fieldValues[field.FieldName] = field.Value
		}
	}

	if schema.ExtendModel && haveCreatedAt == false {
		field := schema.GetField("CreatedAt")
		if schema.RecordValue(field, omitEmpty, isUpdate) {
			fieldValues[field.FieldName] = field.Value
		}
	}

	if schema.ExtendModel && haveUpdatedAt == false {
		field := schema.GetField("UpdatedAt")
		if schema.RecordValue(field, omitEmpty, isUpdate) {
			fieldValues[field.FieldName] = field.Value
		}
	}
	return fieldValues
}

func (schema *Schema) RecordValue(field *Field, omitEmpty, isUpdate bool) bool {
	if field.Raw {
		return false
	}

	if field.AutoIncrement {
		return false
	}

	if field.Name == "CreatedAt" ||
		field.Name == "UpdatedAt" {

		if isUpdate && field.Name == "CreatedAt" {
			return false
		}

		if field.DataType == Time {
			accuracy := strings.Repeat("0", int(field.Size))
			if accuracy != "" {
				accuracy = "." + accuracy
			}
			field.Value = time.Now().Format("2006-01-02 15:04:05" + accuracy)
			return true
		} else if field.Size == 32 && (field.DataType == Int || field.DataType == Uint) {
			field.Value = time.Now().Unix()
			return true
		} else if field.Size == 64 && field.DataType == Int {
			field.Value = time.Now().UnixMilli()
			return true
		} else if field.Size == 64 && field.DataType == Uint {
			field.Value = time.Now().UnixMicro()
			return true
		}
	}

	destVal := schema.Value.FieldByName(field.Name)
	value := destVal.Interface()
	if destVal.IsZero() {
		if omitEmpty {
			return false
		}

		if field.AutoIncrement {
			return false
		}

		if field.DefaultValue == DefaultNull {
			field.Value = sql.NullString{}
			return true
		}

		if field.HavDefaultValue == true {
			field.Value = field.DefaultValue
		} else {
			field.Value = value
		}
		return true
	}

	if field.DataType == Bool {
		if value == true {
			field.Value = 1
		} else {
			field.Value = 0
		}
		return true
	}

	if field.IsJson {
		val, err := json.Marshal(value)
		if err != nil {
			return false
		}
		field.Value = string(val)
		return true
	}

	field.Value = value
	return true
}

// Parse a struct to a Schema instance
func Parse(dest any, dialect IDialect, tablePrefix string) *Schema {
	modelValue := reflect.Indirect(reflect.ValueOf(dest))
	modelType := modelValue.Type()
	firstType := modelType

	if modelType.Kind() == reflect.Slice || modelType.Kind() == reflect.Array || modelType.Kind() == reflect.Ptr {
		modelType = modelType.Elem()
	}

	model := reflect.New(modelType).Interface()
	var tableName string
	t, ok := model.(ITableName)
	if !ok {
		tableName = SnakeString(tablePrefix + modelType.Name())
	} else {
		tableName = t.TableName()
	}

	cacheKey := dialect.Name() + dialect.GetDSN() + tableName
	if schema, ok := schemas[cacheKey]; ok {
		schema.Value = modelValue
		return schema
	}

	schema := &Schema{
		TablePrefix: tablePrefix,
		Value:       modelValue,
		FirstType:   firstType,
		Type:        modelType,
		Model:       model,
		Name:        modelType.Name(),
		TableName:   tableName,
		fieldMap:    make(map[string]*Field),
		Withs:       make(map[string]*With),
		IndexKeys:   make(IndexList),
		UniqueKeys:  make(IndexList),
		FullKeys:    make(IndexList),
	}

	if modelType.Kind() == reflect.Struct {
		for i := 0; i < modelType.NumField(); i++ {
			p := modelType.Field(i)

			if p.Anonymous {
				if !schema.ExtendModel && p.Type.String() == "oorm.Model" {
					schema.ExtendModel = true
				}
				defaultModelType := reflect.New(p.Type).Type().Elem()
				for j := 0; j < defaultModelType.NumField(); j++ {
					p1 := defaultModelType.Field(j)
					parseField(p1, dialect, schema, true)
				}
			} else {
				parseField(p, dialect, schema, false)
			}

		}
	}
	schemas[cacheKey] = schema

	return schema
}

func MakeSlice(ModelType reflect.Type) reflect.Value {
	slice := reflect.MakeSlice(reflect.SliceOf(ModelType), 0, 20)
	results := reflect.New(slice.Type())
	results.Elem().Set(slice)
	return results
}
