package main

import (
	"github/akhil/ecommerce-yt/controllers"
	"github/akhil/ecommerce-yt/database"
	"github/akhil/ecommerce-yt/middleware"
	"github/akhil/ecommerce-yt/routes"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env file
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, using system environment variables")
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}

	app := controllers.NewApplication(database.ProductData(database.Client, "Products"), database.UserData(database.Client, "Users"))

	router := gin.New()
	router.Use(gin.Logger())

	// Set trusted proxies for security (only trust localhost and private networks in development)
	// In production, set this to your actual proxy/load balancer IPs
	router.SetTrustedProxies([]string{"127.0.0.1", "::1"})

	// Public routes (no authentication required)
	routes.UserRoutes(router)

	// Protected routes (authentication required)
	protected := router.Group("/")
	protected.Use(middleware.Authentication())
	{
		// Product routes (accessible to all authenticated users)
		routes.ProductRoutes(protected)

		// Cart routes
		routes.CartRoutes(protected, app)

		// Address routes
		routes.AddressRoutes(protected, app)
	}

	// Admin routes (authentication + admin privileges required)
	admin := router.Group("/")
	admin.Use(middleware.Authentication())
	admin.Use(middleware.AdminAuth())
	{
		routes.AdminRoutes(admin)
	}

	log.Fatal(router.Run(":" + port))
}
