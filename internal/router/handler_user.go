package router

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/sappy5678/DeeliAi/internal/app"
	"github.com/sappy5678/DeeliAi/internal/domain/common"
	"github.com/sappy5678/DeeliAi/internal/domain/user"
)

type signUpRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type signUpResponse struct {
	User  *user.User `json:"user"`
	Token string     `json:"token"`
}

func SignUp(app *app.Application) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req signUpRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			respondWithError(c, common.NewError(common.ErrorCodeParameterInvalid, err, common.WithMsg("invalid parameter")))
			return
		}

		createdUser, token, cerr := app.UserService.SignUp(c.Request.Context(), req.Email, req.Username, req.Password)
		if cerr != nil {
			respondWithError(c, cerr)
			return
		}

		// Remove password hash from response
		createdUser.PasswordHash = ""

		respondWithJSON(c, http.StatusCreated, signUpResponse{
			User:  createdUser,
			Token: token,
		})
	}
}

type loginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type loginResponse struct {
	User  *user.User `json:"user"`
	Token string     `json:"token"`
}

func Login(app *app.Application) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req loginRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			respondWithError(c, common.NewError(common.ErrorCodeParameterInvalid, err, common.WithMsg("invalid parameter")))
			return
		}

		foundUser, token, cerr := app.UserService.Login(c.Request.Context(), req.Email, req.Password)
		if cerr != nil {
			respondWithError(c, cerr)
			return
		}

		// Remove password hash from response
		foundUser.PasswordHash = ""

		respondWithJSON(c, http.StatusOK, loginResponse{
			User:  foundUser,
			Token: token,
		})
	}
}

func GetCurrentUser(app *app.Application) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, cerr := GetCurrentUserID(c)
		if cerr != nil {
			respondWithError(c, cerr)
			return
		}

		foundUser, cerr := app.UserService.GetUser(c.Request.Context(), userID)
		if cerr != nil {
			respondWithError(c, cerr)
			return
		}

		// Remove password hash from response
		foundUser.PasswordHash = ""

		respondWithJSON(c, http.StatusOK, foundUser)
	}
}
