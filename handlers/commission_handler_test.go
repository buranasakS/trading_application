package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
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

func TestListCommissionsHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mockdb.NewMockQuerier(ctrl)

	commissionId1 := helpers.PgtypeUUID(t, "123e4567-e89b-12d3-a456-426614174000")
	commissionId2 := helpers.PgtypeUUID(t, "123e4567-e89b-12d3-a456-426614174001")
	orderId := helpers.PgtypeUUID(t, "123e4567-e89b-12d3-a456-426614174002")
	affiliateId := helpers.PgtypeUUID(t, "123e4567-e89b-12d3-a456-426614174003")

	commissions := []db.Commission{
		{ID: commissionId1, OrderID: orderId, AffiliateID: affiliateId, Amount: 100.0},
		{ID: commissionId2, OrderID: orderId, AffiliateID: affiliateId, Amount: 150.0},
	}

	tests := []struct {
		name           string
		mockReturnData []db.Commission
		mockReturnErr  error
		expectedStatus int
		expectedLength int
		expectedError  string
	}{
		{
			name:           "Success Commissions found",
			mockReturnData: commissions,
			mockReturnErr:  nil,
			expectedStatus: http.StatusOK,
			expectedLength: 2,
			expectedError:  "",
		},
		{
			name:           "Success No commissions found",
			mockReturnData: []db.Commission{},
			mockReturnErr:  nil,
			expectedStatus: http.StatusOK,
			expectedLength: 0,
			expectedError:  "",
		},
		{
			name:           "Error Failed to fetch commissions",
			mockReturnData: nil,
			mockReturnErr:  errors.New("database error"),
			expectedStatus: http.StatusInternalServerError,
			expectedLength: 0,
			expectedError:  "Failed to fetch commissions",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB.EXPECT().ListCommissions(gomock.Any()).Return(tt.mockReturnData, tt.mockReturnErr).Times(1)

			gin.SetMode(gin.TestMode)
			router := gin.New()
			router.GET("/commissions/list", func(c *gin.Context) {
				NewHandler(mockDB).ListCommissionsHandler(c)
			})

			req, err := http.NewRequest(http.MethodGet, "/commissions/list", nil)
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
				var response []db.Commission
				err := json.Unmarshal(recorder.Body.Bytes(), &response)
				require.NoError(t, err)
				require.Equal(t, tt.expectedLength, len(response))
			}
		})
	}
}

func TestGetCommissionDetailHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mockdb.NewMockQuerier(ctrl)

	commissionId := helpers.PgtypeUUID(t, "123e4567-e89b-12d3-a456-426614174000")
	orderId := helpers.PgtypeUUID(t, "123e4567-e89b-12d3-a456-426614174002")
	affiliateId := helpers.PgtypeUUID(t, "123e4567-e89b-12d3-a456-426614174003")

	commission := db.Commission{
		ID:          commissionId,
		OrderID:     orderId,
		AffiliateID: affiliateId,
		Amount:      100.0,
	}

	tests := []struct {
		name           string
		commissionID   string
		mockReturnData db.Commission
		mockReturnErr  error
		expectedStatus int
		expectedBody   interface{}
	}{
		{
			name:           "Success Commission found",
			commissionID:   "123e4567-e89b-12d3-a456-426614174000",
			mockReturnData: commission,
			mockReturnErr:  nil,
			expectedStatus: http.StatusOK,
			expectedBody:   commission,
		},
		{
			name:           "Error Commission not found",
			commissionID:   "123e4567-e89b-12d3-a456-426614174001",
			mockReturnData: db.Commission{},
			mockReturnErr:  errors.New("Commission not found"),
			expectedStatus: http.StatusNotFound,
			expectedBody:   map[string]string{"error": "Commission not found"},
		},
		{
			name:           "Error Invalid commission ID format",
			commissionID:   "invalid-uuid",
			mockReturnData: db.Commission{},
			mockReturnErr:  nil,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   map[string]string{"error": "Invalid commission ID"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.expectedStatus != http.StatusBadRequest {
				mockCommissionID := pgtype.UUID{}
				err := mockCommissionID.Scan(tt.commissionID)
				require.NoError(t, err)
				mockDB.EXPECT().GetCommissionByID(gomock.Any(), mockCommissionID).Return(tt.mockReturnData, tt.mockReturnErr).Times(1)
			}

			gin.SetMode(gin.ReleaseMode)
			router := gin.New()
			router.GET("/commissions/:id", func(c *gin.Context) {
				NewHandler(mockDB).GetCommissionDetailHandler(c)
			})

			req, err := http.NewRequest(http.MethodGet, "/commissions/"+tt.commissionID, nil)
			require.NoError(t, err)
			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, req)

			require.Equal(t, tt.expectedStatus, recorder.Code)

			if tt.expectedStatus == http.StatusOK {
				var response db.Commission
				err := json.Unmarshal(recorder.Body.Bytes(), &response)
				require.NoError(t, err)
				require.Equal(t, tt.mockReturnData, response)
			} else {
				var response map[string]string
				err := json.Unmarshal(recorder.Body.Bytes(), &response)
				require.NoError(t, err)
				require.Equal(t, tt.expectedBody, response)
			}
		})
	}
}

func TestGetCommissionDistributionHandler(t *testing.T) {
    ctrl := gomock.NewController(t)
    defer ctrl.Finish()

    mockDB := mockdb.NewMockQuerier(ctrl)

    orderId := helpers.PgtypeUUID(t, "123e4567-e89b-12d3-a456-426614174002")
    masteraffiliateId1 := helpers.PgtypeUUID(t, "123e4567-e89b-12d3-a456-426614174003")
    affiliateId2 := helpers.PgtypeUUID(t, "123e4567-e89b-12d3-a456-426614174004")

    commissions := []db.GetCommissionByOrderIDRow{
        {ID: masteraffiliateId1, Name: "MasterAffiliate1", Amount: 20.0},
        {ID: affiliateId2, Name: "Affiliate2", Amount: 5.0},
    }

    tests := []struct {
        name               string
        orderID            string
        mockTotal          float64
        mockCommissions    []db.GetCommissionByOrderIDRow
        mockTotalErr       error
        mockCommissionsErr error
        expectedStatus     int
        expectedBody       interface{}
    }{
        {
            name:            "Success",
            orderID:         "123e4567-e89b-12d3-a456-426614174002",
            mockTotal:       25,
            mockCommissions: commissions,
            mockTotalErr:    nil,
            mockCommissionsErr: nil,
            expectedStatus:  http.StatusOK,
            expectedBody: CommsisionDistributionResponse{
                OrderID:         orderId,
                TotalCommission: 25.0,
                Details: []CommissionAffiliateDetail{
                    {AffiliateID: masteraffiliateId1, AffiliateName: "MasterAffiliate1", Commission: 20.0},
                    {AffiliateID: affiliateId2, AffiliateName: "Affiliate2", Commission: 5.0},
                },
            },
        },
        {
            name:            "Invalid Order ID",
            orderID:         "invalid-uuid",
            mockTotal:       0.0,
            mockCommissions: nil,
            mockTotalErr:    nil,
            mockCommissionsErr: nil,
            expectedStatus:  http.StatusBadRequest,
            expectedBody:    map[string]string{"error": "Invalid order ID"},
        },
        {
            name:            "Get Total Commission Error",
            orderID:         "123e4567-e89b-12d3-a456-426614174002",
            mockTotal:       0.0,
            mockCommissions: nil,
            mockTotalErr:    sql.ErrNoRows,
            mockCommissionsErr: nil,
            expectedStatus:  http.StatusBadRequest,
            expectedBody:    map[string]string{"error": "Order not found or no commission available"},
        },
        {
            name:            "Get Commissions Error",
            orderID:         "123e4567-e89b-12d3-a456-426614174002",
            mockTotal:       100,
            mockCommissions: nil,
            mockTotalErr:    nil,
            mockCommissionsErr: sql.ErrNoRows,
            expectedStatus:  http.StatusBadRequest,
            expectedBody:    map[string]string{"error": "Failed to fetch commissions"},
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            if tt.orderID != "invalid-uuid" { 
                mockOrderID := pgtype.UUID{}
                err := mockOrderID.Scan(tt.orderID)
                require.NoError(t, err)

                mockDB.EXPECT().GetTotalCommission(gomock.Any(), mockOrderID).Return(tt.mockTotal, tt.mockTotalErr).Times(1)

                if tt.mockTotalErr == nil {
                    mockDB.EXPECT().GetCommissionByOrderID(gomock.Any(), mockOrderID).Return(tt.mockCommissions, tt.mockCommissionsErr).Times(1)
                }
            }

            gin.SetMode(gin.TestMode)
            router := gin.New()
            router.GET("/commissions/distribution/:order_id", func(c *gin.Context) {
                NewHandler(mockDB).GetCommissionDistributionHandler(c)
            })

            req, err := http.NewRequest(http.MethodGet, "/commissions/distribution/"+tt.orderID, nil)
            require.NoError(t, err)

            recorder := httptest.NewRecorder()
            router.ServeHTTP(recorder, req)

            require.Equal(t, tt.expectedStatus, recorder.Code)

            if tt.expectedStatus == http.StatusOK {
                var response CommsisionDistributionResponse
                err := json.Unmarshal(recorder.Body.Bytes(), &response)
                require.NoError(t, err)
                require.Equal(t, tt.expectedBody, response)
            } else {
                var response map[string]string
                err := json.Unmarshal(recorder.Body.Bytes(), &response)
                require.NoError(t, err)
                require.Equal(t, tt.expectedBody, response)
            }
        })
    }
}
