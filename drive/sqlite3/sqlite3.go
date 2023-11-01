package sqlite3

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"math"
	"oorm/drive"
	"oorm/drive/mysql/migrator"
	"oorm/schema"
)

const DriverName = "sqlite3"

type Config struct {
	DSN  string
	Conn drive.IConnPool
}

type Dialect struct {
	*Config
}

var _ schema.IDialect = (*Dialect)(nil)

func (dialect *Dialect) Name() string {
	return "sqlite"
}

func (dialect *Dialect) GetDSN() string {
	return dialect.Config.DSN
}

func (dialect *Dialect) DataTypeOf(field *schema.Field) (fieldType string) {
	switch field.DataType {
	case schema.Bool:
		return "tinyint"
	case schema.String:
		if field.Size == 0 {
			field.Size = 255
		}
		if field.Size > 0 && field.Size < 65536 {
			return fmt.Sprintf("varchar(%d)", field.Size)
		} else if field.Size >= 65536 && field.Size <= int64(math.Pow(2, 24)) {
			return "mediumtext"
		} else {
			return "longtext"
		}
	case schema.Int, schema.Uint:
		if field.Size <= 8 {
			fieldType = "tinyint"
		} else if field.Size <= 16 {
			fieldType = "smallint"
		} else if field.Size <= 32 {
			fieldType = "int"
		} else {
			fieldType = "bigint"
		}
		if field.DataType == schema.Uint {
			fieldType += " unsigned"
		}
		return fieldType
	case schema.Float:
		if field.Decimal != "" {
			return fmt.Sprintf("decimal(%s)", field.Decimal)
		} else if field.Size <= 32 {
			return "float"
		} else {
			return "double"
		}
	case schema.Time:
		if field.Size == 0 {
			field.Size = 3
		}
		return fmt.Sprintf("datetime(%d)", field.Size)
	}

	return
}

func (dialect *Dialect) Init() (connPool drive.IConnPool, err error) {
	if dialect.Conn == nil {
		db, err := sql.Open(DriverName, dialect.DSN)

		if err != nil {
			return db, err
		}
		err = db.Ping()
		if err != nil {
			return db, err
		}
		dialect.Conn = db
	}

	connPool = dialect.Conn

	return
}

func Open(dsn string) *Dialect {
	return &Dialect{Config: &Config{DSN: dsn}}
}

func (dialect *Dialect) Migrate(d schema.IDBParse) schema.IMigrator {
	migrate := &migrator.Migrator{
		DB: d,
	}

	return migrate
}
