package middleware

import (
	"context"
	"net/http"
	"time"

	// "strings"

	"github/akhil/ecommerce-yt/database"
	"github/akhil/ecommerce-yt/models"
	"github/akhil/ecommerce-yt/tokens"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// Authentication middleware validates JWT tokens and sets user info in context
func Authentication() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get token from header
		clientToken := c.GetHeader("token")

		if clientToken == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "No Authorization header provided"})
			c.Abort()
			return
		}

		// Validate token
		claims, err := tokens.ValidateToken(clientToken)
		if err != "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err})
			c.Abort()
			return
		}

		// Set user information in context for use in handlers
		c.Set("email", claims.Email)
		c.Set("first_name", claims.First_Name)
		c.Set("last_name", claims.Last_Name)
		c.Set("user_id", claims.User_ID)

		// Continue to next handler
		c.Next()
	}
}

// AdminAuth middleware checks if the authenticated user is an admin
func AdminAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user_id from context (set by Authentication middleware)
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
			c.Abort()
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Find the user and check if they are admin
		var user models.User
		err := database.UserData(database.Client, "Users").FindOne(ctx, bson.M{"user_id": userID}).Decode(&user)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
				c.Abort()
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error checking admin status"})
			c.Abort()
			return
		}

		// Check if user is admin
		if !user.IsAdmin {
			c.JSON(http.StatusForbidden, gin.H{"error": "access denied: admin privileges required"})
			c.Abort()
			return
		}

		// Continue to next handler
		c.Next()
	}
}
