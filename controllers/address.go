package controllers

import (
	"net/http"

	"github/akhil/ecommerce-yt/database"
	"github/akhil/ecommerce-yt/models"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// AddAddress adds a new address to the user's address list
func (app *Application) AddAddress() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user_id from context (set by middleware)
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
			return
		}

		// Bind address from request body
		var address models.Address
		if err := c.BindJSON(&address); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Call database function
		addressID, err := database.AddAddress(app.UserCollection, userID.(string), address)
		if err != nil {
			if err == database.ErrCantFindUser {
				c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message":    "address added successfully",
			"address_id": addressID,
		})
	}
}

// EditHomeAddress updates the home address (first address in the list)
func (app *Application) EditHomeAddress() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user_id from context (set by middleware)
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
			return
		}

		// Bind address from request body
		var address models.Address
		if err := c.BindJSON(&address); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Call database function
		addressID, err := database.EditHomeAddress(app.UserCollection, userID.(string), address)
		if err != nil {
			if err == database.ErrCantFindUser {
				c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
				return
			}
			if err == database.ErrNoAddresses {
				c.JSON(http.StatusBadRequest, gin.H{"error": "no addresses found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message":    "home address updated successfully",
			"address_id": addressID,
		})
	}
}

// EditWorkAddress updates the work address (second address in the list, or creates if doesn't exist)
func (app *Application) EditWorkAddress() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user_id from context (set by middleware)
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
			return
		}

		// Bind address from request body
		var address models.Address
		if err := c.BindJSON(&address); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Call database function
		addressID, err := database.EditWorkAddress(app.UserCollection, userID.(string), address)
		if err != nil {
			if err == database.ErrCantFindUser {
				c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message":    "work address updated successfully",
			"address_id": addressID,
		})
	}
}

// GetAddresses retrieves all addresses for the authenticated user
func (app *Application) GetAddresses() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user_id from context (set by middleware)
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
			return
		}

		// Call database function
		addresses, err := database.GetAddresses(app.UserCollection, userID.(string))
		if err != nil {
			if err == database.ErrCantFindUser {
				c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    addresses,
			"count":   len(addresses),
		})
	}
}

// DeleteAddress deletes an address by address_id
func (app *Application) DeleteAddress() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user_id from context (set by middleware)
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
			return
		}

		// Get address ID from query parameter
		addressQueryID := c.Query("id")
		if addressQueryID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "address id is required"})
			return
		}

		addressID, err := primitive.ObjectIDFromHex(addressQueryID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid address id"})
			return
		}

		// Call database function
		err = database.DeleteAddress(app.UserCollection, userID.(string), addressID)
		if err != nil {
			if err == database.ErrCantFindUser {
				c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
				return
			}
			if err == database.ErrAddressNotFound {
				c.JSON(http.StatusNotFound, gin.H{"error": "address not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "address deleted successfully"})
	}
}
