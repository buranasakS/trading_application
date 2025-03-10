package handlers

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	mockdb "github.com/buranasakS/trading_application/db/mocks"
	db "github.com/buranasakS/trading_application/db/sqlc"
	"github.com/buranasakS/trading_application/helpers"
	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"
)

func TestUserOrderProductHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mockdb.NewMockQuerier(ctrl)

	userId := helpers.PgtypeUUID(t, "123e4567-e89b-12d3-a456-426614174000")
	productId := helpers.PgtypeUUID(t, "123e4567-e89b-12d3-a456-426614174001")
	affiliateId := helpers.PgtypeUUID(t, "123e4567-e89b-12d3-a456-426614174002")
	masterAffiliateId := helpers.PgtypeUUID(t, "123e4567-e89b-12d3-a456-426614174003")

	tests := []struct {
		name                    string
		reqBody                 interface{}
		mockUser                db.GetUserDetailByIDRow
		mockUserErr             error
		mockProduct             db.Product
		mockProductErr          error
		mockDeductUserRows      int64
		mockDeductUserErr       error
		mockDeductProductRows   int64
		mockDeductProductErr    error
		mockAffiliateList       []db.Affiliate
		mockAffiliateErr        error
		mockCreateCommissionErr error
		mockAddBalanceErr       error
		expectedStatus          int
		expectedError           string
		mockCommitErr           error
	}{
		{
			name: "Invalid Quantity (Zero)",
			reqBody: OrderRequest{
				UserID:    userId,
				ProductID: productId,
				Quantity:  0,
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Quantity must be more than 0",
		},
		{
			name: "Invalid Quantity (Negative)",
			reqBody: OrderRequest{
				UserID:    userId,
				ProductID: productId,
				Quantity:  -1,
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Quantity must be more than 0",
		},
		{
			name:           "Invalid JSON",
			reqBody:        `{user_id: "invalid-uuid", "product_id": "invalid-uuid", "quantity": "invalid"}`,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid character",
		},
		{
			name: "User not found",
			reqBody: OrderRequest{
				UserID:    userId,
				ProductID: productId,
				Quantity:  1,
			},
			mockUserErr:    sql.ErrNoRows,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "User not found",
		},
		{
			name: "Product Not Found",
			reqBody: OrderRequest{
				UserID:    userId,
				ProductID: productId,
				Quantity:  1,
			},
			mockUser: db.GetUserDetailByIDRow{
				ID:       userId,
				Username: "testuser",
				Balance:  1000.0,
			},
			mockUserErr:    nil,
			mockProductErr: sql.ErrNoRows,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Product not found",
		},
		{
			name: "Success with no affiliate",
			reqBody: OrderRequest{
				UserID:    userId,
				ProductID: productId,
				Quantity:  1,
			},
			mockUser: db.GetUserDetailByIDRow{
				ID:          userId,
				Username:    "testuser",
				Balance:     1000.0,
				AffiliateID: pgtype.UUID{Valid: false},
			},
			mockUserErr: nil,
			mockProduct: db.Product{
				ID:       productId,
				Name:     "testproduct",
				Quantity: 100,
				Price:    100.0,
			},
			mockProductErr:        nil,
			mockDeductUserRows:    1,
			mockDeductUserErr:     nil,
			mockDeductProductRows: 1,
			mockDeductProductErr:  nil,
			expectedStatus:        http.StatusCreated,
			expectedError:         "",
		},
		{
			name: "Success with calculate affiliate commission ",
			reqBody: OrderRequest{
				UserID:    userId,
				ProductID: productId,
				Quantity:  1,
			},
			mockUser: db.GetUserDetailByIDRow{
				ID:          userId,
				Username:    "testuser",
				Balance:     1000.0,
				AffiliateID: affiliateId,
			},
			mockUserErr: nil,
			mockProduct: db.Product{
				ID:       productId,
				Name:     "testproduct",
				Quantity: 100,
				Price:    100.0,
			},
			mockProductErr:        nil,
			mockDeductUserRows:    1,
			mockDeductUserErr:     nil,
			mockDeductProductRows: 1,
			mockDeductProductErr:  nil,
			mockAffiliateList: []db.Affiliate{
				{
					ID:              affiliateId,
					Name:            "affiliate",
					MasterAffiliate: masterAffiliateId,
				},
				{
					ID:              masterAffiliateId,
					Name:            "masteraffiliate",
					MasterAffiliate: pgtype.UUID{Valid: false},
				},
			},
			mockAffiliateErr:        nil,
			mockCreateCommissionErr: nil,
			mockAddBalanceErr:       nil,
			expectedStatus:          http.StatusCreated,
			expectedError:           "",
		},
		{
			name: "Insufficient Balance",
			reqBody: OrderRequest{
				UserID:    userId,
				ProductID: productId,
				Quantity:  1,
			},
			mockUser: db.GetUserDetailByIDRow{
				ID:       userId,
				Username: "testuser",
				Balance:  50.0,
			},
			mockUserErr: nil,
			mockProduct: db.Product{
				ID:       productId,
				Name:     "testproduct",
				Quantity: 100,
				Price:    100.0,
			},
			mockProductErr: nil,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Not enough balance",
		},
		{
			name: "Not Enough Product in Stock",
			reqBody: OrderRequest{
				UserID:    userId,
				ProductID: productId,
				Quantity:  15,
			},
			mockUser: db.GetUserDetailByIDRow{
				ID:       userId,
				Username: "testuser",
				Balance:  1000.0,
			},
			mockUserErr: nil,
			mockProduct: db.Product{
				ID:       productId,
				Name:     "testproduct",
				Quantity: 10,
				Price:    100.0,
			},
			mockProductErr: nil,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Not enough product in stock",
		},
		{
			name: "Failed to deduct user balance",
			reqBody: OrderRequest{
				UserID:    userId,
				ProductID: productId,
				Quantity:  1,
			},
			mockUser: db.GetUserDetailByIDRow{
				ID:       userId,
				Username: "testuser",
				Balance:  1000,
			},
			mockUserErr:       nil,
			mockProduct:       db.Product{ID: productId, Name: "testproduct", Quantity: 100, Price: 100},
			mockProductErr:    nil,
			mockDeductUserErr: errors.New("Failed to deduct user balance"),
			expectedStatus:    http.StatusInternalServerError,
			expectedError:     "Failed to deduct balance",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var body []byte
			var err error

			if reqStr, ok := tt.reqBody.(string); ok {
				body = []byte(reqStr)
			} else {
				body, err = json.Marshal(tt.reqBody)
				require.NoError(t, err)
			}

			var quantity int
			if req, ok := tt.reqBody.(OrderRequest); ok {
				quantity = req.Quantity
			}

			if tt.name != "Invalid JSON" {
				if tt.name == "Insufficient Balance" || tt.name == "Not Enough Product in Stock" {
					mockDB.EXPECT().GetUserDetailByID(gomock.Any(), userId).Return(tt.mockUser, tt.mockUserErr).Times(1)
					if tt.mockUserErr == nil {
						mockDB.EXPECT().GetProductByID(gomock.Any(), productId).Return(tt.mockProduct, tt.mockProductErr).Times(1)
					}
				} else if quantity > 0 {
					mockDB.EXPECT().GetUserDetailByID(gomock.Any(), userId).Return(tt.mockUser, tt.mockUserErr).Times(1)
					if tt.mockUserErr == nil {
						mockDB.EXPECT().GetProductByID(gomock.Any(), productId).Return(tt.mockProduct, tt.mockProductErr).Times(1)
					}
				}
			}

			if tt.name == "Success with no affiliate" || tt.name == "Success with calculate affiliate commission " {
				if tt.mockDeductUserErr == nil {
					mockDB.EXPECT().DeductUserBalance(gomock.Any(), gomock.Any()).Return(tt.mockDeductUserRows, tt.mockDeductUserErr).Times(1)
				}
				if tt.mockDeductUserErr == nil && tt.mockDeductProductErr == nil {
					mockDB.EXPECT().DeductProductQuantity(gomock.Any(), gomock.Any()).Return(tt.mockDeductProductRows, tt.mockDeductProductErr).Times(1)
				}
			} else if tt.name == "Failed to deduct user balance" && tt.mockUserErr == nil && tt.mockProductErr == nil {
				mockDB.EXPECT().DeductUserBalance(gomock.Any(), gomock.Any()).Return(tt.mockDeductUserRows, tt.mockDeductUserErr).Times(1)
			}

			if tt.name == "Success with calculate affiliate commission " {
				for _, affiliate := range tt.mockAffiliateList {
					mockDB.EXPECT().GetAffiliateByID(gomock.Any(), affiliate.ID).Return(affiliate, tt.mockAffiliateErr).Times(1)
				}
				if tt.mockCreateCommissionErr == nil {
					mockDB.EXPECT().CreateCommission(gomock.Any(), gomock.Any()).
						DoAndReturn(func(ctx context.Context, params db.CreateCommissionParams) (db.Commission, error) {
							for _, aff := range tt.mockAffiliateList {
								if aff.ID == params.AffiliateID {
									return db.Commission{}, tt.mockCreateCommissionErr
								}
							}
							return db.Commission{}, fmt.Errorf("unexpected affiliate ID")
						}).Times(len(tt.mockAffiliateList))

					mockDB.EXPECT().AddAffiliateBalance(gomock.Any(), gomock.Any()).
						DoAndReturn(func(ctx context.Context, params db.AddAffiliateBalanceParams) error {
							for _, aff := range tt.mockAffiliateList {
								if aff.ID == params.ID {
									return tt.mockAddBalanceErr
								}
							}
							return fmt.Errorf("unexpected affiliate ID")
						}).Times(len(tt.mockAffiliateList))
				}
			}

			router := gin.New()
			router.POST("/users/order", func(c *gin.Context) {
				NewHandler(mockDB).UserOrderProductHandler(c)
			})

			req, err := http.NewRequest(http.MethodPost, "/users/order", bytes.NewBuffer(body))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, req)

			require.Equal(t, tt.expectedStatus, recorder.Code)

			if tt.expectedError != "" {
				var response gin.H
				err := json.Unmarshal(recorder.Body.Bytes(), &response)
				require.NoError(t, err, "Failed to unmarshal response: %s", recorder.Body.String())
				require.Contains(t, response["error"], tt.expectedError)
			}
		})
	}
}
