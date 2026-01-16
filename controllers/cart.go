package controllers

import (
	"net/http"

	"github/akhil/ecommerce-yt/database"
	"github/akhil/ecommerce-yt/models"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// AddToCart adds a product to the user's cart
func (app *Application) AddToCart() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user_id from context (set by middleware)
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
			return
		}

		// Get product ID from query parameter
		productQueryID := c.Query("id")
		if productQueryID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "product id is required"})
			return
		}

		productID, err := primitive.ObjectIDFromHex(productQueryID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid product id"})
			return
		}

		// Call database function
		err = database.AddProductToCart(app.ProductCollection, app.UserCollection, productID, userID.(string))
		if err != nil {
			handleCartError(c, err)
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "product added to cart successfully"})
	}
}

// RemoveItem removes an item from the user's cart
func (app *Application) RemoveItem() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user_id from context (set by middleware)
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
			return
		}

		// Get product ID from query parameter
		productQueryID := c.Query("id")
		if productQueryID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "product id is required"})
			return
		}

		productID, err := primitive.ObjectIDFromHex(productQueryID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid product id"})
			return
		}

		// Call database function
		err = database.RemoveProductFromCart(app.UserCollection, productID, userID.(string))
		if err != nil {
			handleCartError(c, err)
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "product removed from cart successfully"})
	}
}

// GetItemFromCart retrieves items from the user's cart
func (app *Application) GetItemFromCart() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user_id from context (set by middleware)
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
			return
		}

		// Call database function
		cart, err := database.GetUserCart(app.UserCollection, userID.(string))
		if err != nil {
			handleCartError(c, err)
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"cart":  cart,
			"count": len(cart),
		})
	}
}

// BuyFromCart processes checkout from the cart
func (app *Application) BuyFromCart() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user_id from context (set by middleware)
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
			return
		}

		// Get payment method from request body (optional)
		var paymentMethod *models.Payment
		if c.Request.ContentLength > 0 {
			var payment models.Payment
			if err := c.ShouldBindJSON(&payment); err == nil {
				paymentMethod = &payment
			}
		}

		// Call database function
		orderID, totalPrice, err := database.BuyItemFromCart(app.UserCollection, userID.(string), paymentMethod)
		if err != nil {
			handleCartError(c, err)
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message":     "order placed successfully",
			"order_id":    orderID,
			"total_price": totalPrice,
		})
	}
}

// InstantBuy processes an instant purchase without adding to cart
func (app *Application) InstantBuy() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user_id from context (set by middleware)
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
			return
		}

		// Get product ID from query parameter
		productQueryID := c.Query("id")
		if productQueryID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "product id is required"})
			return
		}

		productID, err := primitive.ObjectIDFromHex(productQueryID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid product id"})
			return
		}

		// Get payment method from request body (optional)
		var paymentMethod *models.Payment
		if c.Request.ContentLength > 0 {
			var payment models.Payment
			if err := c.ShouldBindJSON(&payment); err == nil {
				paymentMethod = &payment
			}
		}

		// Call database function
		orderID, price, err := database.InstantBuy(app.ProductCollection, app.UserCollection, productID, userID.(string), paymentMethod)
		if err != nil {
			handleCartError(c, err)
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message":  "instant buy successful",
			"order_id": orderID,
			"price":    price,
		})
	}
}

// handleCartError is a helper function to handle common cart operation errors
func handleCartError(c *gin.Context, err error) {
	switch err {
	case database.ErrCantFindProduct:
		c.JSON(http.StatusNotFound, gin.H{"error": "product not found"})
	case database.ErrCantDecodeProducts:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error processing product data"})
	case database.ErrUserIdNotMatch:
		c.JSON(http.StatusForbidden, gin.H{"error": "unauthorized access to cart"})
	case database.ErrCantUpdateUser:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update cart"})
	case database.ErrCantRemoveItemCart:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to remove item from cart"})
	case database.ErrCantGetItem:
		c.JSON(http.StatusBadRequest, gin.H{"error": "cart is empty or item not found"})
	case database.ErrCantBuyCartItem:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to process order"})
	case database.ErrProductAlreadyInCart:
		c.JSON(http.StatusBadRequest, gin.H{"error": "product already in cart"})
	case database.ErrProductAlreadySold:
		c.JSON(http.StatusBadRequest, gin.H{"error": "product has already been sold"})
	case database.ErrDuplicateOrder:
		c.JSON(http.StatusConflict, gin.H{"error": "order already processed"})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}
