package dao

import (
	"chatapp/base/appconfig"
	"chatapp/base/db"
	"chatapp/objects"
	"errors"
	"fmt"
	"log"
)

type UserDao interface {
	Create(user objects.User) (int, error)
	DeleteById(id int) error
	QueryAll(offset int, limit int) ([]*objects.User, int, error)
	FindById(id int) (*objects.User, error)
	UpdateById(id int, user objects.User) (*objects.User, error)
	FindByUsernameAndPassword(username string, password string) (*objects.User, error)
}

type userDao struct {
	db.DAO
}

func NewUserDao(config *appconfig.Config) UserDao {
	_db := db.NewMysql()
	err := _db.Init(config.DataBase.MySql.Url, config.DataBase.MySql.MaxIdleConns, config.DataBase.MySql.MaxOpenConns)
	if err != nil {
		fmt.Println("sql connection error: ", err)
	}
	return &userDao{DAO: _db}
}

func (dao *userDao) Create(user objects.User) (int, error) {
	result, err := dao.DAO.Exec(
		"INSERT INTO user (username, password, avatar) VALUES (?,?,?)",
		user.Username, user.Password, user.Avatar,
	)
	if err != nil {
		fmt.Println("create error: ", err)
		return -1, err
	}

	id, _ := result.LastInsertId()

	return int(id), nil
}

func (dao *userDao) DeleteById(id int) error {
	result, err := dao.DAO.Exec(
		"DELETE FROM user WHERE ID = ?", id,
	)
	if err != nil {
		fmt.Println("delete error: ", err)
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		fmt.Println("delete error: ", id, " not found")
	}
	return nil
}

func (dao *userDao) QueryAll(offset int, limit int) ([]*objects.User, int, error) {
	count, err := dao.DAO.Query("SELECT COUNT(*) FROM user")
	if err != nil {
		fmt.Println("Query error: ", err)
		return nil, 0, err
	}

	defer count.Close()
	total := 0
	count.Next()
	count.Scan(&total)

	rows, err := dao.DAO.Query(
		"SELECT * FROM user LIMIT ? OFFSET ?",
		limit, offset,
	)
	if err != nil {
		fmt.Println("Query error: ", err)
		return nil, 0, err
	}
	defer rows.Close()

	var users []*objects.User
	for rows.Next(){
		var user objects.User
		err := rows.Scan(&user.Id, &user.Username, &user.Password, &user.Avatar)
		if err != nil {
			fmt.Println("Query error: ", err)
			return nil, 0, err
		}
		users= append(users, &user)
	}
	return users, total, nil
}

func (dao *userDao) FindById(id int) (*objects.User, error) {
	rows, err := dao.DAO.Query("SELECT * FROM user WHERE ID = ?", id)
	if err != nil {
		log.Println("Query error: ", err)
		return nil, err
	}
	defer rows.Close()

	exists := rows.Next()
	if !exists {
		log.Println("Query error: ", id, " not found")
		err := errors.New("404 not found")
		return nil, err
	}
	var user objects.User
	err = rows.Scan(&user.Id, &user.Username, &user.Password, &user.Avatar)
	return &user, nil
}

func  (dao *userDao) UpdateById(id int, user objects.User) (*objects.User, error) {
	result, err := dao.DAO.Exec("UPDATE user SET (username = ?, password = ?) WHERE ID = ?", user.Username, user.Password, id)
	if err != nil {
		log.Println("update error: ", err)
	}
	rows,_ := result.RowsAffected()
	if rows == 0 {
		log.Println("update error: ", id, " not found")
	}

	return &user, nil
}

func (dao *userDao) FindByUsernameAndPassword(username string, password string) (*objects.User, error) {
	rows, err := dao.DAO.Query("SELECT * FROM user WHERE username = ? and password = ?", username, password)
	if err != nil {
		fmt.Println("FindByUsernameAndPassword error: ", err)
		return nil, err
	}

	defer rows.Close()

	exists := rows.Next()
	if !exists {
		log.Println("Query error: ", username, " with ", password, " does not exist!")
		err := errors.New("404 not found")
		return nil, err
	}
	var user objects.User
	err = rows.Scan(&user.Id, &user.Username, &user.Password, &user.Avatar)
	if err != nil {
		fmt.Println("scan error: ", err)
	}
	return &user, nil
}
