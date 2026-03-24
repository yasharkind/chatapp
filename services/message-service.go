package service

import (
	"chatapp/base/appconfig"
	"chatapp/dao"
	"chatapp/objects"
	"fmt"
)

type MessageService interface {
	Save(objects.Message) *objects.Message
	FindAll(offset int,limit int) []*objects.Message
	FindByLimitOffset(limit int, offset int) []*objects.Message
	FindById(id int) (*objects.Message, error)
	UpdateById(int, objects.Message) (*objects.Message, error)
	DeleteById(int) error
}

type messageService struct {
	messagedao dao.MessageDao
	config *appconfig.Config
}

func (service *messageService) Save(msg objects.Message) *objects.Message {
	id, err := service.messagedao.Create(msg)
	if err != nil {
		fmt.Println("create error: ", err)
		return nil
	}
	msg.Id = id
	return &msg
}

func (service *messageService) FindAll(offset int, limit int) []*objects.Message {
	msgs,_,err := service.messagedao.QueryAll(offset, limit)
	if err != nil {
		fmt.Println("queryall error: ", err)
	}
	for _, msg := range msgs {
		fmt.Println(msg.Content)
	}
	return msgs
}

func (service *messageService) FindByLimitOffset(limit int, offset int) []*objects.Message {
	msgs,_,err := service.messagedao.QueryByLimitOffset(limit, offset)
	if err != nil {
		fmt.Println("queryall error: ", err)
	}
	return msgs
}

func (service *messageService) FindById(id int) (*objects.Message, error) {
	msg, err := service.messagedao.FindById(id)
	if err != nil {
		fmt.Println("FindById error: ", err)
		return nil, err 
	}

	return msg, nil 
}

func (service *messageService) UpdateById(id int, msg objects.Message) (*objects.Message, error) {
	newMsg, err := service.messagedao.UpdateById(id, msg)
	if err != nil {
		fmt.Println("UpdateById error: ", err)
		return nil, err 
	}

	return newMsg, nil 
}


func (service *messageService) DeleteById(id int) error {
	err := service.messagedao.DeleteById(id)
	if err != nil {
		fmt.Println("DeleteById error: ", err)
	}
	return nil
}

func NewMessageService(config *appconfig.Config) MessageService {
	return &messageService{messagedao : dao.NewMessageDao(config), config: config}
}
