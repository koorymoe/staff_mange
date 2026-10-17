package storage

import (
	"strconv"
	"testing"
	"time"
)

func TestFileTokenCarriesEmployee(t *testing.T) {
	secret := []byte("s3cret-for-tests-only")
	tok := NewFileToken(secret, "emp-1")
	id, err := VerifyFileToken(secret, tok)
	if err != nil || id != "emp-1" {
		t.Fatalf("got %q %v", id, err)
	}
	if _, err := VerifyFileToken([]byte("other"), tok); err == nil {
		t.Fatal("wrong secret accepted")
	}
	// تبديل رقم الموظف يكسر التوقيع
	forged := "emp-2" + tok[len("emp-1"):]
	if _, err := VerifyFileToken(secret, forged); err == nil {
		t.Fatal("forged employee accepted")
	}
}

func TestLegacyFileTokenStillVerifiesWithoutEmployee(t *testing.T) {
	secret := []byte("s3cret-for-tests-only")
	payload := strconv.FormatInt(time.Now().Add(time.Minute).Unix(), 10)
	id, err := VerifyFileToken(secret, payload+"."+sign(secret, payload))
	if err != nil || id != "" {
		t.Fatalf("legacy: got %q %v", id, err)
	}
	old := strconv.FormatInt(time.Now().Add(-time.Minute).Unix(), 10)
	if _, err := VerifyFileToken(secret, old+"."+sign(secret, old)); err == nil {
		t.Fatal("expired accepted")
	}
}
