package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/muhali16/listmak-service/internal/services"
	"github.com/muhali16/listmak-service/pkg/utils"
)

type ConfigController interface {
	GetPublicConfig(c *gin.Context) // public: frontend reads testing_mode for banner + mode
	UpdateConfig(c *gin.Context)    // admin: toggle testing_mode
}

type configController struct {
	appConfig services.AppConfig
}

func NewConfigController(appConfig services.AppConfig) ConfigController {
	return &configController{appConfig: appConfig}
}

// GetPublicConfig godoc
// @Summary  Public runtime config (testing mode)
// @Tags     config
// @Produce  json
// @Success  200  {object}  utils.Response
// @Router   /config [get]
func (cc *configController) GetPublicConfig(c *gin.Context) {
	utils.SendResponse(c, http.StatusOK, true, "OK", gin.H{
		"testing_mode": cc.appConfig.TestingMode(),
	})
}

// UpdateConfig godoc
// @Summary  Toggle testing mode (admin)
// @Tags     admin
// @Accept   json
// @Produce  json
// @Param    payload  body      map[string]interface{}  true  "Config"
// @Success  200      {object}  utils.Response
// @Router   /admin/config [put]
func (cc *configController) UpdateConfig(c *gin.Context) {
	var body struct {
		TestingMode *bool `json:"testing_mode"`
	}
	if err := c.ShouldBindJSON(&body); err != nil || body.TestingMode == nil {
		utils.SendResponse(c, http.StatusBadRequest, false, "Invalid payload", nil)
		return
	}
	if err := cc.appConfig.SetTestingMode(*body.TestingMode); err != nil {
		utils.SendResponse(c, http.StatusInternalServerError, false, "Gagal menyimpan konfigurasi", nil)
		return
	}
	utils.SendResponse(c, http.StatusOK, true, "Konfigurasi disimpan", gin.H{
		"testing_mode": cc.appConfig.TestingMode(),
	})
}
