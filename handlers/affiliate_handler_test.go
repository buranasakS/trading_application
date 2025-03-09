package handlers

import (
	"bytes"
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

func TestCreateAffiliateHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mockdb.NewMockQuerier(ctrl)

	affiliateId := helpers.PgtypeUUID(t,"123e4567-e89b-12d3-a456-426614174000")

	tests := []struct {
		name           string
		reqBody        interface{}
		mockReturnData db.Affiliate
		mockReturnErr  error
		expectedStatus int
		expectedError  string	
	}{
		{
			name: "Success Affiliate Created",
			reqBody: RequestAffiliate{
				Name:            "Affiliate 1",
				MasterAffiliate: pgtype.UUID{Valid: false},
			},
			mockReturnData: db.Affiliate{
				ID:              affiliateId,
				Name:            "Affiliate 1",
				MasterAffiliate: pgtype.UUID{Valid: false},
				Balance:         0,
			},
			mockReturnErr:  nil,
			expectedStatus: http.StatusCreated,
			expectedError:  "",
		},
		{
			name:           "Error Invalid request body",
			reqBody:        "invalid json", 
			mockReturnData: db.Affiliate{},
			mockReturnErr:  nil,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid character 'i' looking for beginning of value",
		},
		{
			name: "Error Database error",
			reqBody: RequestAffiliate{
				Name:            "Affiliate 1",
				MasterAffiliate: pgtype.UUID{Valid: false},
			},
			mockReturnData: db.Affiliate{},
			mockReturnErr:  errors.New("DB error"),
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "Failed to create affiliate",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var body []byte
			if str, ok := tt.reqBody.(string); ok {
				body = []byte(str)
			} else {
				var err error
				body, err = json.Marshal(tt.reqBody)				
				require.NoError(t, err)
			}

			if tt.expectedStatus != http.StatusBadRequest {
				req := tt.reqBody.(RequestAffiliate)
				mockDB.EXPECT().CreateAffiliate(
					gomock.Any(),
					db.CreateAffiliateParams{
						Name:            req.Name,
						MasterAffiliate: req.MasterAffiliate,
					},
				).Return(tt.mockReturnData, tt.mockReturnErr).Times(1)
			}

			gin.SetMode(gin.ReleaseMode)
			router := gin.New()
			router.POST("/affiliates", func(c *gin.Context) {
				NewHandler(mockDB).CreateAffiliateHandler(c)
			})

			req, err := http.NewRequest(http.MethodPost, "/affiliates", bytes.NewBuffer(body))
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
				var response db.Affiliate
				err := json.Unmarshal(recorder.Body.Bytes(), &response)
				require.NoError(t, err)
				require.Equal(t, tt.mockReturnData, response)
			}
		})
	}
}

func TestListAffiliatesHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mockdb.NewMockQuerier(ctrl)

	affiliateId1 := helpers.PgtypeUUID(t,"123e4567-e89b-12d3-a456-426614174000")
	affiliate2Id := helpers.PgtypeUUID(t,"123e4567-e89b-12d3-a456-426614174001")
	

	tests := []struct {
		name           string
		mockReturnData []db.Affiliate
		mockReturnErr  error
		expectedStatus int
		expectedLength int
		expectedError  string
	}{
		{
			name: "Success Affiliates found",
			mockReturnData: []db.Affiliate{
				{
					ID:              affiliateId1,
					Name:            "Affiliate 1",
					MasterAffiliate: pgtype.UUID{},
					Balance:         0,
				},
				{
					ID:              affiliate2Id,
					Name:            "Affiliate 2",
					MasterAffiliate: pgtype.UUID{},
					Balance:         10,
				},
			},
			mockReturnErr:  nil,
			expectedStatus: http.StatusOK,
			expectedLength: 2,
			expectedError:  "",
		},
		{
			name:           "Success No affiliates found",
			mockReturnData: []db.Affiliate{},
			mockReturnErr:  nil,
			expectedStatus: http.StatusOK,
			expectedLength: 0,
			expectedError:  "",
		},
		{
			name:           "Error Failed to fetch affiliates",
			mockReturnData: nil,
			mockReturnErr:  errors.New("DB error"),
			expectedStatus: http.StatusInternalServerError,
			expectedLength: 0,
			expectedError:  "Failed to fetch affiliates",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB.EXPECT().ListAffiliates(gomock.Any()).Return(tt.mockReturnData, tt.mockReturnErr).Times(1)

			gin.SetMode(gin.TestMode)
			router := gin.New()
			router.GET("/affiliates/list", func(c *gin.Context) {
				NewHandler(mockDB).ListAffiliatesHandler(c)
			})

			req, err := http.NewRequest(http.MethodGet, "/affiliates/list", nil)
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
				var response []db.Affiliate
				err := json.Unmarshal(recorder.Body.Bytes(), &response)
				require.NoError(t, err)
				require.Equal(t, tt.expectedLength, len(response))
			}
		})
	}
}

func TestGetAffiliateDetailHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mockdb.NewMockQuerier(ctrl)

	affiliateId := helpers.PgtypeUUID(t,"123e4567-e89b-12d3-a456-426614174000")
	
	affiliate := db.Affiliate{
		ID:              affiliateId,
		Name:            "Affiliate 1",
		MasterAffiliate: pgtype.UUID{},
		Balance:         0,
	}

	tests := []struct {
		name           string
		paramID        string
		mockReturnData db.Affiliate
		mockReturnErr  error
		expectedStatus int
		expectedBody   interface{}
	}{
		{
			name:           "Success Affiliate found",
			paramID:        "123e4567-e89b-12d3-a456-426614174000",
			mockReturnData: affiliate,
			mockReturnErr:  nil,
			expectedStatus: http.StatusOK,
			expectedBody:   affiliate,
		},
		{
			name:           "Error Affiliate not found",
			paramID:        "123e4567-e89b-12d3-a456-426614174001",
			mockReturnData: db.Affiliate{},
			mockReturnErr:  errors.New("affiliate not found"),
			expectedStatus: http.StatusNotFound,
			expectedBody:   map[string]string{"error": "Affiliate not found"},
		},
		{
			name:           "Error Invalid affiliate ID",
			paramID:        "invalid-uuid", 
			mockReturnData: db.Affiliate{}, 
			mockReturnErr:  nil,           
			expectedStatus: http.StatusBadRequest,
			expectedBody:   map[string]string{"error": "Invalid affiliate ID"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.expectedStatus != http.StatusBadRequest {
				mockAffiliateID := pgtype.UUID{}
				err := mockAffiliateID.Scan(tt.paramID)
				require.NoError(t, err)
				mockDB.EXPECT().GetAffiliateByID(gomock.Any(), mockAffiliateID).Return(tt.mockReturnData, tt.mockReturnErr).Times(1)
			}

			gin.SetMode(gin.ReleaseMode)
			router := gin.New()
			router.GET("/affiliates/:id", func(c *gin.Context) {
				NewHandler(mockDB).GetAffiliateDetailHandler(c)
			})

			req, err := http.NewRequest(http.MethodGet, "/affiliates/"+tt.paramID, nil)
			require.NoError(t, err)
			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, req)

			require.Equal(t, tt.expectedStatus, recorder.Code)

			if tt.expectedStatus == http.StatusOK {
				var response db.Affiliate
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
