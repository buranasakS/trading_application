package routes

import (
	"github.com/buranasakS/trading_application/handlers"
	"github.com/buranasakS/trading_application/middleware"
	"github.com/gin-gonic/gin"
	"github.com/gin-contrib/cors"
	"github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func SetupRoutes(router *gin.Engine, h *handlers.Handler) {
	router.Use(cors.Default())

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// protected := router.Group("/", middleware.JwtMiddleware())
	router.POST("/login", h.LoginUserHandler)
	router.POST("/register", h.RegisterUserHandler)

	productRoutes := router.Group("/products")
	productRoutes.Use(middleware.JwtMiddleware())
    {
        productRoutes.POST("", h.CreateProductHandler)
        productRoutes.GET("/list", h.ListProductsHandler)
        productRoutes.GET("/:id", h.GetProductDetailHandler)
    }

	affiliateRoutes := router.Group("/affiliates") 
	affiliateRoutes.Use(middleware.JwtMiddleware())
	{
		affiliateRoutes.POST("", h.CreateAffiliateHandler)
		affiliateRoutes.GET("/list", h.ListAffiliatesHandler)
		affiliateRoutes.GET("/:id", h.GetAffiliateDetailHandler)
	}

	commissionRoutes := router.Group("/commissions")
	commissionRoutes.Use(middleware.JwtMiddleware())
	{
		commissionRoutes.GET("/list", h.ListCommissionsHandler)
		commissionRoutes.GET("/:id", h.GetCommissionDetailHandler)
	}

	userRoutes := router.Group("/users")
	// userRoutes.Use(middleware.JwtMiddleware())
	{
		userRoutes.GET("/all", h.ListUsersHandler)
		userRoutes.GET("/:id", h.GetUserDetailHandler)
		userRoutes.PATCH("/deduct/balance/:id", h.DeductUserBalanceHandler)
		userRoutes.PATCH("/add/balance/:id", h.AddUserBalanceHandler)
		userRoutes.POST("/order", h.UserOrderProductHandler)
	}

}
