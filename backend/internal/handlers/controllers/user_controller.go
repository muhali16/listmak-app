package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/muhali16/listmak-service/internal/models"
	"github.com/muhali16/listmak-service/internal/services"
	"github.com/muhali16/listmak-service/pkg/utils"
)

type UserController interface {
	GetUsers(c *gin.Context)
	CreateUser(c *gin.Context)
}

type userController struct {
	UserService services.UserService
}

func NewUserController(userService services.UserService) UserController {
	return &userController{UserService: userService}
}

// GetUsers godoc
// @Summary      Get all users
// @Description  Mengambil daftar semua user yang terdaftar
// @Tags         users
// @Accept       json
// @Produce      json
// @Success      200  {object}  utils.Response{data=[]models.User}
// @Failure      500  {object}  utils.Response
// @Router       /users [get]
func (uc *userController) GetUsers(c *gin.Context) {
	data, err := uc.UserService.GetAllUsers()
	if err != nil {
		utils.SendResponse(c, http.StatusInternalServerError, false, "Failed to get users", nil)
		return
	}

	utils.SendResponse(c, http.StatusOK, true, "Success get users", data)
}

// CreateUser godoc
// @Summary      Create new user
// @Description  Membuat user baru
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        user  body     models.User  true  "User object"
// @Success      200   {object}  utils.Response{data=models.User}
// @Failure      500   {object}  utils.Response
// @Router       /users [post]
func (uc *userController) CreateUser(c *gin.Context) {
	var user models.User
	if err := c.ShouldBindJSON(&user); err != nil {
		utils.SendResponse(c, http.StatusBadRequest, false, "Invalid request payload", nil)
		return
	}

	createdUser, err := uc.UserService.CreateUser(user)
	if err != nil {
		utils.SendResponse(c, http.StatusInternalServerError, false, "Failed to create user", nil)
		return
	}

	utils.SendResponse(c, http.StatusOK, true, "Success create user", createdUser)
}
