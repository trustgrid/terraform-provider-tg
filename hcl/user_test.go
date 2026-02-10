package hcl

import (
	"testing"

	"github.com/trustgrid/terraform-provider-tg/tg"
)

func TestUser_ToTG(t *testing.T) {
	hclUser := User{
		Email: "test@example.com",
	}

	tgUser := hclUser.ToTG()

	if tgUser.Email != hclUser.Email {
		t.Errorf("Expected Email %s, got %s", hclUser.Email, tgUser.Email)
	}
}

func TestUser_UpdateFromTG(t *testing.T) {
	tgUser := tg.User{
		Email: "test@example.com",
	}

	hclUser := User{}.UpdateFromTG(tgUser)
	result, ok := hclUser.(User)
	if !ok {
		t.Fatal("UpdateFromTG did not return User type")
	}

	if result.Email != tgUser.Email {
		t.Errorf("Expected Email %s, got %s", tgUser.Email, result.Email)
	}
}
