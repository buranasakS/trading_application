package handlers

import (
	"context"
	"net/http"

	"github.com/buranasakS/trading_application/config"
	db "github.com/buranasakS/trading_application/db/sqlc"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type OrderRequest struct {
	UserID    pgtype.UUID `json:"user_id"`
	ProductID pgtype.UUID `json:"product_id"`
	Quantity  int         `json:"quantity"`
}

type OrderResponse struct {
	Status    string  `json:"status"`
	Message   string  `json:"message"`
	OrderID   string  `json:"order_id"`
	TotalCost float64 `json:"total_cost"`
}

// UserOrderProductHandler godoc
// @Summary      ordering a product and calculate commission
// @Description  ordering a product and calculate commission
// @Tags         User ordering a product
// @Security BearerAuth
// @Accept       json
// @Produce      json
// @Param        request body    OrderRequest true "Order product detail"
// @Success      201  {object}   OrderResponse  "Order completed"
// @Failure 401 {object} handlers.ErrorResponse "Unauthorized"
// @Router       /users/order [post]
func (h *Handler) UserOrderProductHandler(c *gin.Context) {
	var req OrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}

	if req.Quantity < 1 {
		c.JSON(http.StatusBadRequest, gin.H{"Error": "Quantity must be more than 0"})
		return
	}

	tx, err := config.ConnectDatabase().DB.BeginTx(context.Background(), pgx.TxOptions{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"Error": "Failed to start transaction"})
		return
	}
	defer tx.Rollback(context.Background())

	qtx := db.New(tx)

	user, err := qtx.GetUserDetailByID(context.Background(), req.UserID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"Error": "User not found"})
		return
	}

	product, err := qtx.GetProductByID(context.Background(), req.ProductID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"Error": "Product not found"})
		return
	}

	totalPrice := product.Price * float64(req.Quantity)
	if user.Balance < totalPrice {
		c.JSON(http.StatusBadRequest, gin.H{"Error": "Not enough balance"})
		return
	}

	if product.Quantity < int32(req.Quantity) {
		c.JSON(http.StatusBadRequest, gin.H{"Error": "Not enough product in stock"})
		return
	}

	_, err = qtx.DeductUserBalance(context.Background(), db.DeductUserBalanceParams{
		Balance: totalPrice,
		ID:      req.UserID,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"Error": "Failed to deduct balance"})
		return
	}

	_, err = qtx.DeductProductQuantity(context.Background(), db.DeductProductQuantityParams{
		Quantity: int32(req.Quantity),
		ID:       req.ProductID,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"Error": "Failed to deduct product quantity"})
		return
	}

	orderID := uuid.New()
	if user.AffiliateID.Valid {
		affiliateLevel := 0
		currentAffiliateID := user.AffiliateID
		var previousCommissionRate float64 = 0 

		for currentAffiliateID.Valid {
			affiliate, err := qtx.GetAffiliateByID(context.Background(), currentAffiliateID)
			if err != nil || !affiliate.ID.Valid {
				break
			}

			commissionRate := 0.0
			if affiliateLevel == 0 {
				commissionRate = 0.20
			} else if affiliateLevel == 1 {
				commissionRate = 0.15
			} else if affiliateLevel == 2 {
				commissionRate = 0.10
			} else if affiliateLevel == 3 {
				commissionRate = 0.05
			} else if affiliateLevel > 3 {
				commissionRate = previousCommissionRate - 0.01
				if commissionRate < 0 {
					commissionRate = 0
				}
			}

			if commissionRate > 0 {
				var commissionAmount float64
				if affiliateLevel <= 3 {
					commissionAmount = commissionRate * totalPrice
				} else {
					commissionAmount = (previousCommissionRate - commissionRate) * totalPrice
				}
				_, err = qtx.CreateCommission(context.Background(), db.CreateCommissionParams{
					OrderID:     pgtype.UUID{Bytes: orderID, Valid: true},
					AffiliateID: affiliate.ID,
					Amount:      commissionAmount,
				})
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create commission"})
					return
				}

				err = qtx.AddAffiliateBalance(context.Background(), db.AddAffiliateBalanceParams{
					ID:      affiliate.ID,
					Balance: commissionAmount,
				})
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add affiliate balance"})
					return
				}
			}

			previousCommissionRate = commissionRate
			affiliateLevel++
			currentAffiliateID = affiliate.MasterAffiliate
		}
	}

	err = tx.Commit(context.Background())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"Error": "Failed to commit transaction"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"status":     "success",
		"message":    "Purchase completed",
		"order_id":   orderID.String(),
		"total_cost": totalPrice,
	})
}
