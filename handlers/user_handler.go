package handlers

import (
	"context"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/buranasakS/trading_application/config"
	db "github.com/buranasakS/trading_application/db/sqlc"
	"github.com/buranasakS/trading_application/helpers"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"golang.org/x/crypto/bcrypt"
)

type RequestUserLogin struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type RequestUserRegister struct {
	Username    string      `json:"username" binding:"required"`
	Password    string      `json:"password" binding:"required"`
	AffiliateID pgtype.UUID `json:"affiliate_id" binding:"required"`
}
type ResponseUser struct {
	Page       int32   `json:"page"`
	TotalPage  int32   `json:"total_page"`
	Count      int32   `json:"count"`
	TotalCount int32   `json:"total_count"`
	Data       []Users `json:"data"`
}
type Users struct {
	ID          pgtype.UUID `json:"id"`
	Username    string      `json:"username"`
	Balance     float64     `json:"balance"`
	AffiliateID pgtype.UUID `json:"affiliate_id"`
}

type RequestAmount struct {
	Amount float64 `json:"amount" binding:"required"`
}

func (h *Handler) LoginUserHandler(c *gin.Context) {
	var req RequestUserLogin
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.db.GetUserByUsernameForLogin(context.Background(), req.Username)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid username or password"})
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid password"})
		return
	}

	secretKey := os.Getenv("SECRET_KEY")
	if secretKey == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Missing token key"})
		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":      user.ID,
		"exp":      time.Now().Add(time.Hour * 72).Unix(),
		"iat":      time.Now().Unix(),
		"username": user.Username,
	})

	tokenString, err := token.SignedString([]byte(os.Getenv(secretKey)))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"Error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token": tokenString,
	})
}

// RegisterUserHandler godoc
// @Summary      register a new user
// @Description  register a new user
// @Tags         Users
// @Accept       json
// @Produce      json
// @Param        request body      db.CreateUserParams true "User details"
// @Success      201  {object}  db.User "User created successfully"
// @Router       /users [post]
func (h *Handler) RegisterUserHandler(c *gin.Context) {
	var req RequestUserRegister
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Username == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing username"})
		return
	}

	if req.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing password"})
		return
	}

	hashedPassword, err := helpers.HashedPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	user, err := h.db.CreateUser(context.Background(), db.CreateUserParams{
		Username:    req.Username,
		Password:    hashedPassword,
		AffiliateID: req.AffiliateID,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	c.JSON(http.StatusCreated, Users{
		ID:          user.ID,
		Username:    user.Username,
		Balance:     user.Balance,
		AffiliateID: user.AffiliateID,
	})
}

// ListUsersHandler godoc
// @Summary      List all users with pagination
// @Description  Fetch a paginated list of users from the database
// @Tags         Users
// @Security BearerAuth
// @Accept       json
// @Produce      json
// @Param        limit  query   int  false  "Number of users per page (default 10)"
// @Param        page   query   int  false  "Page number (default 1)"
// @Success      200  {object}  ResponseUser
// @Failure 401 {object} handlers.ErrorResponse
// @Router       /users/all [get]
func (h *Handler) ListUsersHandler(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "10")
	pageStr := c.DefaultQuery("page", "1")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid limit value. Must be a positive integer."})
		return
	}

	page, err := strconv.Atoi(pageStr)
	if err != nil || page <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid page value. Must be a positive integer."})
		return
	}

	if limit < 1 {
		limit = 10
	}
	if page < 1 {
		page = 1
	}

	offset := (page - 1) * limit

	userRows, err := h.db.ListUsers(context.Background(), db.ListUsersParams{
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch users"})
		return
	}

	var users []Users
	for _, userRow := range userRows {
		users = append(users, Users{
			ID:          userRow.ID,
			Username:    userRow.Username,
			Balance:     userRow.Balance,
			AffiliateID: userRow.AffiliateID,
		})
	}

	totalCount := len(users)
	totalPage := (int32(totalCount) + int32(limit) - 1) / int32(limit)

	c.JSON(http.StatusOK, ResponseUser{
		Page:       int32(page),
		TotalPage:  totalPage,
		Count:      int32(len(users)),
		TotalCount: int32(totalCount),
		Data:       users,
	})
}

// GetUserDetailByIDHandler godoc
// @Summary      Get user details by ID
// @Description  Retrieve user details by their unique ID
// @Tags         Users
// @Security BearerAuth
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "User ID (UUID)"
// @Success      200  {object}  db.User
// @Failure 401 {object} handlers.ErrorResponse
// @Router       /users/{id} [get]
func (h *Handler) GetUserDetailHandler(c *gin.Context) {
	var userId pgtype.UUID
	if err := userId.Scan(c.Param("id")); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid User ID"})
		return
	}

	user, err := h.db.GetUserDetailByID(context.Background(), userId)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, user)
}

// DeductUserBalanceHandler godoc
// @Summary Deduct user balance
// @Description deduct balance from user account
// @Tags Users
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Param request body RequestAmount true "Amount to deduct"
// @Success      200  {object}  map[string]string "Balance deducted successfully"
// @Failure 401 {object} handlers.ErrorResponse
// @Router /users/deduct/balance/{id} [patch]
func (h *Handler) DeductUserBalanceHandler(c *gin.Context) {
	var userId pgtype.UUID
	if err := userId.Scan(c.Param("id")); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid User ID"})
		return
	}

	user, err := h.db.GetUserDetailByID(context.Background(), userId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "User not found"})
		return
	}

	var req RequestAmount
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body, 'amount' should be a positive number"})
		return
	}

	if req.Amount <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Amount must be more than 0"})
		return
	}

	tx, err :=config.ConnectDatabase().DB.BeginTx(context.Background(), pgx.TxOptions{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start transaction"})
		return
	}

	defer func() {
		if err != nil {
			tx.Rollback(context.Background())
		}
	}()

	qtx := h.db.(*db.Queries).WithTx(tx)

	result, err := qtx.DeductUserBalance(context.Background(), db.DeductUserBalanceParams{
		Balance: req.Amount,
		ID:      userId,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"Error": "Failed to deduct balance"})
		return
	}

	if result == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Insufficient balance"})
		return
	}

	err = tx.Commit(context.Background())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction"})
		return
	}


	c.JSON(http.StatusOK, user)
}

// AddUserBalanceHandler godoc
// @Summary Add user balance
// @Description add balance to user account
// @Tags Users
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Param request body RequestAmount true "Amount to add"
// @Success      200  {object}  map[string]string "Balance added successfully"
// @Failure 401 {object} handlers.ErrorResponse
// @Router /users/add/balance/{id} [patch]
func (h *Handler) AddUserBalanceHandler(c *gin.Context) {
	var userId pgtype.UUID
	if err := userId.Scan(c.Param("id")); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid User ID"})
		return
	}

	user, err := h.db.GetUserDetailByID(context.Background(), userId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "User not found"})
		return
	}

	var req RequestAmount
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body, 'amount' should be a positive number"})
		return
	}

	if req.Amount <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Amount must be more than 0"})
		return
	}

	tx, err := config.ConnectDatabase().DB.BeginTx(context.Background(), pgx.TxOptions{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start transaction"})
		return
	}
	defer tx.Rollback(context.Background())

	qtx := h.db.(*db.Queries).WithTx(tx)

	result, err := qtx.AddUserBalance(context.Background(), db.AddUserBalanceParams{
		Balance: req.Amount,
		ID:      userId,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update balance"})
		return
	}

	if result == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Insufficient balance"})
		return
	}

	err = tx.Commit(context.Background())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction"})
		return
	}

	c.JSON(http.StatusOK, user)
}
