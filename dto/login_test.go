package dto

import "testing"

func TestLoginRequestValidate(t *testing.T) {
	tests := []struct {
		name      string
		request   LoginRequest
		wantCount int
	}{
		{
			name:      "valid request",
			request:   LoginRequest{Username: "tiago", Password: "strong-password"},
			wantCount: 0,
		},
		{
			name:      "blank username and password",
			request:   LoginRequest{},
			wantCount: 2,
		},
		{
			name:      "short username",
			request:   LoginRequest{Username: "abc", Password: "strong-password"},
			wantCount: 1,
		},
		{
			name:      "short password",
			request:   LoginRequest{Username: "tiago", Password: "123456"},
			wantCount: 1,
		},
		{
			name:      "trims spaces before validating",
			request:   LoginRequest{Username: "   ", Password: "      "},
			wantCount: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := tt.request.Validate()
			if len(errs) != tt.wantCount {
				t.Fatalf("Expected %d validation errors, got %d", tt.wantCount, len(errs))
			}
		})
	}
}
