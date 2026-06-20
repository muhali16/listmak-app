package controllers

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/muhali16/listmak-service/internal/models"
	"github.com/muhali16/listmak-service/internal/services"
	"github.com/muhali16/listmak-service/pkg/utils"
	"golang.org/x/oauth2"
)

type AuthController interface {
	GoogleLogin(c *gin.Context)
	GoogleCallback(c *gin.Context)
	GetUserAuthenticated(c *gin.Context)
	Logout(c *gin.Context)

	getGoogleConfig() *oauth2.Config
}

type authController struct {
	UserService services.UserService
}

func NewAuthController(userService services.UserService) AuthController {
	return &authController{UserService: userService}
}

func (ac *authController) getGoogleConfig() *oauth2.Config {
	return &oauth2.Config{
		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		RedirectURL:  os.Getenv("GOOGLE_REDIRECT_URL"),
		Scopes:       []string{"email", "profile"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://accounts.google.com/o/oauth2/auth",
			TokenURL: "https://accounts.google.com/o/oauth2/token",
		},
	}
}

func isAdminEmail(email string) bool {
	adminEmails := os.Getenv("ADMIN_EMAILS")
	if adminEmails == "" {
		return false
	}
	for _, e := range strings.Split(adminEmails, ",") {
		if strings.EqualFold(strings.TrimSpace(e), email) {
			return true
		}
	}
	return false
}

// LoginGoogle godoc
// @Summary      Redirect to Google Login
// @Tags         auth
// @Success      307
// @Router       /auth/google/login [get]
func (ac *authController) GoogleLogin(c *gin.Context) {
	url := ac.getGoogleConfig().AuthCodeURL("random-state")
	c.Redirect(http.StatusTemporaryRedirect, url)
}

// GoogleCallback godoc
// @Summary      Google callback
// @Description  Google callback
// @Tags         auth
// @Accept       json
// @Produce      json
// @Success      200  {object}  utils.Response{data=models.User}
// @Failure      500  {object}  utils.Response
// @Router       /auth/google/callback [get]
func (ac *authController) GoogleCallback(c *gin.Context) {
	config := ac.getGoogleConfig()
	code := c.Query("code")

	// exchange code
	token, err := config.Exchange(context.Background(), code)
	if err != nil {
		utils.SendResponse(c, http.StatusInternalServerError, false, "Failed to exchange code", nil)
		return
	}

	//get user info
	response, err := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + token.AccessToken)
	if err != nil {
		utils.SendResponse(c, http.StatusInternalServerError, false, "Failed to get user info", nil)
		return
	}

	defer response.Body.Close()

	var googleUser struct {
		GoogleID string `json:"id"`
		Email    string `json:"email"`
		Name     string `json:"name"`
		Picture  string `json:"picture"`
	}
	json.NewDecoder(response.Body).Decode(&googleUser)

	user, err := ac.UserService.GetUserByGoogleId(googleUser.GoogleID)
	if err != nil {
		role := "user"
		if isAdminEmail(googleUser.Email) {
			role = "admin"
		}
		user = models.User{
			GoogleID: googleUser.GoogleID,
			Email:    googleUser.Email,
			Name:     googleUser.Name,
			Avatar:   googleUser.Picture,
			Role:     role,
		}
		user, err = ac.UserService.CreateUser(user)
		if err != nil {
			utils.SendResponse(c, http.StatusInternalServerError, false, "Failed to create user", nil)
			return
		}
	} else {
		// Existing user: enforce ADMIN_EMAILS promotion on every login
		if isAdminEmail(user.Email) && user.Role != "admin" {
			ac.UserService.UpdateRole(user.ID, "admin")
			user.Role = "admin"
		}
	}

	jwtToken, err := utils.GenerateJWT(strconv.FormatUint(uint64(user.ID), 10), user.Email, user.Role)
	if err != nil {
		utils.SendResponse(c, http.StatusInternalServerError, false, "Failed to generate token", nil)
		return
	}

	c.SetCookie(
		"X-User-Authentication-Token",
		jwtToken,
		60*60*24,
		"/",
		os.Getenv("FRONTEND_DOMAIN"),
		false,
		true,
	)

	// utils.SendResponse(c, http.StatusOK, true, "Logged in successfully", gin.H{
	// 	"token": jwtToken,
	// 	"user":  user,
	// })
	c.Redirect(http.StatusPermanentRedirect, os.Getenv("FRONTEND_URL"))
}

// GetUserAuthenticated godoc
// @Summary      Get user authenticated
// @Tags         auth
// @Accept       json
// @Produce      json
// @Success      200  {object}  utils.Response{data=models.User}
// @Failure      500  {object}  utils.Response
// @Router       /auth/user [get]
func (ac *authController) GetUserAuthenticated(c *gin.Context) {
	userIdStr := c.MustGet("user_id").(string)
	userId, err := strconv.ParseUint(userIdStr, 10, 32)
	if err != nil {
		utils.SendResponse(c, http.StatusBadRequest, false, "Invalid user ID", nil)
		return
	}
	user, err := ac.UserService.GetUserById(uint(userId))
	if err != nil {
		utils.SendResponse(c, http.StatusInternalServerError, false, "Failed to get user", nil)
		return
	}
	utils.SendResponse(c, http.StatusOK, true, "User authenticated", user)
}

// Logout godoc
// @Summary      Logout
// @Tags         auth
// @Accept       json
// @Produce      json
// @Success      200  {object}  utils.Response
// @Failure      500  {object}  utils.Response
// @Router       /auth/logout [get]
func (ac *authController) Logout(c *gin.Context) {
	c.SetCookie("X-User-Authentication-Token", "", -1, "/", os.Getenv("FRONTEND_DOMAIN"), false, true)
	utils.SendResponse(c, http.StatusOK, true, "Logged out successfully", nil)
}
