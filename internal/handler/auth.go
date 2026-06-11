// Package handler implements HTTP request handlers for the API endpoints.
package handler

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/golang-api-server/internal/domain"
	"github.com/golang-api-server/internal/model"
	"github.com/golang-api-server/internal/service"
)

// AuthHandler handles authentication-related HTTP requests.
type AuthHandler struct {
	authService service.AuthService
}

// NewAuthHandler returns a new AuthHandler.
func NewAuthHandler(authService service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req model.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request", "details": err.Error()})
		return
	}

	resp, err := h.authService.Register(c.Request.Context(), &req)
	if err != nil {
		if errors.Is(err, service.ErrEmailTaken) {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to register"})
		return
	}

	c.JSON(http.StatusCreated, resp)
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req model.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request", "details": err.Error()})
		return
	}

	resp, err := h.authService.Login(c.Request.Context(), &req)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to login"})
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req model.RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request", "details": err.Error()})
		return
	}

	resp, err := h.authService.RefreshToken(c.Request.Context(), req.RefreshToken)
	if err != nil {
		if errors.Is(err, service.ErrInvalidRefreshToken) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to refresh token"})
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *AuthHandler) Logout(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "logged out"})
}

func (h *AuthHandler) UpdateProfile(c *gin.Context) {
	userID := GetUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req model.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request", "details": err.Error()})
		return
	}

	user, err := h.authService.UpdateProfile(c.Request.Context(), userID, &req)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update profile"})
		return
	}

	c.JSON(http.StatusOK, user)
}

func (h *AuthHandler) GetProfile(c *gin.Context) {
	userID := GetUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	user, err := h.authService.GetProfile(c.Request.Context(), userID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get profile"})
		return
	}

	c.JSON(http.StatusOK, user)
}

// UserHandler handles user management HTTP requests.
type UserHandler struct {
	userService service.UserService
}

// NewUserHandler returns a new UserHandler.
func NewUserHandler(userService service.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

func (h *UserHandler) CheckAvailability(c *gin.Context) {
	var req model.CheckAvailabilityRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid query params", "details": err.Error()})
		return
	}
	emailAvail, usernameAvail, err := h.userService.CheckAvailability(c.Request.Context(), req.Email, req.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "availability check failed"})
		return
	}
	c.JSON(http.StatusOK, model.CheckAvailabilityResponse{
		EmailAvailable:    emailAvail,
		UsernameAvailable: usernameAvail,
	})
}

func (h *UserHandler) List(c *gin.Context) {
	users, err := h.userService.List(c.Request.Context(), 0, 100)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list users"})
		return
	}
	c.JSON(http.StatusOK, users)
}

func (h *UserHandler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	var id int64
	if _, err := fmt.Sscanf(idStr, "%d", &id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	if err := h.userService.Delete(c.Request.Context(), id); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "user deleted"})
}

func GetUserID(c *gin.Context) int64 {
	val, exists := c.Get("user_id")
	if !exists {
		return 0
	}
	id, ok := val.(int64)
	if !ok {
		return 0
	}
	return id
}

func GetUserRole(c *gin.Context) string {
	val, exists := c.Get("role")
	if !exists {
		return ""
	}
	role, ok := val.(string)
	if !ok {
		return ""
	}
	return role
}
