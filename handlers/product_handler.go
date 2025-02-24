package handlers

import (
	"context"
	"net/http"

	"github.com/buranasakS/trading_application/config"
	db "github.com/buranasakS/trading_application/db/sqlc"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"
)

// CreateProductHandler godoc
// @Summary      Create a new product
// @Description  Create a new product details
// @Tags         Products
// @Accept       json
// @Produce      json
// @Param        request body      db.CreateProductParams true "Product details"
// @Success      201  {object}  db.Product "Product created successfully"
// @Router       /products [post]
func CreateProductHandler(c *gin.Context) {
	var req db.CreateProductParams
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	if req.Quantity <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Quantity must be more than 0"})
		return
	}

	if req.Price <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Price must be more than 0"})
		return
	}

	queries := db.New(config.ConnectDatabase().DB)
	product, err := queries.CreateProduct(context.Background(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create product"})
		return
	}

	c.JSON(http.StatusCreated, product)
}

// ListProductsHandler godoc
// @Summary      List all products
// @Description  List all products
// @Tags         Products
// @Accept       json
// @Produce      json
// @Success      200  {object}  []db.Product "List of products"
// @Router       /products/list [get]
func ListProductsHandler(c *gin.Context) {
	queries := db.New(config.ConnectDatabase().DB)
	products, err := queries.ListProducts(context.Background())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list products"})
		return
	}

	c.JSON(http.StatusOK, products)
}

// GetProductByIDHandler godoc
// @Summary      Get product details by ID
// @Description  Retrieve product details by their unique ID
// @Tags         Products
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "Product ID (UUID)"
// @Success      200  {object}  db.Product
// @Router       /products/{id} [get]
func GetProductByIDHandler(c *gin.Context) {
	var productId pgtype.UUID
	if err := productId.Scan(c.Param("id")); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID"})
		return
	}

	queries := db.New(config.ConnectDatabase().DB)
	product, err := queries.GetProductByID(context.Background(), productId)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
		return
	}

	c.JSON(http.StatusOK, product)
}
