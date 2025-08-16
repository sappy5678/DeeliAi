package router

import (
	"net/http"

	"github.com/gin-gonic/gin"

	userService "github.com/sappy5678/DeeliAi/internal/app/service/user" // Alias user service
	"github.com/sappy5678/DeeliAi/internal/domain/common"
	"github.com/sappy5678/DeeliAi/internal/domain/user" // Keep domain user
)

type UserHandler struct {
	userSvc userService.Service // Use aliased service
}

func NewUserHandler(userSvc userService.Service) *UserHandler { // Use aliased service
	return &UserHandler{
		userSvc: userSvc,
	}
}

type signUpRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type signUpResponse struct {
	User  *user.User `json:"user"`
	Token string     `json:"token"`
}

func (h *UserHandler) SignUp(c *gin.Context) {
	var req signUpRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondWithError(c, common.NewError(common.ErrorCodeParameterInvalid, err, common.WithMsg("invalid parameter")))
		return
	}

	createdUser, token, cerr := h.userSvc.SignUp(c.Request.Context(), req.Email, req.Username, req.Password)
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

type loginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type loginResponse struct {
	User  *user.User `json:"user"`
	Token string     `json:"token"`
}

func (h *UserHandler) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondWithError(c, common.NewError(common.ErrorCodeParameterInvalid, err, common.WithMsg("invalid parameter")))
		return
	}

	foundUser, token, cerr := h.userSvc.Login(c.Request.Context(), req.Email, req.Password)
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

func (h *UserHandler) GetCurrentUser(c *gin.Context) {
	userID, cerr := GetCurrentUserID(c)
	if cerr != nil {
		respondWithError(c, cerr)
		return
	}

	foundUser, cerr := h.userSvc.GetUser(c.Request.Context(), userID)
	if cerr != nil {
		respondWithError(c, cerr)
		return
	}

	// Remove password hash from response
	foundUser.PasswordHash = ""

	respondWithJSON(c, http.StatusOK, foundUser)
}
