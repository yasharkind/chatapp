package db

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
)

type mySqlDB struct {
	db *sql.DB
}

func NewMysql() *mySqlDB {
	return &mySqlDB{nil}
}

func (db *mySqlDB) Init(url string, MaxIdleCons int, MaxOpenCons int) error {
	var err error
	db.db, err = sql.Open("mysql", url+"?parseTime=true")
	if err != nil {
		return err
	}

	db.db.SetMaxIdleConns(MaxIdleCons)
	db.db.SetMaxOpenConns(MaxOpenCons)
	return nil
}

func (db *mySqlDB) Close() {
	if db.db != nil {
		db.db.Close()
	}
}

func (db *mySqlDB) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return db.db.Query(query, args...)
}

func (db *mySqlDB) Exec(query string, args ...interface{}) (sql.Result, error) {
	return db.db.Exec(query, args...)
}

func (db *mySqlDB) BeginTX () (*sql.Tx, error) {
	return db.db.Begin()
}


