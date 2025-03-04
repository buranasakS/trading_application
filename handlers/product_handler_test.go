package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	mockdb "github.com/buranasakS/trading_application/db/mocks"
	db "github.com/buranasakS/trading_application/db/sqlc"
	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/require"
)

func LoadEnv() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	
}
func TestListProducts(t *testing.T) {
	gin.SetMode(gin.TestMode)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	
	mockQueries := mockdb.NewMockQuerier(ctrl)

	products := []db.Product{
		{
			ID:       pgtype.UUID{Bytes: [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}},
			Name:     "Product 1",
			Quantity: 10,
			Price:    100.0,
		},
		{
			ID:       pgtype.UUID{Bytes: [16]byte{2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17}},
			Name:     "Product 2",
			Quantity: 5,
			Price:    200.0,
		},
	}

	mockQueries.EXPECT().
		ListProducts(gomock.Any()).  
		Return(products, nil).
		AnyTimes()

	router := gin.Default()
		router.GET("/products/list", ListProductsHandler)

	req, err := http.NewRequest(http.MethodGet, "/products/list", nil)
	require.NoError(t, err)

	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	require.Equal(t, http.StatusOK, recorder.Code)

	var resp []db.Product
	err = json.Unmarshal(recorder.Body.Bytes(), &resp)
	require.NoError(t, err)

	require.Len(t, resp, len(products))
	require.Equal(t, products[0].ID, resp[0].ID)
	require.Equal(t, products[1].ID, resp[1].ID)
}
