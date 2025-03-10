package helpers

import (
	"testing"
	"unicode"
)

func TestGenerateRandomString(t *testing.T) {
	tests := []struct {
		name          string
		length        int
		expectedError bool
	}{
		{
			name:          "Valid length",
			length:        10,
			expectedError: false,
		},
		{
			name:          "Zero length",
			length:        0,
			expectedError: true,
		},
		{
			name:          "Negative length",
			length:        -5,
			expectedError: true, 
		},
		{
			name:          "Large length",
			length:        1000,
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			randomString, err := GenerateRandomString(tt.length)

			if tt.expectedError {
				if err == nil {
					t.Errorf("Expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if len(randomString) != tt.length {
				t.Errorf("Expected string length %d, got %d", tt.length, len(randomString))
			}

			for _, r := range randomString {
				if !unicode.IsLetter(r) && !unicode.IsDigit(r) {
					t.Errorf("String contains invalid character: %c", r)
				}
			}
		})
	}
}
