package handlers

import (
	"database/sql"
	"errors"
	"net/http"
	"testing"

	mockdb "github.com/buranasakS/trading_application/db/mocks"
	db "github.com/buranasakS/trading_application/db/sqlc"
	"github.com/buranasakS/trading_application/helpers"
	"github.com/golang/mock/gomock"
	"github.com/jackc/pgx/v5/pgtype"
)


func TestUserOrderProductHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// mockDB := mockdb.NewMockQuerier(ctrl)

	userId := helpers.PgtypeUUID(t, "123e4567-e89b-12d3-a456-426614174000")
	productId := helpers.PgtypeUUID(t, "123e4567-e89b-12d3-a456-426614174001")
	affiliateId := helpers.PgtypeUUID(t, "123e4567-e89b-12d3-a456-426614174002")
	masterAffiliateId := helpers.PgtypeUUID(t, "123e4567-e89b-12d3-a456-426614174003")

	tests := []struct {
		name           string
		reqBody        OrderRequest
		mockUser       db.GetUserDetailByIDRow
		mockUserErr    error
		mockProduct    db.Product
		mockProductErr error
		mockDeductUserRows int64
		mockDeductUserErr  error
		mockDeductProductRows int64
		mockDeductProductErr  error
		mockAffiliateList []db.Affiliate
		mockAffiliateErr  error
		mockCreateCommissionErr error
		mockAddBalanceErr error
		expectedStatus int
		expectedError  string
	}{
		{
			name: "Success",
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
			mockProductErr: nil,
			mockDeductUserRows: 1,
			mockDeductUserErr:  nil,
			mockDeductProductRows: 1,
			mockDeductProductErr:  nil,
			expectedStatus: http.StatusCreated,
			expectedError:  "",
		},
		{
			name: "Success with commission",
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
			mockProductErr: nil,
			mockDeductUserRows: 1,
			mockDeductUserErr:  nil,
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
			mockAffiliateErr:    nil,
			mockCreateCommissionErr: nil,
			mockAddBalanceErr: nil,
			expectedStatus: http.StatusCreated,
			expectedError:  "",
		},
		{
			name: "Invalid Quantity",
			reqBody: OrderRequest{
				UserID:    userId,
				ProductID: productId,
				Quantity:  0,
			},
			mockUser:       db.GetUserDetailByIDRow{},
			mockUserErr:    nil,
			mockProduct:    db.Product{},
			mockProductErr: nil,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Quantity must be more than 0",
		},
		{
			name: "User Not Found",
			reqBody: OrderRequest{
				UserID:    userId,
				ProductID: productId,
				Quantity:  1,
			},
			mockUser:       db.GetUserDetailByIDRow{},
			mockUserErr:    sql.ErrNoRows,
			mockProduct:    db.Product{},
			mockProductErr: nil,
			expectedStatus: http.StatusNotFound,
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
			mockProduct:    db.Product{},
			mockProductErr: sql.ErrNoRows,
			expectedStatus: http.StatusNotFound,
			expectedError:  "Product not found",
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
			name: "Insufficient Product in Stock",
			reqBody: OrderRequest{
				UserID:    userId,
				ProductID: productId,
				Quantity:  101,
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
				Quantity: 100,
				Price:    100.0,
			},
			mockProductErr: nil,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Not enough product in stock",
		},
		{
			name: "Failed to deduct User Balance",
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
			mockUserErr: nil,
			mockProduct: db.Product{
				ID:       productId,
				Name:     "testproduct",
				Quantity: 100,
				Price:    100.0,
			},
			mockProductErr: nil,
			mockDeductUserRows: 0,
			mockDeductUserErr:  errors.New("Failed to deduct user balance"),
			mockDeductProductRows: 1,
			mockDeductProductErr:  nil,
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "Failed to deduct balance",
		},
		{
			name: "Failed to deduct Product Quantity",
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
			mockUserErr: nil,
			mockProduct: db.Product{
				ID:       productId,
				Name:     "testproduct",
				Quantity: 100,
				Price:    100.0,
			},
			mockProductErr: nil,
			mockDeductUserRows: 1,
			mockDeductUserErr:  nil,
			mockDeductProductRows: 0,
			mockDeductProductErr:  errors.New("Failed to deduct product quantity"),
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "Failed to deduct product quantity",
		},
		{
			name: "Failed to Create Commission",
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
			mockProductErr: nil,
			mockDeductUserRows: 1,
			mockDeductUserErr:  nil,
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
			mockAffiliateErr:    nil,
			mockCreateCommissionErr: errors.New("Failed to create commission"),
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "Failed to create commission",
		},
		{
			name: "Failed to Add Balance",
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
			mockProductErr: nil,
			mockDeductUserRows: 1,
			mockDeductUserErr:  nil,
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
			mockAffiliateErr:    nil,
			mockCreateCommissionErr: nil,
			mockAddBalanceErr: errors.New("Failed to add affiliate balance"),
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "Failed to add affiliate balance",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := mockdb.NewMockQuerier(ctrl)
			mockDB.EXPECT().GetUserDetailByID(gomock.Any(), gomock.Any()).Return(tt.mockUser, tt.mockUserErr)
		})
	}

}
