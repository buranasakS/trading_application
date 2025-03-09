package handlers

import (
	"context"
	"net/http"

	db "github.com/buranasakS/trading_application/db/sqlc"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"
)

type RequestProduct struct{
	Name     string  `json:"name" binding:"required"`
	Quantity int32   `json:"quantity" binding:"required"`
	Price    float64 `json:"price" binding:"required"`
}

// CreateProductHandler godoc
// @Summary      Create a new product
// @Description  Create a new product details
// @Tags         Products
// @Security BearerAuth
// @Accept       json
// @Produce      json
// @Param        request body      db.CreateProductParams true "Product details"
// @Success      201  {object}  db.Product "Product created successfully"
// @Failure 401 {object} handlers.ErrorResponse "Unauthorized"
// @Router       /products [post]
func (h *Handler) CreateProductHandler(c *gin.Context) {
	var req RequestProduct
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

	product, err := h.db.CreateProduct(context.Background(), db.CreateProductParams{
		Name: req.Name,
		Quantity: req.Quantity,
		Price: req.Price,
	})
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
// @Security BearerAuth
// @Accept       json
// @Produce      json
// @Success      200  {object}  []db.Product "List of products"
// @Failure 401 {object} handlers.ErrorResponse "Unauthorized"
// @Router       /products/list [get]
func (h *Handler) ListProductsHandler(c *gin.Context) {

	products, err := h.db.ListProducts(context.Background())
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
// @Security BearerAuth
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "Product ID (UUID)"
// @Success      200  {object}  db.Product
// @Failure 401 {object} handlers.ErrorResponse "Unauthorized"
// @Router       /products/{id} [get]
func (h *Handler) GetProductDetailHandler(c *gin.Context) {
	var productId pgtype.UUID
	if err := productId.Scan(c.Param("id")); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID"})
		return
	}
	product, err := h.db.GetProductByID(context.Background(), productId)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
		return
	}

	c.JSON(http.StatusOK, product)
}
