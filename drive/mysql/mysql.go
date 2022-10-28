package mysql

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"math"
	"oorm/drive"
	"oorm/drive/mysql/migrator"
	"oorm/schema"
)

type Config struct {
	DriverName string
	DSN        string
	Conn       drive.IConnPool
}

type Dialect struct {
	*Config
}

var _ schema.IDialect = (*Dialect)(nil)

func (dialect *Dialect) Name() string {
	return "mysql"
}

func (dialect *Dialect) GetDSN() string {
	return dialect.Config.DSN
}

func (dialect *Dialect) DataTypeOf(field *schema.Field) string {
	var fieldType string
	switch field.DataType {
	case schema.Bool:
		fieldType = "tinyint"
	case schema.String:
		if field.Size == 0 {
			field.Size = 255
		}
		if field.Size > 0 && field.Size < 65536 {
			return fmt.Sprintf("varchar(%v)", field.Size)
		} else if field.Size >= 65536 && field.Size <= int(math.Pow(2, 24)) {
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
	case schema.Float:
		if field.Decimal != "" {
			fieldType = fmt.Sprintf("decimal(%s)", field.Decimal)
		} else if field.Size <= 32 {
			fieldType = "float"
		} else {
			fieldType = "double"
		}
	//case schema.Bytes:
	//	if field.Size > 0 && field.Size < 65536 {
	//		fieldType = fmt.Sprintf("varbinary(%d)", field.Size)
	//	} else if field.Size >= 65536 && field.Size <= int(math.Pow(2, 24)) {
	//		fieldType = "mediumblob"
	//	} else {
	//		fieldType = "longblob"
	//	}
	case schema.Time:
		fieldType = "datetime(3)"
	}

	return fieldType
}

func (dialect *Dialect) Init() (connPool drive.IConnPool, err error) {

	if dialect.DriverName == "" {
		dialect.DriverName = "mysql"
	}

	if dialect.Conn == nil {
		db, err := sql.Open(dialect.DriverName, dialect.DSN)

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
