package database

import (
	"context"
	"errors"
	"time"

	"github/akhil/ecommerce-yt/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var (
	ErrCantFindProduct      = errors.New("can't find product")
	ErrCantDecodeProducts   = errors.New("can't decode products")
	ErrUserIdNotMatch       = errors.New("this user is not the owner of the cart")
	ErrCantUpdateUser       = errors.New("can't update user")
	ErrCantRemoveItemCart   = errors.New("can't remove item from cart")
	ErrCantGetItem          = errors.New("no item in cart")
	ErrCantBuyCartItem      = errors.New("cannot buy cart item")
	ErrProductAlreadyInCart = errors.New("product already in cart")
	ErrProductAlreadySold   = errors.New("product has already been sold")
	ErrDuplicateOrder       = errors.New("order already processed")
)

// AddProductToCart adds a product to the user's cart
func AddProductToCart(productCollection, userCollection *mongo.Collection, productID primitive.ObjectID, userID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Find the product
	var product models.Product
	err := productCollection.FindOne(ctx, bson.M{"product_id": productID}).Decode(&product)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return ErrCantFindProduct
		}
		return ErrCantDecodeProducts
	}

	// Find the user
	var user models.User
	err = userCollection.FindOne(ctx, bson.M{"user_id": userID}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return ErrCantFindProduct
		}
		return ErrCantUpdateUser
	}

	// Check if product already exists in cart
	for _, item := range user.User_Cart {
		if item.Product_ID == productID {
			return ErrProductAlreadyInCart
		}
	}

	// Convert product to ProductUser format
	price := int(*product.Price)
	rating := uint(*product.Rating)
	productUser := models.ProductUser{
		Product_ID:   product.Product_ID,
		Product_Name: product.Product_Name,
		Price:        &price,
		Rating:       &rating,
		Image:        product.Image,
	}

	// Add product to cart
	user.User_Cart = append(user.User_Cart, productUser)

	// Update user in database
	filter := bson.D{primitive.E{Key: "user_id", Value: userID}}
	update := bson.D{{Key: "$set", Value: bson.D{
		{Key: "user_cart", Value: user.User_Cart},
		{Key: "updated_at", Value: time.Now()},
	}}}

	result, err := userCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		return ErrCantUpdateUser
	}
	if result.MatchedCount == 0 {
		return ErrCantFindProduct
	}
	if result.ModifiedCount == 0 {
		// This shouldn't happen when adding to cart, but log it
		return ErrCantUpdateUser
	}

	return nil
}

// RemoveProductFromCart removes a product from the user's cart
func RemoveProductFromCart(userCollection *mongo.Collection, productID primitive.ObjectID, userID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Find the user
	var user models.User
	err := userCollection.FindOne(ctx, bson.M{"user_id": userID}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return ErrCantFindProduct
		}
		return ErrCantUpdateUser
	}

	// Remove product from cart
	var updatedCart []models.ProductUser
	found := false
	for _, item := range user.User_Cart {
		if item.Product_ID != productID {
			updatedCart = append(updatedCart, item)
		} else {
			found = true
		}
	}

	if !found {
		return ErrCantGetItem
	}

	// Update user in database
	filter := bson.D{primitive.E{Key: "user_id", Value: userID}}
	update := bson.D{{Key: "$set", Value: bson.D{
		{Key: "user_cart", Value: updatedCart},
		{Key: "updated_at", Value: time.Now()},
	}}}

	_, err = userCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		return ErrCantRemoveItemCart
	}

	return nil
}

// GetUserCart retrieves all items from the user's cart
func GetUserCart(userCollection *mongo.Collection, userID string) ([]models.ProductUser, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Find the user
	var user models.User
	err := userCollection.FindOne(ctx, bson.M{"user_id": userID}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, ErrCantFindProduct
		}
		return nil, ErrCantUpdateUser
	}

	return user.User_Cart, nil
}

func BuyItemFromCart(userCollection *mongo.Collection, userID string, paymentMethod *models.Payment) (string, int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Find the user
	var user models.User
	err := userCollection.FindOne(ctx, bson.M{"user_id": userID}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return "", 0, ErrCantFindProduct
		}
		return "", 0, ErrCantUpdateUser
	}

	// Check if cart is empty - if so, check for recent duplicate order (idempotency)
	if len(user.User_Cart) == 0 {
		// Check if there's a recent order (within last 10 seconds) - idempotency check
		if len(user.Order_Status) > 0 {
			lastOrder := user.Order_Status[len(user.Order_Status)-1]
			// If order was created within last 10 seconds, return it (idempotent)
			if time.Since(lastOrder.Ordered_At) < 10*time.Second {
				return lastOrder.Order_ID.Hex(), lastOrder.Price, nil
			}
		}
		return "", 0, ErrCantGetItem
	}

	// Check if any product in cart has already been sold
	soldProductIDs, err := GetSoldProductIDs(userCollection)
	if err != nil {
		return "", 0, ErrCantDecodeProducts
	}
	for _, item := range user.User_Cart {
		if soldProductIDs[item.Product_ID] {
			return "", 0, ErrProductAlreadySold
		}
	}

	// Calculate total price
	var totalPrice int
	for _, item := range user.User_Cart {
		if item.Price != nil {
			totalPrice += *item.Price
		}
	}

	// Set default payment method if not provided (defaults to COD)
	if paymentMethod == nil {
		paymentMethod = &models.Payment{
			Digital: false,
			COD:     true,
		}
	}

	// Create order
	order := models.Order{
		Order_ID:       primitive.NewObjectID(),
		Order_Cart:     user.User_Cart,
		Ordered_At:     time.Now(),
		Price:          totalPrice,
		Discount:       0,
		Payment_Method: paymentMethod,
	}

	// Add order to user's order status
	user.Order_Status = append(user.Order_Status, order)

	// Clear cart and update user
	filter := bson.D{primitive.E{Key: "user_id", Value: userID}}
	update := bson.D{{Key: "$set", Value: bson.D{
		{Key: "user_cart", Value: []models.ProductUser{}},
		{Key: "order_status", Value: user.Order_Status},
		{Key: "updated_at", Value: time.Now()},
	}}}

	_, err = userCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		return "", 0, ErrCantBuyCartItem
	}

	return order.Order_ID.Hex(), totalPrice, nil
}

// InstantBuy processes an instant purchase without adding to cart and returns order ID and price
// paymentMethod can be nil, in which case it defaults to COD
func InstantBuy(productCollection, userCollection *mongo.Collection, productID primitive.ObjectID, userID string, paymentMethod *models.Payment) (string, int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Check if product has already been sold
	soldProductIDs, err := GetSoldProductIDs(userCollection)
	if err != nil {
		return "", 0, ErrCantDecodeProducts
	}
	if soldProductIDs[productID] {
		return "", 0, ErrProductAlreadySold
	}

	// Find the product
	var product models.Product
	err = productCollection.FindOne(ctx, bson.M{"product_id": productID}).Decode(&product)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return "", 0, ErrCantFindProduct
		}
		return "", 0, ErrCantDecodeProducts
	}

	// Find the user
	var user models.User
	err = userCollection.FindOne(ctx, bson.M{"user_id": userID}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return "", 0, ErrCantFindProduct
		}
		return "", 0, ErrCantUpdateUser
	}

	// Convert product to ProductUser format
	price := int(*product.Price)
	rating := uint(*product.Rating)
	productUser := models.ProductUser{
		Product_ID:   product.Product_ID,
		Product_Name: product.Product_Name,
		Price:        &price,
		Rating:       &rating,
		Image:        product.Image,
	}

	// Set default payment method if not provided (defaults to COD)
	if paymentMethod == nil {
		paymentMethod = &models.Payment{
			Digital: false,
			COD:     true,
		}
	}

	// Create order with single product
	order := models.Order{
		Order_ID:       primitive.NewObjectID(),
		Order_Cart:     []models.ProductUser{productUser},
		Ordered_At:     time.Now(),
		Price:          int(*product.Price),
		Discount:       0,
		Payment_Method: paymentMethod,
	}

	// Add order to user's order status
	user.Order_Status = append(user.Order_Status, order)

	// Update user
	filter := bson.D{primitive.E{Key: "user_id", Value: userID}}
	update := bson.D{{Key: "$set", Value: bson.D{
		{Key: "order_status", Value: user.Order_Status},
		{Key: "updated_at", Value: time.Now()},
	}}}

	_, err = userCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		return "", 0, ErrCantBuyCartItem
	}

	return order.Order_ID.Hex(), int64(*product.Price), nil
}

// GetSoldProductIDs retrieves all product IDs that have been sold (appear in any order)
func GetSoldProductIDs(userCollection *mongo.Collection) (map[primitive.ObjectID]bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Get all users with their orders
	cursor, err := userCollection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	soldProductIDs := make(map[primitive.ObjectID]bool)

	// Iterate through all users
	for cursor.Next(ctx) {
		var user models.User
		if err := cursor.Decode(&user); err != nil {
			continue
		}

		// Extract product IDs from all orders
		for _, order := range user.Order_Status {
			for _, product := range order.Order_Cart {
				soldProductIDs[product.Product_ID] = true
			}
		}
	}

	return soldProductIDs, nil
}
