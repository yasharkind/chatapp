package dao

import (
	"chatapp/base/appconfig"
	"chatapp/base/db"
	"chatapp/objects"
	"errors"
	"fmt"
	"log"
)

type MessageDao interface {
	Create(message objects.Message) (int, error)
	DeleteById(id int) error
	QueryAll(offset int, limit int) ([]*objects.Message, int, error)
	QueryFromEnd(limit int) ([]*objects.Message, int, error)
	FindById(id int) (*objects.Message, error)
	UpdateById(id int, msg objects.Message) (*objects.Message, error)
}

type messageDao struct {
	db.DAO
}

func NewMessageDao(config *appconfig.Config) MessageDao {
	_db := db.NewMysql()
	err := _db.Init(config.DataBase.MySql.Url, config.DataBase.MySql.MaxIdleConns, config.DataBase.MySql.MaxOpenConns)
	if err != nil {
		fmt.Println("sql connection error: ", err)
	}
	return &messageDao{DAO: _db}
}

func (dao *messageDao) Create(message objects.Message) (int, error) {
	result, err := dao.DAO.Exec(
		"INSERT INTO message (sender, timestamp, content) VALUES (?,?,?)",
		message.Sender, message.Time, message.Content,
	)
	if err != nil {
		fmt.Println("create error: ", err)
		return -1, err
	}

	id, _ := result.LastInsertId()

	return int(id), nil
}

func (dao *messageDao) DeleteById(id int) error {
	result, err := dao.DAO.Exec(
		"DELETE FROM message WHERE id = ?", id,
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

func (dao *messageDao) QueryAll(offset int, limit int) ([]*objects.Message, int, error) {
	count, err := dao.DAO.Query("SELECT COUNT(*) FROM message")
	if err != nil {
		fmt.Println("Query error: ", err)
		return nil, 0, err
	}

	defer count.Close()
	total := 0
	count.Next()
	count.Scan(&total)

	rows, err := dao.DAO.Query(
		"SELECT * FROM message ORDER BY message.ID DESC LIMIT ? OFFSET ?",
		limit, offset,
	)
	if err != nil {
		fmt.Println("Query error: ", err)
		return nil, 0, err
	}
	defer rows.Close()

	var messages []*objects.Message
	for rows.Next(){
		var msg objects.Message
		err := rows.Scan(&msg.Id, &msg.Sender, &msg.Time, &msg.Content)
		if err != nil {
			fmt.Println("Query error: ", err)
			return nil, 0, err
		}
		messages = append(messages, &msg)
	}
	return messages, total, nil
}
func (dao *messageDao) QueryFromEnd(limit int) ([]*objects.Message, int, error) {
	count, err := dao.DAO.Query("SELECT COUNT(*) FROM message")
	if err != nil {
		fmt.Println("Query error: ", err)
		return nil, 0, err
	}

	defer count.Close()
	total := 0
	count.Next()
	count.Scan(&total)

	offset := 0
	if (limit <= total) {offset = total - limit}
	rows, err := dao.DAO.Query(
		"SELECT * FROM message LIMIT ? OFFSET ?",
		limit, offset,
	)
	if err != nil {
		fmt.Println("Query error: ", err)
		return nil, 0, err
	}
	defer rows.Close()

	var messages []*objects.Message
	for rows.Next(){
		var msg objects.Message
		err := rows.Scan(&msg.Id, &msg.Sender, &msg.Time, &msg.Content)
		if err != nil {
			fmt.Println("Query error: ", err)
			return nil, 0, err
		}
		messages = append(messages, &msg)
	}
	return messages, total, nil
}

func (dao *messageDao) FindById(id int) (*objects.Message, error) {
	rows, err := dao.DAO.Query("SELECT * FROM message WHERE id = ?", id)
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
	var msg objects.Message
	err = rows.Scan(&msg.Id, &msg.Sender, &msg.Content, &msg.Time)
	return &msg, nil
}

func  (dao *messageDao) UpdateById(id int, msg objects.Message) (*objects.Message, error) {
	result, err := dao.DAO.Exec("UPDATE message SET content = ? WHERE id = ?", msg.Content, id)
	if err != nil {
		log.Println("update error: ", err)
	}
	rows,_ := result.RowsAffected()
	if rows == 0 {
		log.Println("update error: ", id, " not found")
	}

	return &msg, nil
}
