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

func TestCreateProductHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mockdb.NewMockQuerier(ctrl)

	productId := helpers.PgtypeUUID(t, "123e4567-e89b-12d3-a456-426614174000")

	tests := []struct {
		name           string
		reqBody        interface{}
		mockReturnData db.Product
		mockReturnErr  error
		expectedStatus int
		expectedError  string
	}{
		{
			name: "Success Product Created",
			reqBody: RequestProduct{
				Name:     "New Product",
				Quantity: 10,
				Price:    100.0,
			},
			mockReturnData: db.Product{
				ID:       productId,
				Name:     "New Product",
				Quantity: 10,
				Price:    100.0,
			},
			mockReturnErr:  nil,
			expectedStatus: http.StatusCreated,
			expectedError:  "",
		},
		{
			name: "Invalid quantity (<= 0)",
			reqBody: RequestProduct{
				Name:     "Invalid Product",
				Quantity: 0,
				Price:    100.0,
			},
			mockReturnData: db.Product{},
			mockReturnErr:  nil,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Key: 'RequestProduct.Quantity' Error:Field validation for 'Quantity' failed on the 'required' tag",
		},
		{
			name: "Invalid price (<= 0)",
			reqBody: RequestProduct{
				Name:     "Invalid Product",
				Quantity: 10,
				Price:    -5.0,
			},
			mockReturnData: db.Product{},
			mockReturnErr:  nil,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Price must be more than 0",
		},
		{
			name: "Missing Name field",
			reqBody: RequestProduct{
				Quantity: 10,
				Price:    100.0,
			},
			mockReturnData: db.Product{},
			mockReturnErr:  nil,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Key: 'RequestProduct.Name' Error:Field validation for 'Name' failed on the 'required' tag",
		},
		{
			name: "Missing Quantity field",
			reqBody: RequestProduct{
				Name:  "Product Without Quantity",
				Price: 100.0,
			},
			mockReturnData: db.Product{},
			mockReturnErr:  nil,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Key: 'RequestProduct.Quantity' Error:Field validation for 'Quantity' failed on the 'required' tag",
		},
		{
			name: "Missing Price field",
			reqBody: RequestProduct{
				Name:     "Product Without Price",
				Quantity: 10,
			},
			mockReturnData: db.Product{},
			mockReturnErr:  nil,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Key: 'RequestProduct.Price' Error:Field validation for 'Price' failed on the 'required' tag",
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
				req := tt.reqBody.(RequestProduct)
				mockDB.EXPECT().CreateProduct(
					gomock.Any(),
					db.CreateProductParams{
						Name:     req.Name,
						Quantity: req.Quantity,
						Price:    req.Price,
					},
				).Return(tt.mockReturnData, tt.mockReturnErr).Times(1)
			}

			gin.SetMode(gin.ReleaseMode)
			router := gin.New()
			router.POST("/products", func(c *gin.Context) {
				NewHandler(mockDB).CreateProductHandler(c)
			})

			req, err := http.NewRequest(http.MethodPost, "/products", bytes.NewBuffer(body))
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
				var response db.Product
				err := json.Unmarshal(recorder.Body.Bytes(), &response)
				require.NoError(t, err)
				require.Equal(t, tt.mockReturnData, response)
			}
		})
	}
}
func TestListProductsHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mockdb.NewMockQuerier(ctrl)

	products := []db.Product{
		{ID: pgtype.UUID{}, Name: "Product 1", Quantity: 10, Price: 100},
		{ID: pgtype.UUID{}, Name: "Product 2", Quantity: 5, Price: 50},
	}

	mockDB.EXPECT().ListProducts(gomock.Any()).Return(products, nil).AnyTimes()

	gin.SetMode(gin.ReleaseMode)
	router := gin.New()

	router.GET("/products/list", func(c *gin.Context) {
		NewHandler(mockDB).ListProductsHandler(c)
	})

	req, err := http.NewRequest(http.MethodGet, "/products/list", nil)
	require.NoError(t, err)

	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	require.Equal(t, http.StatusOK, recorder.Code)

	var response []db.Product
	err = json.Unmarshal(recorder.Body.Bytes(), &response)
	require.NoError(t, err)

	require.Equal(t, len(response), len(products))
	require.Equal(t, response[0].ID, products[0].ID)
	require.Equal(t, response[1].ID, products[1].ID)
}

func TestProductDetailHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mockdb.NewMockQuerier(ctrl)

	productId := helpers.PgtypeUUID(t, "123e4567-e89b-12d3-a456-426614174000")

	product := db.Product{
		ID:       productId,
		Name:     "Product 1",
		Quantity: 10,
		Price:    100.0,
	}

	tests := []struct {
		name           string
		paramID        string
		mockReturnData db.Product
		mockReturnErr  error
		expectedStatus int
		expectedBody   interface{}
	}{
		{
			name:           "Success Get Product Detail",
			paramID:        productId.String(),
			mockReturnData: product,
			mockReturnErr:  nil,
			expectedStatus: http.StatusOK,
			expectedBody:   product,
		},
		{
			name:           "Product Not Found",
			paramID:        "123e4567-e89b-12d3-a456-426614174001",
			mockReturnData: db.Product{},
			mockReturnErr:  errors.New("Commission not found"),
			expectedStatus: http.StatusNotFound,
			expectedBody:   map[string]string{"error": "Product not found"},
		},
		{
			name:           "Error Invalid product ID Format",
			paramID:        "invalid-uuid",
			mockReturnData: db.Product{},
			mockReturnErr:  nil,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   map[string]string{"error": "Invalid product ID"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.expectedStatus != http.StatusBadRequest {
				mockProductId := pgtype.UUID{}
				err := mockProductId.Scan(tt.paramID)
				require.NoError(t, err)
				mockDB.EXPECT().GetProductByID(gomock.Any(), mockProductId).Return(tt.mockReturnData, tt.mockReturnErr).Times(1)
			}

			gin.SetMode(gin.TestMode)
			router := gin.New()
			router.GET("/products/:id", func(c *gin.Context) {
				NewHandler(mockDB).GetProductDetailHandler(c)
			})

			req, err := http.NewRequest(http.MethodGet, "/products/"+tt.paramID, nil)
			require.NoError(t, err)
			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, req)

			require.Equal(t, tt.expectedStatus, recorder.Code)

			if tt.expectedStatus == http.StatusOK {
				var response db.Product
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
