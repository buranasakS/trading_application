package handlers

import (
	"context"
	"net/http"

	"github.com/buranasakS/trading_application/config"
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
// @Accept       json
// @Produce      json
// @Param        request body   RequestAffiliate true "Affiliate details"
// @Success      201  {object}  db.Affiliate "Affiliate created successfully"
// @Router       /affiliates [post]
func CreateAffiliateHandler(c *gin.Context) {
	var req RequestAffiliate
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	queries := db.New(config.ConnectDatabase().DB)
	affiliate, err := queries.CreateAffiliate(context.Background(), db.CreateAffiliateParams{
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
// @Accept       json
// @Produce      json
// @Success      200  {object}  []db.Affiliate "List of affiliates"
// @Router       /affiliates/list [get]
func ListAffiliatesHandler(c *gin.Context) {
	queries := db.New(config.ConnectDatabase().DB)

	affiliates, err := queries.ListAffiliates(context.Background())
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
// @Accept       json
// @Produce      json
// @Param        id path string true "Affiliate ID"
// @Success      200 {object} db.Affiliate	"Affiliate details"
// @Router       /affiliates/{id} [get]
func GetAffiliateByIDHandler(c *gin.Context) {
	var affiliateId pgtype.UUID
	if err := affiliateId.Scan(c.Param("id")); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid affiliate ID"})
		return
	}

	queries := db.New(config.ConnectDatabase().DB)
	affiliate, err := queries.GetAffiliateByID(context.Background(), affiliateId)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Affiliate not found"})
		return
	}

	c.JSON(http.StatusOK, affiliate)
}
