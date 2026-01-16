package controllers

import (
	"context"
	"fmt"
	"log"
	"net/http"
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
		user.IsAdmin = true // Set as admin
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
		if len(soldProductIDs) > 0 {
			// Create array of sold product IDs
			soldIDsArray := make([]primitive.ObjectID, 0, len(soldProductIDs))
			for productID := range soldProductIDs {
				soldIDsArray = append(soldIDsArray, productID)
			}
			// Exclude sold products
			filter["product_id"] = bson.M{"$nin": soldIDsArray}
		}

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

		// Create Mongodb filter for text search
		filter := bson.M{"product_name": bson.M{"$regex": query, "$options": "i"}}

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

func SearchProductByQuery() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		// Get Search Query from request
		query := c.Query("search")
		if query == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "search query is required"})
			return
		}

		// Search in product name and price fields
		filter := bson.M{
			"$or": []bson.M{
				{
					"product_name": bson.M{
						"$regex":   query,
						"$options": "i",
					},
				},
				{
					"price": bson.M{
						"$gte": query,
						"$lte": query,
					},
				},
			},
		}

		cursor, err := ProductCollection.Find(ctx, filter)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err})
			return
		}
		defer cursor.Close(ctx)

		var products []models.ProductUser
		if err := cursor.All(ctx, &products); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode results"})
			return
		}

		defer cancel()
		c.JSON(http.StatusOK, gin.H{"success": true, "total_count": len(products), "data": products})

	}

}
