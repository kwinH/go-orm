package oorm

import (
	"database/sql"
	"fmt"
	"github.com/kwinH/go-oorm/drive"
	"github.com/kwinH/go-oorm/logger"
	"github.com/kwinH/go-oorm/schema"
	"github.com/kwinH/go-sql-builder"
	"time"
)

type Orm struct {
	*Config
}

type Config struct {
	TablePrefix string
	connPool    drive.IConnPool
	dialector   schema.IDialect
	Migrate     schema.IMigrator
	Logger      logger.ILogger
}

type DB struct {
	*Config
	tx *sql.Tx

	omitField  map[string]bool
	b          sqlBuilder.Builder
	schema     *schema.Schema
	withs      map[string]WithFunc
	childWiths map[string][]WithFunc
	clone      int
	Error      error
	sql        string
	bindings   []interface{}
	withDel    bool
	omitEmpty  bool
	startTime  time.Time
	tableAlias string
}

func Open(dialector schema.IDialect, c ...*Config) (orm *Orm, err error) {
	var config *Config
	if len(c) == 1 {
		config = c[0]
	} else {
		config = &Config{}
	}

	config.dialector = dialector

	if config.Logger == nil {
		config.Logger = logger.Logger{
			LogLevel: logger.Trace,
		}
	}

	orm = &Orm{
		Config: config,
	}

	if config.dialector != nil {
		config.connPool, err = config.dialector.Init()
		config.Migrate = config.dialector.Migrate(orm.NewDB())
	}

	return
}

// AddError add error to db
func (d *DB) AddError(err error) error {
	if d.Error == nil {
		d.Error = err
	} else if err != nil {
		d.Error = fmt.Errorf("%v; %w", d.Error, err)
	}
	return d.Error
}

func (d *DB) Exec(query string, args ...interface{}) (res sql.Result, err error) {
	db := d.getInstance()

	var stmt *sql.Stmt
	if db.tx != nil {
		stmt, err = db.tx.Prepare(query)
	} else {
		stmt, err = db.connPool.Prepare(query)
	}

	if err != nil {
		fmt.Printf("prepare failed, err:%v\n", err)
		return
	}
	defer stmt.Close()

	res, err = stmt.Exec(args...)

	db.Logger.Trace(query, args, db.startTime)
	return
}

func (d *DB) Query(query string, args ...interface{}) (res *sql.Rows, err error) {
	db := d.getInstance()

	stmt, err := db.connPool.Prepare(query)
	if err != nil {
		return
	}

	res, err = stmt.Query(args...)

	db.Logger.Trace(query, args, db.startTime)

	return
}

func (d *DB) resetClone() {
	d.clone = 0
}

func (d *DB) Clone() *DB {
	db := &DB{
		Config: d.Config,
		tx:     d.tx,

		b:         d.b,
		schema:    d.schema,
		clone:     d.clone,
		withDel:   d.withDel,
		omitEmpty: d.omitEmpty,
	}

	db.withs = make(map[string]WithFunc)
	db.childWiths = make(map[string][]WithFunc)

	if d.withs != nil {
		for s, withFunc := range d.withs {
			db.withs[s] = withFunc
		}
	}

	if db.childWiths != nil {
		for s, withFuncs := range d.childWiths {
			db.childWiths[s] = withFuncs
		}
	}

	return db
}

func (d *DB) ClonePure(clones ...int) *DB {
	clone := d.clone
	if len(clones) == 1 {
		clone = clones[0]
	}
	db := &DB{
		Config:    d.Config,
		tx:        d.tx,
		clone:     clone,
		withDel:   d.withDel,
		omitEmpty: d.omitEmpty,
	}

	if clone == 1 {
		db.startTime = time.Now()
	}

	return db
}

func (d *DB) getInstance() *DB {
	if d.clone == 0 {
		return d.ClonePure(1)
	}
	return d
}

func (o *Orm) DBPool() *sql.DB {
	return o.connPool.(*sql.DB)
}

func (o *Orm) NewDB() (db *DB) {
	db = &DB{
		Config: o.Config,
	}

	return
}
