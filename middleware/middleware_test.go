package middleware

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJwtMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	testSecretKey := "test_secret_key"
	os.Setenv("SECRET_KEY", testSecretKey)
	defer os.Unsetenv("SECRET_KEY")

	validToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": "testuser",
		"exp": time.Now().Add(time.Hour).Unix(),
		"iat": time.Now().Unix(),
	})
	validTokenString, err := validToken.SignedString([]byte(testSecretKey))
	require.NoError(t, err)

	expiredToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": "testuser",
		"exp": time.Now().Add(-time.Hour).Unix(),
		"iat": time.Now().Add(-2 * time.Hour).Unix(),
	})

	expiredTokenString, err := expiredToken.SignedString([]byte(testSecretKey))
	require.NoError(t, err)

	tests := []struct {
		name            string
		tokenHeader     string
		expectedStatus  int
		expectedError   string
		expectedUserID  string
		testNextHandler bool
	}{
		{
			name:            "Valid Token",
			tokenHeader:     "Bearer " + validTokenString,
			expectedStatus:  http.StatusOK,
			expectedError:   "",
			expectedUserID:  "testuser",
			testNextHandler: true,
		},
		{
			name:            "Missing Authorization Header",
			tokenHeader:     "",
			expectedStatus:  http.StatusUnauthorized,
			expectedError:   "Authorization header missing",
			expectedUserID:  "",
			testNextHandler: false,
		},
		{
			name:            "Missing Token String",
			tokenHeader:     "Bearer ",
			expectedStatus:  http.StatusUnauthorized,
			expectedError:   "Authorization token missing",
			expectedUserID:  "",
			testNextHandler: false,
		},
		{
			name:            "Invalid Token",
			tokenHeader:     "Bearer invalid_token",
			expectedStatus:  http.StatusUnauthorized,
			expectedError:   "Invalid or expired token",
			expectedUserID:  "",
			testNextHandler: false,
		},
		{
			name:            "Expired Token",
			tokenHeader:     "Bearer " + expiredTokenString,
			expectedStatus:  http.StatusUnauthorized,
			expectedError:   "Invalid or expired token",
			expectedUserID:  "",
			testNextHandler: false,
		},
		{
			name:            "Wrong Signing Method",
			tokenHeader:     "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c",
			expectedStatus:  http.StatusUnauthorized,
			expectedError:   "Invalid or expired token",
			expectedUserID:  "",
			testNextHandler: false,
		},
		{
			name: "Missing sub in token",
			tokenHeader: "Bearer " + func() string {
				missingSubToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
					"exp": time.Now().Add(time.Hour).Unix(),
					"iat": time.Now().Unix(),
				})
				missingSubTokenString, err := missingSubToken.SignedString([]byte(testSecretKey))
				require.NoError(t, err)
				return missingSubTokenString
			}(),
			expectedStatus:  http.StatusUnauthorized,
			expectedError:   "User ID not found in token",
			expectedUserID:  "",
			testNextHandler: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			router := gin.New()
			var userIDFromContext string
			nextHandlerCalled := false

			if tt.testNextHandler {
				router.Use(JwtMiddleware())
				router.GET("/test", func(c *gin.Context) {
					nextHandlerCalled = true
					userIDFromContext = c.Request.Context().Value("user_id").(string)
					c.JSON(http.StatusOK, gin.H{"message": "success"})
				})
			} else {
				router.GET("/test", JwtMiddleware(), func(c *gin.Context) {
					nextHandlerCalled = true
					userIDFromContext = c.Request.Context().Value("user_id").(string)
					c.JSON(http.StatusOK, gin.H{"message": "success"})
				})
			}

			req := httptest.NewRequest("GET", "/test", nil)
			req.Header.Set("Authorization", tt.tokenHeader)
			recorder := httptest.NewRecorder()

			router.ServeHTTP(recorder, req)

			assert.Equal(t, tt.expectedStatus, recorder.Code)

			if tt.expectedError != "" {
				assert.Contains(t, recorder.Body.String(), tt.expectedError)
			}

			if tt.testNextHandler {
				require.True(t, nextHandlerCalled, "Next handler should have been called")
				require.Equal(t, tt.expectedUserID, userIDFromContext)
			} else {
				require.False(t, nextHandlerCalled, "Next handler should not have been called")
			}
		})
	}
}
