package handlers

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	mockdb "github.com/buranasakS/trading_application/db/mocks"
	db "github.com/buranasakS/trading_application/db/sqlc"
	"github.com/buranasakS/trading_application/helpers"
	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

type responseUserDetail struct {
	ID          pgtype.UUID `json:"id"`
	Username    string      `json:"username"`
	Balance     float64     `json:"balance"`
	AffiliateID pgtype.UUID `json:"affiliate_id"`
}

func TestLoginUserHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	userId := helpers.PgtypeUUID(t, "123e4567-e89b-12d3-a456-426614174000")
	hashPassword, errHash := helpers.HashedPassword("password123")
	require.NoError(t, errHash)

	tests := []struct {
		name           string
		reqBody        interface{}
		mockUser       *db.GetUserByUsernameForLoginRow
		mockError      error
		expectedStatus int
		secretKey      string
		expectedBody   string
	}{
		{
			name: "Success",
			reqBody: RequestUserLogin{
				Username: "testuser",
				Password: "password123",
			},
			mockUser: &db.GetUserByUsernameForLoginRow{
				ID:       pgtype.UUID{Bytes: uuid.New(), Valid: true},
				Username: "testuser",
				Password: hashPassword,
			},
			mockError:      nil,
			expectedStatus: http.StatusOK,
			secretKey:      "SECRET_KEY",
			expectedBody:   `"token":`,
		},
		{
			name:           "Invalid JSON",
			reqBody:        `{"username": "testuser", "password": 123}`,
			mockUser:       nil,
			mockError:      nil,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `"error":`,
		},
		{
			name: "User Not Found",
			reqBody: RequestUserLogin{
				Username: "wronguser",
				Password: "password123",
			},
			mockUser:       nil,
			mockError:      sql.ErrNoRows,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `"Invalid username or password"`,
		},
		{
			name: "Incorrect Password",
			reqBody: RequestUserLogin{
				Username: "testuser",
				Password: "wrongpassword",
			},
			mockUser: &db.GetUserByUsernameForLoginRow{
				ID:       userId,
				Username: "testuser",
				Password: hashPassword,
			},
			mockError:      errors.New("Invalid username or password"),
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Invalid username or password",
		},
		{
			name: "Failed to Generate Token",
			reqBody: RequestUserLogin{
				Username: "testuser",
				Password: "password123",
			},
			mockUser: &db.GetUserByUsernameForLoginRow{
				ID:       userId,
				Username: "testuser",
				Password: hashPassword,
			},
			mockError:      nil,
			secretKey:      "",
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `"Missing token key"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockDB := mockdb.NewMockQuerier(ctrl)

			if tt.mockUser != nil || tt.mockError != nil {
				mockDB.EXPECT().GetUserByUsernameForLogin(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, username string) (db.GetUserByUsernameForLoginRow, error) {
						if tt.mockUser != nil {
							return *tt.mockUser, tt.mockError
						}
						return db.GetUserByUsernameForLoginRow{}, tt.mockError
					}).
					Times(1)
			}

			os.Setenv("SECRET_KEY", tt.secretKey)

			gin.SetMode(gin.TestMode)
			router := gin.New()
			router.POST("/login", func(c *gin.Context) {
				NewHandler(mockDB).LoginUserHandler(c)
			})

			body, err := json.Marshal(tt.reqBody)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, req)

			require.Equal(t, tt.expectedStatus, recorder.Code)

			responseBody, err := io.ReadAll(recorder.Body)
			require.NoError(t, err)
			require.Contains(t, string(responseBody), tt.expectedBody)
		})
	}

}

func TestRegisterUserHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mockdb.NewMockQuerier(ctrl)

	userId := helpers.PgtypeUUID(t, "123e4567-e89b-12d3-a456-426614174000")
	affiliateId := helpers.PgtypeUUID(t, "123e4567-e89b-12d3-a456-426614174001")

	tests := []struct {
		name           string
		reqBody        interface{}
		mockReturnData db.User
		mockReturnErr  error
		expectedStatus int
		expectedError  string
	}{
		{
			name: "Success",
			reqBody: RequestUserRegister{
				Username:    "testuser",
				Password:    "testpassword",
				AffiliateID: affiliateId,
			},
			mockReturnData: db.User{
				ID:          userId,
				Username:    "testuser",
				Balance:     0,
				AffiliateID: affiliateId,
			},
			mockReturnErr:  nil,
			expectedStatus: http.StatusCreated,
			expectedError:  "",
		},
		{
			name: "Invalid Request Missing Username",
			reqBody: RequestUserRegister{
				Password:    "testpassword",
				AffiliateID: affiliateId,
			},
			mockReturnData: db.User{},
			mockReturnErr:  nil,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Key: 'RequestUserRegister.Username' Error:Field validation for 'Username' failed on the 'required' tag",
		},

		{
			name: "Invalid Request Missing Password",
			reqBody: RequestUserRegister{
				Username:    "testuser",
				AffiliateID: affiliateId,
			},
			mockReturnData: db.User{},
			mockReturnErr:  nil,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Key: 'RequestUserRegister.Password' Error:Field validation for 'Password' failed on the 'required' tag",
		},
		{
			name: "Failed to create user",
			reqBody: RequestUserRegister{
				Username:    "testuser",
				Password:    "testpass",
				AffiliateID: affiliateId,
			},
			mockReturnErr:  errors.New("Failed to create user"),
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "Failed to create user",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var body []byte
			body, err := json.Marshal(tt.reqBody)
			require.NoError(t, err)

			if tt.expectedStatus != http.StatusBadRequest {
				req := tt.reqBody.(RequestUserRegister)
				mockDB.EXPECT().CreateUser(
					gomock.Any(),
					gomock.Any(),
				).DoAndReturn(func(_ context.Context, params db.CreateUserParams) (db.User, error) {
					err := bcrypt.CompareHashAndPassword([]byte(params.Password), []byte(req.Password))
					require.NoError(t, err)
					return tt.mockReturnData, tt.mockReturnErr
				}).Times(1)
			}

			gin.SetMode(gin.TestMode)
			router := gin.New()
			router.POST("/register", func(c *gin.Context) {
				NewHandler(mockDB).RegisterUserHandler(c)
			})

			req, err := http.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(body))
			require.NoError(t, err)

			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, req)

			require.Equal(t, tt.expectedStatus, recorder.Code)

			if tt.expectedError != "" {
				var response map[string]string
				err := json.Unmarshal(recorder.Body.Bytes(), &response)
				require.NoError(t, err)
				require.Equal(t, tt.expectedError, response["error"])
			} else {
				var response db.User
				err := json.Unmarshal(recorder.Body.Bytes(), &response)
				require.NoError(t, err)
				require.Equal(t, tt.mockReturnData, response)
			}
		})
	}
}

func TestListUsersHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mockdb.NewMockQuerier(ctrl)

	tests := []struct {
		name           string
		queryParams    string
		mockReturn     []db.ListUsersRow
		mockError      error
		expectedStatus int
		expectedCount  int
		expectedBody   interface{}
	}{
		{
			name:        "Success",
			queryParams: "limit=2&page=1",
			mockReturn: []db.ListUsersRow{
				{
					ID:          pgtype.UUID{Bytes: uuid.New(), Valid: true},
					Username:    "user1",
					Balance:     100,
					AffiliateID: pgtype.UUID{Bytes: uuid.New(), Valid: true},
				},
				{
					ID:          pgtype.UUID{Bytes: uuid.New(), Valid: true},
					Username:    "user2",
					Balance:     200,
					AffiliateID: pgtype.UUID{Bytes: uuid.New(), Valid: true},
				},
			},
			mockError:      nil,
			expectedStatus: http.StatusOK,
			expectedBody: []db.ListUsersRow{
				{
					ID:          pgtype.UUID{Bytes: uuid.New(), Valid: true},
					Username:    "user1",
					Balance:     100,
					AffiliateID: pgtype.UUID{Bytes: uuid.New(), Valid: true},
				},
				{
					ID:          pgtype.UUID{Bytes: uuid.New(), Valid: true},
					Username:    "user3",
					Balance:     200,
					AffiliateID: pgtype.UUID{Bytes: uuid.New(), Valid: true},
				},
			},
			expectedCount: 2,
		},
		{
			name:           "Invalid Limit",
			queryParams:    "limit=-1&page=1",
			mockReturn:     nil,
			mockError:      errors.New("Invalid limit value. Must be a positive integer."),
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Invalid Page",
			queryParams:    "limit=2&page=-1",
			mockReturn:     nil,
			mockError:      errors.New("Invalid page value. Must be a positive integer."),
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Failed to fetch users",
			queryParams:    "limit=2&page=1",
			mockReturn:     nil,
			mockError:      errors.New("Failed to fetch users"),
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.expectedStatus == http.StatusOK || tc.expectedStatus == http.StatusInternalServerError {
				mockDB.EXPECT().
					ListUsers(gomock.Any(), gomock.Any()).
					Return(tc.mockReturn, tc.mockError).
					Times(1)
			}

			gin.SetMode(gin.TestMode)
			router := gin.New()
			router.GET("/users/all", func(c *gin.Context) {
				NewHandler(mockDB).ListUsersHandler(c)
			})

			req, err := http.NewRequest(http.MethodGet, "/users/all?"+tc.queryParams, nil)
			require.NoError(t, err)

			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, req)

			assert.Equal(t, tc.expectedStatus, recorder.Code)

			if tc.expectedStatus == http.StatusOK {
				var response ResponseUser
				err := json.Unmarshal(recorder.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedCount, len(response.Data))
			}
		})
	}
}

func TestGetUserDetail(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mockdb.NewMockQuerier(ctrl)

	userId := helpers.PgtypeUUID(t, "123e4567-e89b-12d3-a456-426614174000")
	affiliateId := helpers.PgtypeUUID(t, "123e4567-e89b-12d3-a456-426614174001")

	user := responseUserDetail{
		ID:          userId,
		Username:    "testuser",
		Balance:     100,
		AffiliateID: affiliateId,
	}

	tests := []struct {
		name           string
		userID         string
		mockReturn     responseUserDetail
		mockError      error
		expectedStatus int
		expectedBody   interface{}
	}{
		{
			name:           "Success",
			userID:         affiliateId.String(),
			mockReturn:     user,
			mockError:      nil,
			expectedStatus: http.StatusOK,
			expectedBody:   user,
		},
		{
			name:           "Invalid User ID",
			userID:         "invalid-uuid",
			mockReturn:     responseUserDetail{},
			mockError:      nil,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   map[string]string{"error": "Invalid User ID"},
		},
		{
			name:           "User Not Found",
			userID:         uuid.New().String(),
			mockError:      sql.ErrNoRows,
			expectedStatus: http.StatusNotFound,
			expectedBody:   map[string]string{"error": "User not found"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.expectedStatus != http.StatusBadRequest {
				mockUserID := pgtype.UUID{}
				err := mockUserID.Scan(tc.userID)
				require.NoError(t, err)
				mockDB.EXPECT().
					GetUserDetailByID(gomock.Any(), mockUserID).
					Return(db.GetUserDetailByIDRow{
						ID:          tc.mockReturn.ID,
						Username:    tc.mockReturn.Username,
						Balance:     tc.mockReturn.Balance,
						AffiliateID: tc.mockReturn.AffiliateID}, tc.mockError).Times(1)
			}

			gin.SetMode(gin.ReleaseMode)
			router := gin.New()
			router.GET("/users/:id", func(c *gin.Context) {
				NewHandler(mockDB).GetUserDetailHandler(c)
			})

			req, err := http.NewRequest(http.MethodGet, "/users/"+tc.userID, nil)
			require.NoError(t, err)
			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, req)

			require.Equal(t, tc.expectedStatus, recorder.Code)

			if tc.expectedStatus == http.StatusOK {
				var response db.User
				err := json.Unmarshal(recorder.Body.Bytes(), &response)
				require.NoError(t, err)
				require.Equal(t, tc.mockReturn.Username, response.Username)
			} else {
				var response map[string]string
				err := json.Unmarshal(recorder.Body.Bytes(), &response)
				require.NoError(t, err)
				require.Equal(t, tc.expectedBody, response)
			}
		})
	}
}

func TestDeductUserBalanceHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	userId := helpers.PgtypeUUID(t, "123e4567-e89b-12d3-a456-426614174000")

	tests := []struct {
		name           string
		userId         string
		reqBody        interface{}
		mockUserDetail db.GetUserDetailByIDRow
		mockUserErr    error
		mockDeductErr  error
		mockDeductRows int64
		expectedStatus int
		expectedBody   interface{}
	}{
		{
			name:   "Success",
			userId: userId.String(),
			reqBody: RequestAmount{
				Amount: 100.0,
			},
			mockUserErr:    nil,
			mockDeductErr:  nil,
			mockDeductRows: 1,
			expectedStatus: http.StatusOK,
			expectedBody:   nil,
		},
		{
			name:           "Invalid User ID",
			userId:         "invalid-uuid",
			reqBody:        RequestAmount{Amount: 100.0},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   gin.H{"error": "Invalid User ID"},
		},
		{
			name:   "User Not Found",
			userId: userId.String(),
			reqBody: RequestAmount{
				Amount: 100.0,
			},
			mockUserErr:    sql.ErrNoRows,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   gin.H{"error": "User not found"},
		},
		{
			name:   "Insufficient Balance",
			userId: userId.String(),
			reqBody: RequestAmount{
				Amount: 1000.0,
			},
			mockUserDetail: db.GetUserDetailByIDRow{
				ID:       userId,
				Username: "testuser",
				Balance:  500.0,
			},
			mockUserErr:    nil,
			mockDeductErr:  nil,
			mockDeductRows: 0,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   gin.H{"error": "Insufficient balance"},
		},
		{
			name:           "Invalid Request Body",
			userId:         userId.String(),
			reqBody:        "invalid json",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   gin.H{"error": "Invalid request body, 'amount' should be a positive number"},
		},
		{
			name:   "Zero Amount",
			userId: userId.String(),
			reqBody: map[string]interface{}{
				"amount": 0,
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   gin.H{"error": "Invalid request body, 'amount' should be a positive number"},
		},
		{
			name:   "Negative Amount",
			userId: userId.String(),
			reqBody: map[string]interface{}{
				"amount": -100,
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   gin.H{"error": "Amount must be more than 0"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := mockdb.NewMockQuerier(ctrl)

			if _, err := uuid.Parse(tt.userId); err == nil {
				mockDB.EXPECT().
					GetUserDetailByID(gomock.Any(), gomock.Any()).
					Return(tt.mockUserDetail, tt.mockUserErr).
					Times(1)

				if tt.mockUserErr == nil && tt.reqBody != "invalid json" {
					if amount, ok := tt.reqBody.(RequestAmount); ok && amount.Amount > 0 {
						mockDB.EXPECT().
							DeductUserBalance(gomock.Any(), db.DeductUserBalanceParams{
								Balance: amount.Amount,
								ID:      helpers.PgtypeUUID(t, tt.userId),
							}).
							Return(tt.mockDeductRows, tt.mockDeductErr).
							Times(1)
					}
				}
			}

			router := gin.New()
			router.PATCH("/users/deduct/balance/:id", func(c *gin.Context) {
				NewHandler(mockDB).DeductUserBalanceHandler(c)
			})

			body, err := json.Marshal(tt.reqBody)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPatch, "/users/deduct/balance/"+tt.userId, bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, req)

			require.Equal(t, tt.expectedStatus, recorder.Code)

			if tt.expectedBody != nil {
				var response gin.H
				err = json.Unmarshal(recorder.Body.Bytes(), &response)
				require.NoError(t, err)
				require.Equal(t, tt.expectedBody, response)
			} else {
				var response db.GetUserDetailByIDRow
				err = json.Unmarshal(recorder.Body.Bytes(), &response)
				require.NoError(t, err)
				require.Equal(t, tt.mockUserDetail, response)
			}
		})
	}
}

func TestAddUserBalanceHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	userId := helpers.PgtypeUUID(t, "123e4567-e89b-12d3-a456-426614174000")

	tests := []struct {
		name           string
		userId         string
		reqBody        interface{}
		mockUserDetail db.GetUserDetailByIDRow
		mockUserErr    error
		mockAddErr     error
		mockAddRows    int64
		expectedStatus int
		expectedBody   interface{}
	}{
		{
			name:   "Success",
			userId: userId.String(),
			reqBody: RequestAmount{
				Amount: 100.0,
			},
			mockUserErr:   nil,
			mockAddErr:    nil,
			mockAddRows:   1,
			expectedStatus: http.StatusOK,
			expectedBody:    gin.H{"message": "Add balance completed"},
		},
		{
			name:           "Invalid User ID",
			userId:         "invalid-uuid",
			reqBody:        RequestAmount{Amount: 100.0},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   gin.H{"error": "Invalid User ID"},
		},
		{
			name:   "User Not Found",
			userId: userId.String(),
			reqBody: RequestAmount{
				Amount: 100.0,
			},
			mockUserErr:   sql.ErrNoRows,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   gin.H{"error": "User not found"},
		},
		{
			name:   "Insufficient Balance",
			userId: userId.String(),
			reqBody: RequestAmount{
				Amount: 1000.0,
			},
			mockUserDetail: db.GetUserDetailByIDRow{
				ID:       userId,
				Username: "testuser",
				Balance:  500.0,
			},
			mockUserErr:    nil,
			mockAddErr:     nil,
			mockAddRows:    0,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   gin.H{"error": "Insufficient balance"},
		},
		{
			name:           "Invalid Request Body",
			userId:         userId.String(),
			reqBody:        "invalid json",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   gin.H{"error": "Invalid request body, 'amount' should be a positive number"},
		},
		{
			name:   "Zero Amount",
			userId: userId.String(),
			reqBody: map[string]interface{}{
				"amount": 0,
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   gin.H{"error": "Invalid request body, 'amount' should be a positive number"},
		},
		{
			name:   "Negative Amount",
			userId: userId.String(),
			reqBody: map[string]interface{}{
				"amount": -100,
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   gin.H{"error": "Amount must be more than 0"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := mockdb.NewMockQuerier(ctrl)

			if _, err := uuid.Parse(tt.userId); err == nil {
				mockDB.EXPECT().
					GetUserDetailByID(gomock.Any(), gomock.Any()).
					Return(tt.mockUserDetail, tt.mockUserErr).
					Times(1)

				if tt.mockUserErr == nil && tt.reqBody != "invalid json" {
					if amount, ok := tt.reqBody.(RequestAmount); ok && amount.Amount > 0 {
						mockDB.EXPECT().
							AddUserBalance(gomock.Any(), db.AddUserBalanceParams{
								Balance: amount.Amount,
								ID:      helpers.PgtypeUUID(t, tt.userId),
							}).
							Return(tt.mockAddRows, tt.mockAddErr).
							Times(1)
					}
				}
			}

			router := gin.New()
			router.PATCH("/users/add/balance/:id", func(c *gin.Context) {
				NewHandler(mockDB).AddUserBalanceHandler(c)
			})

			body, err := json.Marshal(tt.reqBody)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPatch, "/users/add/balance/"+tt.userId, bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, req)

			require.Equal(t, tt.expectedStatus, recorder.Code)

			if tt.expectedBody != nil {
				var response gin.H
				err = json.Unmarshal(recorder.Body.Bytes(), &response)
				require.NoError(t, err)
				require.Equal(t, tt.expectedBody, response)
			} else {
				var response db.GetUserDetailByIDRow
				err = json.Unmarshal(recorder.Body.Bytes(), &response)
				require.NoError(t, err)
				require.Equal(t, tt.mockUserDetail, response)
			}
		})
	}
}

