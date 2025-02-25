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

func UserBuyProductHandler(c *gin.Context) {
	var req OrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}

	queries := db.New(config.ConnectDatabase().DB)
	tx, err := config.ConnectDatabase().DB.BeginTx(context.Background(), pgx.TxOptions{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"Error": "Failed to start transaction"})
		return
	}
	defer tx.Rollback(context.Background())

	qtx := queries.WithTx(tx)

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
		affililates := []db.Affiliate{}
		currentID := user.AffiliateID

		for currentID.Valid {
			affiliate, err := qtx.GetAffiliateByID(context.Background(), currentID)
			if err != nil || !affiliate.ID.Valid {
				break
			}

			affililates = append(affililates, affiliate)
			currentID = affiliate.MasterAffiliate
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

	c.JSON(http.StatusOK, gin.H{
		"status":     "success",
		"message":    "Purchase completed",
		"order_id":   orderID,
		"total_cost": totalPrice,
	})
}
