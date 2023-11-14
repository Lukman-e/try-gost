package hash

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerate(t *testing.T) {
	tests := []struct {
		name    string
		data    string
		wantErr bool
	}{
		{"TestGenerateValid", "password123", false},
		{"TestGenerateEmpty", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Generate(tt.data)
			assert.Equal(t, (err != nil), tt.wantErr, "Generate() error")
		})
	}
}

func TestVerify(t *testing.T) {
	// Generate a hashed password for testing
	hashedPassword, _ := Generate("password123")

	tests := []struct {
		name           string
		hashedPassword string
		password       string
		want           bool
		wantErr        bool
	}{
		{"TestVerifyValid", hashedPassword, "password123", true, false},
		{"TestVerifyInvalid", hashedPassword, "wrongpassword", false, false},
		{"TestVerifyEmpty", hashedPassword, "", false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Verify(tt.hashedPassword, tt.password)
			assert.Equal(t, (err != nil), tt.wantErr, "Verify() error")
			assert.Equal(t, got, tt.want, "Verify() result")
		})
	}
}
