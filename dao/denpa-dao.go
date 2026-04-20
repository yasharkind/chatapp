package dao

import (
	"chatapp/base/appconfig"
	"chatapp/base/db"
	"chatapp/objects"
	"fmt"
)

type DenpaDao interface {
	QueryAll() ([]*objects.Denpa, int, error)
}

type denpaDao struct {
	db.DAO
}

func NewDenpaDao(config *appconfig.Config) DenpaDao {
	_db := db.NewMysql()
	err := _db.Init(config.DataBase.MySql.Url, config.DataBase.MySql.MaxIdleConns, config.DataBase.MySql.MaxOpenConns)
	if err != nil {
		fmt.Println("sql connection error: ", err)
	}
	return &denpaDao{DAO: _db}
}

func (dao *denpaDao) QueryAll() ([]*objects.Denpa, int, error) {
	count, err := dao.DAO.Query("SELECT COUNT(*) FROM denpa")
	if err != nil {
		fmt.Println("Query error: ", err)
		return nil, 0, err
	}

	defer count.Close()
	total := 0
	count.Next()
	count.Scan(&total)

	rows, err := dao.DAO.Query(
		"SELECT * FROM denpa ORDER BY singers",
	)
	if err != nil {
		fmt.Println("Query error: ", err)
		return nil, 0, err
	}
	defer rows.Close()

	var denpa []*objects.Denpa
	for rows.Next(){
		var dnp objects.Denpa
		err := rows.Scan(&dnp.Id, &dnp.Title, &dnp.Thumbnail, &dnp.Url, &dnp.Singers)
		if err != nil {
			fmt.Println("Query error: ", err)
			return nil, 0, err
		}
		denpa = append(denpa, &dnp)
	}
	return denpa, total, nil
}
