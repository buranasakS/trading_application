package handlers

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"
)

type CommissionAffiliateDetail struct {
	AffiliateID   pgtype.UUID `json:"affiliate_id"`
	AffiliateName string      `json:"affiliate_name"`
	Commission    float64     `json:"commission"`
}

type CommsisionDistributionResponse struct {
	OrderID         pgtype.UUID                 `json:"order_id"`
	TotalCommission float64                     `json:"total_commission"`
	Details         []CommissionAffiliateDetail `json:"details"`
}

// ListCommissionsHandler godoc
// @Summary      List all commissions
// @Description  List all commissions
// @Tags         Commissions
// @Security BearerAuth
// @Accept       json
// @Produce      json
// @Success      200  {object}  []db.Commission "List of commissions"
// @Failure 401 {object} handlers.ErrorResponse "Unauthorized"
// @Router       /commissions/list [get]
func (h *Handler) ListCommissionsHandler(c *gin.Context) {
	commissions, err := h.db.ListCommissions(context.Background())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch commissions"})
		return
	}

	c.JSON(http.StatusOK, commissions)
}

// GetCommissionByIDHandler godoc
// @Summary      Get commission by ID
// @Description  Get commission by ID
// @Tags         Commissions
// @Security BearerAuth
// @Accept       json
// @Produce      json
// @Param        id path string true "Commission ID"
// @Success      200 {object} db.Commission	"Commission details"
// @Failure 401 {object} handlers.ErrorResponse "Unauthorized"
// @Router       /commissions/{id} [get]
func (h *Handler) GetCommissionDetailHandler(c *gin.Context) {
	var commissionId pgtype.UUID
	if err := commissionId.Scan(c.Param("id")); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid commission ID"})
		return
	}

	commission, err := h.db.GetCommissionByID(context.Background(), commissionId)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Commission not found"})
		return
	}

	c.JSON(http.StatusOK, commission)
}



// GetCommissionDistributionHandler godoc
// @Summary      Get commission by Order ID 
// @Description  Get commission by Order ID
// @Tags         Commissions
// @Security BearerAuth
// @Accept       json
// @Produce      json
// @Param        order_id path string true "Order ID"
// @Success      200 {object} CommsisionDistributionResponse	"Commission by order id details"
// @Failure 401 {object} handlers.ErrorResponse "Unauthorized"
// @Router       /commissions/distribution/{order_id} [get]
func (h *Handler) GetCommissionDistributionHandler(c *gin.Context) {
	var orderId pgtype.UUID

	orderIDStr := c.Param("order_id")
	if err := orderId.Scan(orderIDStr); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order ID"})
		return
	}

	totalCommission, err := h.db.GetTotalCommission(context.Background(), orderId)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Order not found or no commission available"})
		return
	}

	commissions, err := h.db.GetCommissionByOrderID(context.Background(), orderId)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to fetch commissions"})
		return
	}

	var commissionAffiliateDetails []CommissionAffiliateDetail
	for _, commission := range commissions {
		commissionAffiliateDetails = append(commissionAffiliateDetails, CommissionAffiliateDetail{
			AffiliateID:   commission.ID,
			AffiliateName: commission.Name,
			Commission:    commission.Amount,
		})
	}

	c.JSON(http.StatusOK, CommsisionDistributionResponse{OrderID: orderId, TotalCommission: float64(totalCommission), Details: commissionAffiliateDetails})
}



