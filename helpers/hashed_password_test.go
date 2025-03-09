package helpers

import (
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func TestHashedPassword(t *testing.T) {
	testCases := []struct {
		name          string
		password      string
		expectedError bool
	}{
		{
			name:          "Valid Password",
			password:      "testpassword123",
			expectedError: false,
		},
		{
			name:          "Empty Password",
			password:      "",
			expectedError: false,
		},
		{
			name:          "Special Characters",
			password:      "!@#$%^&*()",
			expectedError: false,
		},
		{
			name:          "Long Password",
			password:      "ThisIsAVeryLongPasswordThatShouldWorkJustFine1234567890!@#$%^&*()",
			expectedError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			hashedPassword, err := HashedPassword(tc.password)
			if tc.expectedError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotEmpty(t, hashedPassword)

				err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(tc.password))
				require.NoError(t, err)
			}
		})
	}
}
