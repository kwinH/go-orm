package oorm

import (
	"database/sql"
)

type ITransaction interface {
	Begin() (*sql.Tx, error)
}

func (d *DB) Begin() (*DB, error) {
	var err error
	db := d.ClonePure(0)
	db.Logger.Info("Transaction Begin")
	db.tx, err = db.connPool.(ITransaction).Begin()

	if err != nil {
		db.Logger.Error("Transaction Begin %v", err)
	}

	return db, err
}

func (d *DB) Commit() (err error) {
	d.Logger.Info("Transaction Commit")
	err = d.tx.Commit()

	if err != nil {
		d.Logger.Error("Transaction Commit %v", err)
	}

	return
}

func (d *DB) Rollback() (err error) {
	d.Logger.Info("Transaction Rollback")
	err = d.tx.Rollback()

	if err != nil {
		d.Logger.Error("Transaction Rollback %v", err)
	}

	return
}

type TxFunc func(*DB) error

func (d *DB) Transaction(f TxFunc) (err error) {
	db, err := d.Begin()

	if err != nil {
		return
	}

	defer func() {
		if p := recover(); p != nil {
			_ = db.Rollback()
			panic(p) // re-throw panic after Rollback
		} else if err != nil {
			_ = db.Rollback() // err is non-nil; don't change it
		} else {
			err = db.Commit() // err is nil; if Commit returns error update err
		}
	}()

	err = f(db)
	return
}
