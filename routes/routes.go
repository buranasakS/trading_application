package routes

import (
	// db "github.com/buranasakS/trading_application/db/sqlc"
	"github.com/buranasakS/trading_application/handlers"
	"github.com/buranasakS/trading_application/middleware"
	"github.com/gin-gonic/gin"
	"github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func SetupRoutes(router *gin.Engine) {
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// protected := router.Group("/", middleware.JwtMiddleware())

	router.POST("/login", handlers.LoginUser)
	router.POST("/register", handlers.RegisterUser)

	productRoutes := router.Group("/products")
	productRoutes.Use(middleware.JwtMiddleware())
	{
		productRoutes.POST("/", handlers.CreateProductHandler)
		productRoutes.GET("/list", handlers.ListProductsHandler)
		productRoutes.GET("/:id", handlers.GetProductByIDHandler)
	}

	affiliateRoutes := router.Group("/affiliates")
	affiliateRoutes.Use(middleware.JwtMiddleware())
	{
		affiliateRoutes.POST("/", handlers.CreateAffiliateHandler)
		affiliateRoutes.GET("/list", handlers.ListAffiliatesHandler)
		affiliateRoutes.GET("/:id", handlers.GetAffiliateByIDHandler)
	}

	commissionRoutes := router.Group("/commissions")
	commissionRoutes.Use(middleware.JwtMiddleware())
	{
		commissionRoutes.GET("/list", handlers.ListCommissionsHandler)
		commissionRoutes.GET("/:id", handlers.GetCommissionByIDHandler)
	}

	userRoutes := router.Group("/users")
	userRoutes.Use(middleware.JwtMiddleware())
	{
		// userRoutes.POST("/login", handlers.LoginUser)
		// userRoutes.POST("/register", handlers.RegisterUser)
		userRoutes.GET("/all", handlers.ListUsersHandler)
		userRoutes.GET("/:id", handlers.GetUserDetailByIDHandler)
		userRoutes.PATCH("/deduct/balance/:id", handlers.DeductUserBalanceHandler)
		userRoutes.PATCH("/add/balance/:id", handlers.AddUserBalanceHandler)
		userRoutes.POST("/order", handlers.UserOrderProductHandler)
	}

}
