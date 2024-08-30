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
	haveCreatedAt, haveUpdatedAt := false, false

	for _, field := range schema.Fields {
		if schema.ExtendModel {
			if !haveCreatedAt && field.Name == "CreatedAt" {
				haveCreatedAt = true
			} else if !haveUpdatedAt && field.Name == "UpdatedAt" {
				haveUpdatedAt = true
			}
		}

		if schema.RecordValue(field, omitEmpty, isUpdate) {
			fieldValues[field.FieldName] = field.Value
		}
	}

	if schema.ExtendModel {
		if !haveCreatedAt {
			schema.addDefaultTimeValue(fieldValues, "CreatedAt", omitEmpty, isUpdate)
		}
		if !haveUpdatedAt {
			schema.addDefaultTimeValue(fieldValues, "UpdatedAt", omitEmpty, isUpdate)
		}
	}
	return fieldValues
}

func (schema *Schema) addDefaultTimeValue(fieldValues map[string]any, fieldName string, omitEmpty, isUpdate bool) {
	field := schema.GetField(fieldName)
	if schema.RecordValue(field, omitEmpty, isUpdate) {
		fieldValues[field.FieldName] = field.Value
	}
}

func (schema *Schema) RecordValue(field *Field, omitEmpty, isUpdate bool) bool {
	if field.Raw || field.AutoIncrement {
		return false
	}

	if field.Name == "CreatedAt" || field.Name == "UpdatedAt" {
		if isUpdate && field.Name == "CreatedAt" {
			return false
		}
		return schema.setTimeValue(field)
	}

	destVal := schema.Value.FieldByName(field.Name)
	value := destVal.Interface()

	if destVal.IsZero() {
		if omitEmpty || field.AutoIncrement {
			return false
		}
		if field.DefaultValue == DefaultNull {
			field.Value = sql.NullString{}
			return true
		}
		field.Value = schema.getDefaultFieldValue(field, value)
		return true
	}

	if field.DataType == Bool {
		field.Value = boolToInt(value.(bool))
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

func (schema *Schema) setTimeValue(field *Field) bool {
	now := time.Now()
	switch {
	case field.DataType == Time:
		field.Value = formatTime(now, field.Size)
	case field.Size == 32 && (field.DataType == Int || field.DataType == Uint):
		field.Value = now.Unix()
	case field.Size == 64 && field.DataType == Int:
		field.Value = now.UnixMilli()
	case field.Size == 64 && field.DataType == Uint:
		field.Value = now.UnixMicro()
	default:
		return false
	}
	return true
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func formatTime(t time.Time, size int64) string {
	accuracy := strings.Repeat("0", int(size))
	if accuracy != "" {
		accuracy = "." + accuracy
	}
	return t.Format("2006-01-02 15:04:05" + accuracy)
}

func (schema *Schema) getDefaultFieldValue(field *Field, defaultValue any) any {
	if field.HavDefaultValue {
		return field.DefaultValue
	}
	return defaultValue
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
	tableName := getTableName(model, tablePrefix)

	cacheKey := dialect.Name() + dialect.GetDSN() + tablePrefix + tableName
	if schema, ok := schemas[cacheKey]; ok {
		schema.Value = modelValue
		return schema
	}

	schema := createSchema(tablePrefix, modelValue, firstType, modelType, model, tableName)

	parseStructFields(schema, modelType, dialect)

	schemas[cacheKey] = schema
	return schema
}

func getTableName(model any, tablePrefix string) string {
	if t, ok := model.(ITableName); ok {
		return t.TableName()
	}
	return SnakeString(tablePrefix + reflect.TypeOf(model).Elem().Name())
}

func createSchema(tablePrefix string, modelValue reflect.Value, firstType, modelType reflect.Type, model any, tableName string) *Schema {
	return &Schema{
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
}

func parseStructFields(schema *Schema, modelType reflect.Type, dialect IDialect) {
	for i := 0; i < modelType.NumField(); i++ {
		field := modelType.Field(i)
		if field.Anonymous {
			parseAnonymousField(schema, field, dialect)
		} else {
			parseField(field, dialect, schema, false)
		}
	}
}

func parseAnonymousField(schema *Schema, field reflect.StructField, dialect IDialect) {
	if !schema.ExtendModel && field.Type.String() == "orm.Model" {
		schema.ExtendModel = true
	}
	defaultModelType := reflect.New(field.Type).Type().Elem()
	for j := 0; j < defaultModelType.NumField(); j++ {
		p1 := defaultModelType.Field(j)
		parseField(p1, dialect, schema, true)
	}
}

func MakeSlice(ModelType reflect.Type) reflect.Value {
	slice := reflect.MakeSlice(reflect.SliceOf(ModelType), 0, 20)
	results := reflect.New(slice.Type())
	results.Elem().Set(slice)
	return results
}
