package schema

import (
	"database/sql"
	"oorm/drive"
)

type IDialect interface {
	Name() string
	GetDSN() string
	DataTypeOf(*Field) string
	Init() (drive.IConnPool, error)
	Migrate(IDBParse) IMigrator
}

type ITableName interface {
	TableName() string
}

type TableInfo struct {
	FieldsInfo map[string]string
	PrimaryKey string
	UniqueKeys map[string][]string
	IndexKeys  map[string][]string
	FullKeys   map[string][]string
}

type IMigrator interface {
	TableExist(tableName string) bool
	Create(schema1 *Schema) error
	TableInfo(tableName string) TableInfo

	AddField(TableName string, field *Field) error
	ModifyField(TableName string, field *Field) error
	DropField(TableName string, FiledName string) error

	AddIndex(tableName string, indexType IndexType, indexFields IndexList) error
	DropIndex(indexKey, tableName string) error
	UpdateIndex(schema1 *Schema, schemaKeys IndexList, keys map[string][]string, modify bool, indexType IndexType) (err error)

	Auto(value interface{}, modify, drop bool) error
}

type IDBParse interface {
	Query(string, ...interface{}) (*sql.Rows, error)
	Exec(string, ...interface{}) (sql.Result, error)
	Parse(interface{}) *Schema
}
