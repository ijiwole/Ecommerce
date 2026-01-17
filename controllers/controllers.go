package controllers

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github/akhil/ecommerce-yt/database"
	"github/akhil/ecommerce-yt/helpers"
	"github/akhil/ecommerce-yt/models"
	generate "github/akhil/ecommerce-yt/tokens"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"
)

var UserCollection *mongo.Collection = database.UserData(database.Client, "Users")
var ProductCollection *mongo.Collection = database.ProductData(database.Client, "Products")
var validate = validator.New()

type Application struct {
	ProductCollection *mongo.Collection
	UserCollection    *mongo.Collection
}

func NewApplication(productCollection, userCollection *mongo.Collection) *Application {
	return &Application{
		ProductCollection: productCollection,
		UserCollection:    userCollection,
	}
}

func HashPassword(password string) string {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil {
		log.Panic(err)
	}
	return string(hashedPassword)
}

func VerifyPassword(userPassword string, givenPassword string) (bool, string) {

	hashedPassword := []byte(givenPassword)
	err := bcrypt.CompareHashAndPassword(hashedPassword, []byte(userPassword))
	if err != nil {
		return false, "password is incorrect"
	}
	return true, "password is correct"
}

// addSoldProductExclusion adds a filter condition to exclude sold products from the given filter.
// It handles both simple filters and filters with $and conditions.
func addSoldProductExclusion(filter bson.M, soldProductIDs map[primitive.ObjectID]bool) {
	if len(soldProductIDs) == 0 {
		return
	}

	// Convert sold product IDs map to array
	soldIDsArray := make([]primitive.ObjectID, 0, len(soldProductIDs))
	for productID := range soldProductIDs {
		soldIDsArray = append(soldIDsArray, productID)
	}

	exclusion := bson.M{"product_id": bson.M{"$nin": soldIDsArray}}

	// If filter already has $and, append to it
	if andConditions, ok := filter["$and"].([]bson.M); ok {
		filter["$and"] = append(andConditions, exclusion)
	} else {
		// Otherwise, add directly to filter
		filter["product_id"] = exclusion["product_id"]
	}
}

func SignUp() gin.HandlerFunc {

	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		var user models.User
		if err := c.BindJSON(&user); err != nil {
			helpers.BadRequest(c, err.Error())
			return
		}

		validationErr := validate.Struct(user)
		if validationErr != nil {
			helpers.BadRequest(c, validationErr.Error())
			return
		}

		count, err := UserCollection.CountDocuments(ctx, bson.M{"email": user.Email})
		if err != nil {
			log.Panic(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err})
			return
		}

		if count > 0 {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "user already exists"})
			return
		}

		count, err = UserCollection.CountDocuments(ctx, bson.M{"phone": user.Phone})
		if err != nil {
			log.Panic(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err})
			return
		}

		if count > 0 {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "this phone number is already in use"})
			return
		}
		password := HashPassword(*user.Password)
		user.Password = &password

		user.Created_At, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		user.Updated_At, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		user.ID = primitive.NewObjectID()
		user.User_ID = user.ID.Hex()
		token, refreshtoken, _ := generate.TokenGenerator(*user.Email, *user.First_Name, *user.Last_Name, user.User_ID)
		user.Token = &token
		user.Refresh_Token = &refreshtoken
		user.User_Cart = make([]models.ProductUser, 0)
		user.Address_Details = make([]models.Address, 0)
		user.Order_Status = make([]models.Order, 0)
		user.IsAdmin = false // Regular user signup
		_, insertedErr := UserCollection.InsertOne(ctx, user)
		if insertedErr != nil {
			helpers.InternalServerError(c, insertedErr.Error())
			return
		}
		defer cancel()
		helpers.Success(c, "Signed Up Successfully", nil)
	}
}

// AdminSignUp handles admin registration
func AdminSignUp() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		var user models.User
		if err := c.BindJSON(&user); err != nil {
			helpers.BadRequest(c, err.Error())
			return
		}

		validationErr := validate.Struct(user)
		if validationErr != nil {
			helpers.BadRequest(c, validationErr.Error())
			return
		}

		count, err := UserCollection.CountDocuments(ctx, bson.M{"email": user.Email})
		if err != nil {
			log.Panic(err)
			helpers.InternalServerError(c, err.Error())
			return
		}

		if count > 0 {
			helpers.InternalServerError(c, "admin with this email already exists")
			return
		}

		count, err = UserCollection.CountDocuments(ctx, bson.M{"phone": user.Phone})
		if err != nil {
			log.Panic(err)
			helpers.InternalServerError(c, err.Error())
			return
		}

		if count > 0 {
			helpers.InternalServerError(c, "this phone number is already in use")
			return
		}

		password := HashPassword(*user.Password)
		user.Password = &password

		user.Created_At, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		user.Updated_At, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		user.ID = primitive.NewObjectID()
		user.User_ID = user.ID.Hex()
		user.IsAdmin = true
		token, refreshtoken, _ := generate.TokenGenerator(*user.Email, *user.First_Name, *user.Last_Name, user.User_ID)
		user.Token = &token
		user.Refresh_Token = &refreshtoken
		user.User_Cart = make([]models.ProductUser, 0)
		user.Address_Details = make([]models.Address, 0)
		user.Order_Status = make([]models.Order, 0)

		_, insertedErr := UserCollection.InsertOne(ctx, user)
		if insertedErr != nil {
			helpers.InternalServerError(c, insertedErr.Error())
			return
		}

		helpers.Success(c, "Admin Signed Up Successfully", nil)
	}
}

func Login() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		var user models.User
		if err := c.BindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err})
			return
		}

		var foundUser models.User
		err := UserCollection.FindOne(ctx, bson.M{"email": user.Email}).Decode(&foundUser)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "email or password is incorrect"})
			return
		}

		PasssordIsValid, msg := VerifyPassword(*user.Password, *foundUser.Password)

		if !PasssordIsValid {
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			fmt.Println(msg)
			return
		}
		token, refreshtoken, _ := generate.TokenGenerator(*foundUser.Email, *foundUser.First_Name, *foundUser.Last_Name, foundUser.User_ID)

		generate.UpdateAllTokens(token, refreshtoken, foundUser.User_ID, UserCollection, ctx)

		helpers.LoginSuccess(c, foundUser.User_ID, token, refreshtoken)
	}
}

// AdminLogin handles admin login
func AdminLogin() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		var user models.User
		if err := c.BindJSON(&user); err != nil {
			helpers.BadRequest(c, err.Error())
			return
		}

		var foundUser models.User
		err := UserCollection.FindOne(ctx, bson.M{"email": user.Email}).Decode(&foundUser)
		if err != nil {
			helpers.Unauthorized(c, "email or password is incorrect")
			return
		}

		// Check if user is an admin
		if !foundUser.IsAdmin {
			helpers.Unauthorized(c, "access denied: admin privileges required")
			return
		}

		PasswordIsValid, msg := VerifyPassword(*user.Password, *foundUser.Password)
		defer cancel()

		if !PasswordIsValid {
			helpers.Unauthorized(c, msg)
			fmt.Println(msg)
			return
		}

		token, refreshtoken, _ := generate.TokenGenerator(*foundUser.Email, *foundUser.First_Name, *foundUser.Last_Name, foundUser.User_ID)

		generate.UpdateAllTokens(token, refreshtoken, foundUser.User_ID, UserCollection, ctx)

		helpers.LoginSuccess(c, foundUser.User_ID, token, refreshtoken)
	}
}

func ProductViewerAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		var products models.Product
		if err := c.BindJSON(&products); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err})
			return
		}

		validationErr := validate.Struct(products)
		if validationErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": validationErr.Error()})
			return
		}

		products.Product_ID = primitive.NewObjectID()
		_, insertedErr := ProductCollection.InsertOne(ctx, products)
		if insertedErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": insertedErr})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success":    true,
			"message":    "Product added successfully",
			"product_id": products.Product_ID.Hex(),
		})
	}
}

// GetAllProducts returns all products with pagination (accessible to all authenticated users)
// Excludes products that have been sold
func GetAllProducts() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		// Get pagination parameters
		pagination := helpers.GetPaginationParams(c)

		// Get all sold product IDs
		soldProductIDs, err := database.GetSoldProductIDs(UserCollection)
		if err != nil {
			helpers.InternalServerError(c, "error fetching sold products")
			return
		}

		// Build filter to exclude sold products
		filter := bson.M{}
		addSoldProductExclusion(filter, soldProductIDs)

		// Get total count of available products (not sold)
		total, err := ProductCollection.CountDocuments(ctx, filter)
		if err != nil {
			helpers.InternalServerError(c, "error counting products")
			return
		}

		// Query products with pagination
		opts := options.Find()
		opts.SetSkip(pagination.Skip)
		opts.SetLimit(pagination.PageSize)

		cursor, err := ProductCollection.Find(ctx, filter, opts)
		if err != nil {
			helpers.InternalServerError(c, "error fetching products")
			return
		}
		defer cursor.Close(ctx)

		// Decode results
		var products []models.Product
		if err := cursor.All(ctx, &products); err != nil {
			helpers.InternalServerError(c, "failed to decode results")
			return
		}

		// Return paginated response
		helpers.PaginatedSuccess(c, products, total, pagination)
	}
}

// SearchProduct searches products by name (excludes sold products)
func SearchProduct() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		// Get Search Query from request
		query := c.Query("search")
		if query == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "search query is required"})
			return
		}

		// Get all sold product IDs
		soldProductIDs, err := database.GetSoldProductIDs(UserCollection)
		if err != nil {
			helpers.InternalServerError(c, "error fetching sold products")
			return
		}

		// Create Mongodb filter for text search
		filter := bson.M{"product_name": bson.M{"$regex": query, "$options": "i"}}

		// Exclude sold products from search results
		addSoldProductExclusion(filter, soldProductIDs)

		// Query Mongodb
		cursor, err := ProductCollection.Find(ctx, filter)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err})
			return
		}
		defer cursor.Close(ctx)

		// Decode results
		var products []models.Product
		if err := cursor.All(ctx, &products); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode results"})
			return
		}

		// return empty array if no products found
		if len(products) == 0 {
			c.JSON(http.StatusOK, gin.H{"success": true, "data": []models.Product{}})
			return
		}

		defer cancel()
		c.JSON(http.StatusOK, gin.H{"success": true, "data": products})

	}

}

// SearchProductByQuery searches products by name and/or price range (excludes sold products)
// Query parameters:
//   - search: optional product name search (case-insensitive regex)
//   - min_price: optional minimum price (numeric)
//   - max_price: optional maximum price (numeric)
//
// At least one parameter must be provided
func SearchProductByQuery() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		// Get query parameters
		query := c.Query("search")
		minPriceStr := c.Query("min_price")
		maxPriceStr := c.Query("max_price")

		// Validate that at least one search parameter is provided
		if query == "" && minPriceStr == "" && maxPriceStr == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "at least one search parameter is required (search, min_price, or max_price)"})
			return
		}

		// Get all sold product IDs
		soldProductIDs, err := database.GetSoldProductIDs(UserCollection)
		if err != nil {
			helpers.InternalServerError(c, "error fetching sold products")
			return
		}

		// Build filter conditions
		andConditions := []bson.M{}

		// Add product name search if provided
		if query != "" {
			andConditions = append(andConditions, bson.M{
				"product_name": bson.M{
					"$regex":   query,
					"$options": "i",
				},
			})
		}

		// Add price range filter if min_price or max_price is provided
		priceFilter := bson.M{}
		if minPriceStr != "" {
			minPrice, err := strconv.ParseUint(minPriceStr, 10, 64)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "min_price must be a valid number"})
				return
			}
			priceFilter["$gte"] = minPrice
		}
		if maxPriceStr != "" {
			maxPrice, err := strconv.ParseUint(maxPriceStr, 10, 64)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "max_price must be a valid number"})
				return
			}
			priceFilter["$lte"] = maxPrice
		}

		// Validate price range (min_price should be <= max_price if both are provided)
		if minPriceStr != "" && maxPriceStr != "" {
			minPrice, _ := strconv.ParseUint(minPriceStr, 10, 64)
			maxPrice, _ := strconv.ParseUint(maxPriceStr, 10, 64)
			if minPrice > maxPrice {
				c.JSON(http.StatusBadRequest, gin.H{"error": "min_price must be less than or equal to max_price"})
				return
			}
		}

		// Add price filter if it has any conditions
		if len(priceFilter) > 0 {
			andConditions = append(andConditions, bson.M{"price": priceFilter})
		}

		// Build final filter
		filter := bson.M{}
		if len(andConditions) > 0 {
			filter["$and"] = andConditions
		}

		// Exclude sold products from search results
		addSoldProductExclusion(filter, soldProductIDs)

		// Get pagination parameters
		pagination := helpers.GetPaginationParams(c)

		// Get total count of matching products (not sold)
		total, err := ProductCollection.CountDocuments(ctx, filter)
		if err != nil {
			helpers.InternalServerError(c, "error counting products")
			return
		}

		// Query products with pagination
		opts := options.Find()
		opts.SetSkip(pagination.Skip)
		opts.SetLimit(pagination.PageSize)

		cursor, err := ProductCollection.Find(ctx, filter, opts)
		if err != nil {
			helpers.InternalServerError(c, "error fetching products")
			return
		}
		defer cursor.Close(ctx)

		// Decode results
		var products []models.Product
		if err := cursor.All(ctx, &products); err != nil {
			helpers.InternalServerError(c, "failed to decode results")
			return
		}

		// Return paginated response
		helpers.PaginatedSuccess(c, products, total, pagination)
	}

}
