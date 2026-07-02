package utils

import (
	"testing"
	"time"
)

func TestGenerateToken(t *testing.T) {
	token, err := GenerateToken("test", "1234567890", time.Now().Add(time.Hour*24))
	if err != nil {
		t.Errorf("generate token failed, err: %v", err)
	}
	mc, err := VerifyToken(token, "1234567890")
	if err != nil {
		t.Errorf("verify token failed, err: %v", err)
	}
	t.Logf("mc: %v", mc)
	if mc.Username != "test" {
		t.Errorf("username is not test, got %s", mc.Username)
	}
	if mc.ExpiresAt.Before(time.Now()) {
		t.Errorf("expires at is before now, got %s", mc.ExpiresAt)
	}
	if mc.IssuedAt.After(time.Now()) {
		t.Errorf("issued at is after now, got %s", mc.IssuedAt)
	}
}
