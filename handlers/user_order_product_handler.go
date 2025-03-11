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
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Quantity <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Quantity must be more than 0"})
		return
	}

	tx, err := config.ConnectDatabase().DB.BeginTx(context.Background(), pgx.TxOptions{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start transaction"})
		return
	}
	defer tx.Rollback(context.Background())

	var qtx db.Querier = h.db
	if queriesDB, ok := h.db.(*db.Queries); ok {
		qtx = queriesDB.WithTx(tx)
	}

	user, err := h.db.GetUserDetailByID(context.Background(), req.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User not found"})
		return
	}

	product, err := h.db.GetProductByID(context.Background(), req.ProductID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Product not found"})
		return
	}

	if product.Quantity < int32(req.Quantity) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Not enough product in stock"})
		return
	}

	totalPrice := product.Price * float64(req.Quantity)
	if user.Balance < totalPrice {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Not enough balance"})
		return
	}

	_, err = qtx.DeductUserBalance(context.Background(), db.DeductUserBalanceParams{
		Balance: totalPrice,
		ID:      req.UserID,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to deduct balance"})
		return
	}

	_, err = qtx.DeductProductQuantity(context.Background(), db.DeductProductQuantityParams{
		Quantity: int32(req.Quantity),
		ID:       req.ProductID,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to deduct product quantity"})
		return
	}

	orderID := uuid.New()

	if user.AffiliateID.Valid {
		affililates := []db.Affiliate{}
		currentAffiliateID := user.AffiliateID

		for currentAffiliateID.Valid {
			affiliate, err := h.db.GetAffiliateByID(context.Background(), currentAffiliateID)
			if err != nil || !affiliate.ID.Valid {
				break
			}

			affililates = append(affililates, affiliate)
			currentAffiliateID = affiliate.MasterAffiliate
		}

		if len(affililates) == 0 {
			return
		}

		commissionRates := []float64{0.05, 0.10, 0.15, 0.20}
		commissionLevel := 0
		if len(affililates) <= len(commissionRates) {
			commissionLevel = len(commissionRates) - len(affililates)
		}

		for i := 0; i < len(affililates); i++ {
			var commissionAmount float64
			level := commissionLevel + i

			if level >= len(commissionRates) {
				level = len(commissionRates) - 1
				commissionAmount = (commissionRates[level] - commissionRates[len(commissionRates)-1]) * totalPrice
			} else {
				if i == 0 {
					commissionAmount = commissionRates[level] * totalPrice
				} else {
					commissionAmount = (commissionRates[level] - commissionRates[level-1]) * totalPrice
				}
			}

			if commissionAmount <= 0 {
				continue
			}

			_, err := qtx.CreateCommission(context.Background(), db.CreateCommissionParams{
				OrderID:     pgtype.UUID{Bytes: orderID, Valid: true},
				AffiliateID: affililates[i].ID,
				Amount:      commissionAmount,
			})
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create commission"})
				return
			}

			err = qtx.AddAffiliateBalance(context.Background(), db.AddAffiliateBalanceParams{
				ID:      affililates[i].ID,
				Balance: commissionAmount,
			})
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add affiliate balance"})
				return
			}
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
