package handlers

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"
)

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
