package handlers

import (
	db "github.com/buranasakS/trading_application/db/sqlc"
)

type Handler struct {
	db db.Querier
}

func NewHandler(db db.Querier) *Handler {
	return &Handler{db: db}
}

type ErrorResponse struct {
	Error string `json:"error"`
}

