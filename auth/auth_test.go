package auth

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestJwtService_GetClaims(t *testing.T) {
	service := NewJwtService("testCodePhrase")
	userId := 1

	expiredToken, err := service.GenerateToken(userId)
	assert.Nil(t, err)
	service.TokenTTL = 10 * time.Minute
	token, err := service.GenerateToken(userId)
	assert.Nil(t, err)

	type args struct {
		token string
	}
	tests := []struct {
		name    string
		service *JwtService
		args    args
		want    *TokenClaims
		wantErr bool
	}{
		{
			name:    "ValidToken",
			service: service,
			args:    args{token: "Bearer " + token},
			want: &TokenClaims{
				UserID: userId,
			},
			wantErr: false,
		},
		{
			name:    "ValidButExpired",
			service: service,
			args:    args{token: "Bearer " + expiredToken},
			want: &TokenClaims{
				UserID: userId,
			},
			wantErr: true,
		},
		{
			name:    "InvalidTokenFormat",
			service: service,
			args:    args{token: "InvalidFormatToken"},
			want:    nil,
			wantErr: true,
		},
		{
			name:    "InvalidToken",
			service: service,
			args:    args{token: "Bearer InvalidToken"},
			want:    nil,
			wantErr: true,
		},
		{
			name:    "EmptyToken",
			service: service,
			args:    args{token: ""},
			want:    nil,
			wantErr: true,
		},
		{
			name:    "NoBearer",
			service: service,
			args:    args{token: token},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := tt.service.GetClaims(tt.args.token)
			if (err != nil) != tt.wantErr {
				t.Errorf("JwtService.GetClaims() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != nil && got.UserID != tt.want.UserID {
				t.Errorf("JwtService.GetClaims().UserID = %v, want %v", got.UserID, tt.want.UserID)
			}
		})
	}
}
