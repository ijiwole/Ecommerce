package routes

import (
	"github/akhil/ecommerce-yt/controllers"

	"github.com/gin-gonic/gin"
)

// UserRoutes sets up user-related routes (public routes)
func UserRoutes(incomingRoutes *gin.Engine) {
	incomingRoutes.POST("api/v1/users/signup", controllers.SignUp())
	incomingRoutes.POST("api/v1/users/login", controllers.Login())
	incomingRoutes.POST("api/v1/admin/signup", controllers.AdminSignUp())
	incomingRoutes.POST("api/v1/admin/login", controllers.AdminLogin())
	incomingRoutes.GET("api/v1/users/productview", controllers.SearchProduct())
	incomingRoutes.GET("api/v1/users/search", controllers.SearchProductByQuery())
}

// AdminRoutes sets up admin-related routes (requires authentication and admin privileges)
func AdminRoutes(incomingRoutes *gin.RouterGroup) {
	incomingRoutes.POST("api/v1/admin/addproduct", controllers.ProductViewerAdmin())
}

// ProductRoutes sets up product-related routes (requires authentication)
func ProductRoutes(incomingRoutes *gin.RouterGroup) {
	incomingRoutes.GET("api/v1/products", controllers.GetAllProducts())
}

// CartRoutes sets up cart-related routes (requires authentication)
func CartRoutes(incomingRoutes *gin.RouterGroup, app *controllers.Application) {
	incomingRoutes.POST("api/v1/cart/add", app.AddToCart())
	incomingRoutes.DELETE("api/v1/cart/remove", app.RemoveItem())
	incomingRoutes.GET("api/v1/cart", app.GetItemFromCart())
	incomingRoutes.POST("api/v1/cart/checkout", app.BuyFromCart())
	incomingRoutes.POST("api/v1/cart/instantbuy", app.InstantBuy())
}

// AddressRoutes sets up address-related routes (requires authentication)
func AddressRoutes(incomingRoutes *gin.RouterGroup, app *controllers.Application) {
	incomingRoutes.GET("api/v1/address", app.GetAddresses())
	incomingRoutes.POST("api/v1/address", app.AddAddress())
	incomingRoutes.PUT("api/v1/address/home", app.EditHomeAddress())
	incomingRoutes.PUT("api/v1/address/work", app.EditWorkAddress())
	incomingRoutes.DELETE("api/v1/address", app.DeleteAddress())
}
