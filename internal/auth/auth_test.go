package auth

import (
	"testing"
	"github.com/google/uuid"
	"time"
)

func TestHashAndCheckPassword(t *testing.T) {
	type args struct {
		password string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Hash and check password",
			args: args{
				password: "mysecretpassword",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hashedPassword, err := HashPassword(tt.args.password)
			if (err != nil) != tt.wantErr {
				t.Errorf("HashPassword() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(hashedPassword) == 0 {
				t.Errorf("HashPassword() got empty hash")
				return
			}
			match, err := CheckPasswordHash(tt.args.password, hashedPassword)
			if (err != nil) != tt.wantErr {
				t.Errorf("CheckPasswordHash() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !match {
				t.Errorf("CheckPasswordHash() got = %v, want true", match)
			}
		})
	}
}

func TestMakeAndParseJWT(t *testing.T) {
	type args struct {
		userID     string
		tokenSecret string
		expiresIn  int64
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Make and parse JWT",
			args: args{
				userID:     "550e8400-e29b-41d4-a716-446655440000",
				tokenSecret: "mysecretkey",
				expiresIn:  3600,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := MakeJWT(uuid.MustParse(tt.args.userID), tt.args.tokenSecret, time.Duration(tt.args.expiresIn)*time.Second)
			if (err != nil) != tt.wantErr {
				t.Errorf("MakeJWT() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(token) == 0 {
				t.Errorf("MakeJWT() got empty token")
				return
			}
			parsedUserID, err := ParseJWT(token, tt.args.tokenSecret)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseJWT() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if parsedUserID.String() != tt.args.userID {
				t.Errorf("ParseJWT() got = %v, want %v", parsedUserID.String(), tt.args.userID)
			}
		})
	}
}

	
