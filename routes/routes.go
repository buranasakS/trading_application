package routes

import (
	"github.com/buranasakS/trading_application/handlers"
	"github.com/gin-gonic/gin"
	"github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func SetupRoutes(router *gin.Engine) {
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	productRoutes := router.Group("/products")
	{
		productRoutes.POST("/", handlers.CreateProductHandler)
		productRoutes.GET("/list", handlers.ListProductsHandler)
		productRoutes.GET("/:id", handlers.GetProductByIDHandler)
	}

	affiliateRoutes := router.Group("/affiliates")
	{
		affiliateRoutes.POST("/", handlers.CreateAffiliateHandler)
		affiliateRoutes.GET("/list", handlers.ListAffiliatesHandler)
		affiliateRoutes.GET("/:id", handlers.GetAffiliateByIDHandler)
	}

	commissionRoutes := router.Group("/commissions")
	{
		commissionRoutes.GET("/list", handlers.ListCommissionsHandler)
		commissionRoutes.GET("/:id", handlers.GetCommissionByIDHandler)
	}

	userRoutes := router.Group("/users")
	{
		userRoutes.POST("/", handlers.CreateUserHandler)
		userRoutes.GET("/all", handlers.ListUsersHandler)
		userRoutes.GET("/:id", handlers.GetUserDetailByIDHandler)
		userRoutes.PATCH("/deduct/balance/:id", handlers.DeductUserBalanceHandler)
		userRoutes.PATCH("/add/balance/:id", handlers.AddUserBalanceHandler)
		userRoutes.POST("/order", handlers.UserOrderProductHandler)
	}

}
