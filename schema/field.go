package schema

import (
	sqlBuilder "github.com/kwinh/go-sql-builder"
	"go/ast"
	"log"
	"reflect"
	"strconv"
	"strings"
)

type DataType string

type null string

const DefaultNull null = "NULL"

// Field represents a column of database
type Field struct {
	StructField     reflect.StructField
	FieldName       string
	Name            string
	Type            string
	PrimaryKey      bool
	AutoIncrement   bool
	HavDefaultValue bool
	DefaultValue    any
	Tag             string
	TagSettings     map[string]string
	DataType        DataType
	Size            int64
	Comment         string
	Raw             bool
	Decimal         string
	IsJson          bool
	Value           any
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

		if existingField, exists := schema.fieldMap[field.FieldName]; exists && !isAnonymous {
			// 替换现有字段
			replaceFieldInSchema(schema, field, existingField)
		} else {
			// 添加新字段
			schema.Fields = append(schema.Fields, field)
			schema.fieldMap[field.FieldName] = field
			schema.fieldMap[p.Name] = field
		}
	}
}

func replaceFieldInSchema(schema *Schema, newField, existingField *Field) {
	for i, f := range schema.Fields {
		if f.Name == existingField.Name {
			schema.Fields[i] = newField
		}
	}
	schema.fieldMap[newField.FieldName] = newField
	schema.fieldMap[newField.StructField.Name] = newField
}

func parseTag(field *Field, schema *Schema) {
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
		field.HavDefaultValue = true
		if v == string(DefaultNull) {
			field.DefaultValue = DefaultNull
		} else {
			field.DefaultValue = v
		}
	}

	if num, ok := field.TagSettings["size"]; ok {
		if size, err := strconv.ParseInt(num, 10, 64); err == nil {
			field.Size = size
		} else {
			log.Printf("Invalid size value for field %s: %v", field.Name, num)
			field.Size = 0
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
		var err error
		indexKey = indexKeyVal[0]
		priority, err = strconv.Atoi(indexKeyVal[1])
		if err != nil {
			log.Printf("Invalid priority value for index on field %s: %v", field.Name, indexKeyVal[1])
			priority = 0
		}
	}

	var indexKeys IndexList
	switch indexType {
	case INDEXKEY:
		indexKey += "_key"
		indexKeys = schema.IndexKeys
	case UNIQUEKEY:
		indexKey += "_uni"
		indexKeys = schema.UniqueKeys
	case FULLTEXTKEY:
		indexKey += "_full"
		indexKeys = schema.FullKeys
	}

	if _, ok := indexKeys[indexKey]; ok {
		indexKeys[indexKey] = append(indexKeys[indexKey], Index{
			Priority: priority,
			Field:    field,
		})
	} else {
		indexList := []Index{{
			Priority: priority,
			Field:    field,
		}}
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
		switch field.StructField.Type.String() {
		case "sql.NullBool":
			field.DataType = Bool
		case "sql.NullInt16", "sql.NullInt32", "sql.NullInt64":
			field.DataType = Int
		case "sql.NullFloat64":
			field.DataType = Float
		case "sql.NullString":
			field.DataType = String
		case "sql.NullTime", "time.Time":
			field.DataType = Time
		default:
			field.DataType = Json
		}
	case reflect.Array, reflect.Slice, reflect.Map:
		field.DataType = Json
	default:
		log.Printf("Unsupported data type for field %s: %s", field.Name, field.StructField.Type.Kind())
		field.DataType = String
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
		case reflect.Struct:
			switch field.StructField.Type.String() {
			case "sql.NullInt16":
				field.Size = 16
			case "sql.NullInt32":
				field.Size = 32
			case "sql.NullInt64", "sql.NullFloat64":
				field.Size = 64
			default:
				log.Printf("No default size for struct type %s, field %s", field.StructField.Type.String(), field.Name)
			}
		}
	}
}
