package hcl

import (
	"testing"

	"github.com/trustgrid/terraform-provider-tg/tg"
)

func TestServiceUser_ToTG(t *testing.T) {
	hclServiceUser := ServiceUser{
		Name:      "test-service-user",
		Status:    "active",
		PolicyIDs: []string{"builtin-tg-access-admin", "builtin-tg-node-admin"},
		ClientID:  "client-123",
		Secret:    "secret-456",
	}

	tgServiceUser := hclServiceUser.ToTG()

	if tgServiceUser.Name != hclServiceUser.Name {
		t.Errorf("Expected Name %s, got %s", hclServiceUser.Name, tgServiceUser.Name)
	}
	if tgServiceUser.Status != hclServiceUser.Status {
		t.Errorf("Expected Status %s, got %s", hclServiceUser.Status, tgServiceUser.Status)
	}
	if len(tgServiceUser.PolicyIDs) != len(hclServiceUser.PolicyIDs) {
		t.Errorf("Expected %d PolicyIDs, got %d", len(hclServiceUser.PolicyIDs), len(tgServiceUser.PolicyIDs))
	}
	for i, policyID := range tgServiceUser.PolicyIDs {
		if policyID != hclServiceUser.PolicyIDs[i] {
			t.Errorf("Expected PolicyID[%d] %s, got %s", i, hclServiceUser.PolicyIDs[i], policyID)
		}
	}
	// ClientID and Secret should NOT be included in ToTG
	if tgServiceUser.OrgID != "" {
		t.Errorf("Expected empty OrgID, got %s", tgServiceUser.OrgID)
	}
}

func TestServiceUser_UpdateFromTG(t *testing.T) {
	tgServiceUser := tg.ServiceUser{
		Name:      "test-service-user",
		OrgID:     "org-123",
		Status:    "active",
		PolicyIDs: []string{"builtin-tg-access-admin", "builtin-tg-node-admin"},
	}

	// Create an HCL ServiceUser with existing ClientID and Secret
	existingHCL := ServiceUser{
		ClientID: "existing-client-id",
		Secret:   "existing-secret",
	}

	hclServiceUser := existingHCL.UpdateFromTG(tgServiceUser)
	result, ok := hclServiceUser.(ServiceUser)
	if !ok {
		t.Fatal("UpdateFromTG did not return ServiceUser type")
	}

	if result.Name != tgServiceUser.Name {
		t.Errorf("Expected Name %s, got %s", tgServiceUser.Name, result.Name)
	}
	if result.Status != tgServiceUser.Status {
		t.Errorf("Expected Status %s, got %s", tgServiceUser.Status, result.Status)
	}
	if len(result.PolicyIDs) != len(tgServiceUser.PolicyIDs) {
		t.Errorf("Expected %d PolicyIDs, got %d", len(tgServiceUser.PolicyIDs), len(result.PolicyIDs))
	}
	for i, policyID := range result.PolicyIDs {
		if policyID != tgServiceUser.PolicyIDs[i] {
			t.Errorf("Expected PolicyID[%d] %s, got %s", i, tgServiceUser.PolicyIDs[i], policyID)
		}
	}
	// ClientID and Secret should be preserved from the HCL object
	if result.ClientID != existingHCL.ClientID {
		t.Errorf("Expected ClientID %s, got %s", existingHCL.ClientID, result.ClientID)
	}
	if result.Secret != existingHCL.Secret {
		t.Errorf("Expected Secret %s, got %s", existingHCL.Secret, result.Secret)
	}
}
