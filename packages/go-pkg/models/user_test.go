package models

import "testing"

func TestUserDisplayName(t *testing.T) {
	if got := (User{Email: "a@example.com"}).DisplayName(); got != "a@example.com" {
		t.Errorf("without a name the email is shown: %q", got)
	}
	if got := (User{Email: "a@example.com", Name: "Sam"}).DisplayName(); got != "Sam" {
		t.Errorf("name = %q", got)
	}
}
