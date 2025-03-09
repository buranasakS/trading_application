package helpers

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestPgtypeUUID(t *testing.T) {
	testCases := []struct {
		name          string
		uuidStr       string
		expectedValid bool
		expectedError bool
	}{
		{
			name:          "Valid UUID",
			uuidStr:       "123e4567-e89b-12d3-a456-426614174000",
			expectedValid: true,
			expectedError: false,
		},
		{
			name:          "Empty UUID",
			uuidStr:       "",
			expectedValid: false,
			expectedError: false, 
		},
		{
			name:          "Invalid UUID",
			uuidStr:       "invalid-uuid",
			expectedValid: false,
			expectedError: false, 
		},
		{
			name:          "Too Short",
			uuidStr:       "123e4567-e89b-12d3-a456-426614174",
			expectedValid: false,
			expectedError: false, 
		},
		{
			name:          "Too Long",
			uuidStr:       "123e4567-e89b-12d3-a456-4266141740000",
			expectedValid: false,
			expectedError: false, 
		},
		{
			name:          "Lowercase UUID",
			uuidStr:       "f47ac10b-58cc-4372-a567-0e02b2c3d479",
			expectedValid: true,
			expectedError: false,
		},
		{
			name:          "Uppercase UUID",
			uuidStr:       "F47AC10B-58CC-4372-A567-0E02B2C3D479",
			expectedValid: true,
			expectedError: false,
		},
		{
			name:          "All Zeroes",
			uuidStr:       "00000000-0000-0000-0000-000000000000",
			expectedValid: true,
			expectedError: false,
		},
		{
			name:          "Valid UUID from uuid package",
			uuidStr:       uuid.New().String(),
			expectedValid: true,
			expectedError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			id := PgtypeUUID(t, tc.uuidStr)
			require.Equal(t, tc.expectedValid, id.Valid, "Expected UUID to be valid")

			if tc.expectedValid {
				_, err := uuid.Parse(tc.uuidStr)
				require.NoError(t, err, "String must can parse with uuid package")
			}
		})
	}
}
