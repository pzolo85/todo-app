// Package claim holds a claim of user type
package claim

import (
	"testing"
	"time"
)

func TestUserClaim_Valid(t *testing.T) {
	type fields struct {
		Email     string
		CreatedAt time.Time
		ExpiresAt time.Time
		IsAdmin   bool
		SourceIP  string
		UserAgent string
		ClaimID   string
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := UserClaim{
				Email:     tt.fields.Email,
				CreatedAt: tt.fields.CreatedAt,
				ExpiresAt: tt.fields.ExpiresAt,
				IsAdmin:   tt.fields.IsAdmin,
				SourceIP:  tt.fields.SourceIP,
				UserAgent: tt.fields.UserAgent,
				ClaimID:   tt.fields.ClaimID,
			}
			if err := u.Valid(); (err != nil) != tt.wantErr {
				t.Errorf("UserClaim.Valid() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_validateUser(t *testing.T) {
	type args struct {
		u *UserClaim
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := validateUser(tt.args.u); (err != nil) != tt.wantErr {
				t.Errorf("validateUser() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
