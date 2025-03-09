package handlers

import (
	"context"
	"net/http"

	db "github.com/buranasakS/trading_application/db/sqlc"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"
)

type RequestAffiliate struct {
	Name            string      `json:"name" binding:"required"`
	MasterAffiliate pgtype.UUID `json:"master_id"`
}

// CreateAffiliateHandler godoc
// @Summary      Create a new affiliate
// @Description  Create a new affiliate
// @Tags         Affiliates
// @Security BearerAuth
// @Accept       json
// @Produce      json
// @Param        request body   RequestAffiliate true "Affiliate details"
// @Success      201  {object}  db.Affiliate "Affiliate created successfully"
// @Failure 401 {object} handlers.ErrorResponse "Unauthorized"
// @Router       /affiliates [post]
func (h *Handler) CreateAffiliateHandler(c *gin.Context) {
	var req RequestAffiliate
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	affiliate, err := h.db.CreateAffiliate(context.Background(), db.CreateAffiliateParams{
		Name:            req.Name,
		MasterAffiliate: req.MasterAffiliate,
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create affiliate"})
		return
	}

	c.JSON(http.StatusCreated, affiliate)
}

// ListAffiliatesHandler godoc
// @Summary      List all affiliates
// @Description  List all affiliates
// @Tags         Affiliates
// @Security BearerAuth
// @Accept       json
// @Produce      json
// @Success      200  {object}  []db.Affiliate "List of affiliates"
// @Failure 401 {object} handlers.ErrorResponse "Unauthorized"
// @Router       /affiliates/list [get]
func (h *Handler) ListAffiliatesHandler(c *gin.Context) {
	affiliates, err := h.db.ListAffiliates(context.Background())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch affiliates"})
		return
	}

	c.JSON(http.StatusOK, affiliates)
}

// GetAffiliateByIDHandler godoc
// @Summary      Get affiliate by ID
// @Description  Get affiliate by ID
// @Tags         Affiliates
// @Security BearerAuth
// @Accept       json
// @Produce      json
// @Param        id path string true "Affiliate ID"
// @Success      200 {object} db.Affiliate	"Affiliate details"
// @Failure 401 {object} handlers.ErrorResponse "Unauthorized"
// @Router       /affiliates/{id} [get]
func (h *Handler) GetAffiliateDetailHandler(c *gin.Context) {
	var affiliateId pgtype.UUID
	if err := affiliateId.Scan(c.Param("id")); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid affiliate ID"})
		return
	}

	affiliate, err := h.db.GetAffiliateByID(context.Background(), affiliateId)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Affiliate not found"})
		return
	}

	c.JSON(http.StatusOK, affiliate)
}
