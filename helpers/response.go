package helpers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Response represents a standard API response structure
type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// LoginResponse represents a login-specific response
type LoginResponse struct {
	Success      bool   `json:"success"`
	Message      string `json:"message"`
	UserID       string `json:"user_id"`
	Token        string `json:"token"`
	RefreshToken string `json:"refresh_token"`
}

// Success sends a successful response
func Success(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// Error sends an error response
func Error(c *gin.Context, statusCode int, message string) {
	c.JSON(statusCode, Response{
		Success: false,
		Error:   message,
	})
}

// LoginSuccess sends a successful login response
func LoginSuccess(c *gin.Context, userID, token, refreshToken string) {
	c.JSON(http.StatusOK, LoginResponse{
		Success:      true,
		Message:      "Logged In Successfully",
		UserID:       userID,
		Token:        token,
		RefreshToken: refreshToken,
	})
}

// BadRequest sends a 400 Bad Request response
func BadRequest(c *gin.Context, message string) {
	Error(c, http.StatusBadRequest, message)
}

// Unauthorized sends a 401 Unauthorized response
func Unauthorized(c *gin.Context, message string) {
	Error(c, http.StatusUnauthorized, message)
}

// NotFound sends a 404 Not Found response
func NotFound(c *gin.Context, message string) {
	Error(c, http.StatusNotFound, message)
}

// InternalServerError sends a 500 Internal Server Error response
func InternalServerError(c *gin.Context, message string) {
	Error(c, http.StatusInternalServerError, message)
}
