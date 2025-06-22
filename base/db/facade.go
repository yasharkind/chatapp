package db

import (
	"database/sql"
	"log"
	"chatapp/base/appconfig"
)

var (
	msql *mySqlDB
)

type DAO interface {
	Query(query string, args ...interface{}) (*sql.Rows, error)
	Exec(query string, args ...interface{}) (sql.Result, error)
	BeginTX() (*sql.Tx, error)
}

func Init(config *appconfig.Config) {
	msql = NewMysql()

	msqlCfg := config.DataBase.MySql
	err := msql.Init(msqlCfg.Url, msqlCfg.MaxIdleConns,msqlCfg.MaxOpenConns)
	if err != nil {
		log.Panicf("Mysql init failed: %s", err.Error())
	}
}

func Close() {
	msql.Close()
}

func MySqlDao() DAO {
	return msql
}
