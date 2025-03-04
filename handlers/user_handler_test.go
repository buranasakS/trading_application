package handlers

import (
	"bytes"
	"database/sql"

	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	mockdb "github.com/buranasakS/trading_application/db/mocks"
	db "github.com/buranasakS/trading_application/db/sqlc"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)


func TestLoginUser(t *testing.T) {
	gin.SetMode(gin.TestMode)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockQueries := mockdb.NewMockQuerier(ctrl)

	router := gin.Default()
	router.POST("/login", LoginUser)

	password := "password123"
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	require.NoError(t, err)

	testUser := db.User{
		// ID:       "123456",
		Username: "testuser",
		Password: string(hashedPassword),
	}

	mockQueries.EXPECT().
		GetUserByUsernameForLogin(gomock.Any(), testUser.Username).
		Return(db.GetUserByUsernameForLoginRow{
			Username: testUser.Username,
			Password: testUser.Password,
		}, nil).
		AnyTimes()

	mockQueries.EXPECT().
		GetUserByUsernameForLogin(gomock.Any(), "wronguser").
		Return(db.GetUserByUsernameForLoginRow{}, sql.ErrNoRows).
		AnyTimes()

	tests := []struct {
		name         string
		body         gin.H
		expectedCode int
	}{
		{
			name: "Successful login",
			body: gin.H{
				"username": testUser.Username,
				"password": password,
			},
			expectedCode: http.StatusOK,
		},
		{
			name: "Invalid username",
			body: gin.H{
				"username": "wronguser",
				"password": password,
			},
			expectedCode: http.StatusUnauthorized,
		},
		{
			name: "Invalid password",
			body: gin.H{
				"username": testUser.Username,
				"password": "wrongpassword",
			},
			expectedCode: http.StatusUnauthorized,
		},
	}

	// os.Setenv("SECRET_KEY", "trading_application")

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			bodyBytes, err := json.Marshal(tc.body)
			require.NoError(t, err)

			req, err := http.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(bodyBytes))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, req)

			require.Equal(t, tc.expectedCode, recorder.Code)

			if tc.expectedCode == http.StatusOK {
				var resp map[string]string
				err = json.Unmarshal(recorder.Body.Bytes(), &resp)
				require.NoError(t, err)

				tokenStr, exists := resp["token"]
				require.True(t, exists)
				require.NotEmpty(t, tokenStr)

				token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
					return "trading_application", nil
				})
				require.NoError(t, err)
				require.True(t, token.Valid)
			}
		})
	}
}
