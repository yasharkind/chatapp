package service

import (
	"chatapp/base/appconfig"
	"chatapp/dao"
	"chatapp/objects"
	"fmt"
)

type DenpaService interface {
	FindAll() []*objects.Denpa
}

type denpaService struct {
	denpaDao dao.DenpaDao
	config *appconfig.Config
}

func NewDenpaService(config *appconfig.Config) DenpaService {
	return &denpaService{denpaDao : dao.NewDenpaDao(config), config: config}
}

func (service *denpaService) FindAll() []*objects.Denpa {
	denpas,_,err := service.denpaDao.QueryAll()
	if err != nil {
		fmt.Println("queryall error: ", err)
	}
	return denpas
}

