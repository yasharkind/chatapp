package controller

import (
	"chatapp/base/appconfig"
	"chatapp/middlewares"
	service "chatapp/services"
	"encoding/json"
	"net/http"
)

type DenpaController struct {
	denpaService service.DenpaService
}

func (c *DenpaController) handleDenpa(w http.ResponseWriter, _ *http.Request) {
	denpas := c.denpaService.FindAll()
	marshaled, err := json.Marshal(denpas)
	if err != nil {
		println("handleDenpa, Marshal error")
	}

	w.Write(marshaled)
}

func NewDenpaController(config *appconfig.Config, userService service.UserService) *DenpaController {
	c := DenpaController{
		denpaService: service.NewDenpaService(config),
	}

	http.Handle("/denpa",  middlewares.CorsMiddleware(http.HandlerFunc(c.handleDenpa)))

	return &c
}
