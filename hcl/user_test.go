package hcl

import (
	"testing"

	"github.com/trustgrid/terraform-provider-tg/tg"
)

func TestUser_ToTG(t *testing.T) {
	hclUser := User{
		Email:     "test@example.com",
		FirstName: "Test",
		LastName:  "User",
		Phone:     "+1-555-1234",
		Admin:     true,
		Active:    true,
	}

	tgUser := hclUser.ToTG()

	if tgUser.Email != hclUser.Email {
		t.Errorf("Expected Email %s, got %s", hclUser.Email, tgUser.Email)
	}
	if tgUser.FirstName != hclUser.FirstName {
		t.Errorf("Expected FirstName %s, got %s", hclUser.FirstName, tgUser.FirstName)
	}
	if tgUser.LastName != hclUser.LastName {
		t.Errorf("Expected LastName %s, got %s", hclUser.LastName, tgUser.LastName)
	}
	if tgUser.Phone != hclUser.Phone {
		t.Errorf("Expected Phone %s, got %s", hclUser.Phone, tgUser.Phone)
	}
	if tgUser.Admin != hclUser.Admin {
		t.Errorf("Expected Admin %v, got %v", hclUser.Admin, tgUser.Admin)
	}
	if tgUser.Active != hclUser.Active {
		t.Errorf("Expected Active %v, got %v", hclUser.Active, tgUser.Active)
	}
}

func TestUser_UpdateFromTG(t *testing.T) {
	tgUser := tg.User{
		UID:       "user-123",
		Email:     "test@example.com",
		FirstName: "Test",
		LastName:  "User",
		Phone:     "+1-555-1234",
		Admin:     true,
		Active:    true,
	}

	hclUser := User{}.UpdateFromTG(tgUser)
	result, ok := hclUser.(User)
	if !ok {
		t.Fatal("UpdateFromTG did not return User type")
	}

	if result.UID != tgUser.UID {
		t.Errorf("Expected UID %s, got %s", tgUser.UID, result.UID)
	}
	if result.Email != tgUser.Email {
		t.Errorf("Expected Email %s, got %s", tgUser.Email, result.Email)
	}
	if result.FirstName != tgUser.FirstName {
		t.Errorf("Expected FirstName %s, got %s", tgUser.FirstName, result.FirstName)
	}
	if result.LastName != tgUser.LastName {
		t.Errorf("Expected LastName %s, got %s", tgUser.LastName, result.LastName)
	}
	if result.Phone != tgUser.Phone {
		t.Errorf("Expected Phone %s, got %s", tgUser.Phone, result.Phone)
	}
	if result.Admin != tgUser.Admin {
		t.Errorf("Expected Admin %v, got %v", tgUser.Admin, result.Admin)
	}
	if result.Active != tgUser.Active {
		t.Errorf("Expected Active %v, got %v", tgUser.Active, result.Active)
	}
}
