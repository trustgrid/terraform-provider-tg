package hcl

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/trustgrid/terraform-provider-tg/tg"
)

func TestServiceUser_ToTG(t *testing.T) {
	hclServiceUser := ServiceUser{
		Name:      "test-service-user",
		Status:    "active",
		PolicyIDs: []string{"builtin-tg-access-admin", "builtin-tg-node-admin"},
	}

	tgServiceUser := hclServiceUser.ToTG()

	assert.Equal(t, hclServiceUser.Name, tgServiceUser.Name, "Name does not match")
	assert.Equal(t, hclServiceUser.Status, tgServiceUser.Status, "Status does not match")
	assert.Equal(t, hclServiceUser.PolicyIDs, tgServiceUser.PolicyIDs, "PolicyIDs do not match")
}

func TestServiceUser_UpdateFromTG(t *testing.T) {
	tgServiceUser := tg.ServiceUser{
		Name:      "test-service-user",
		OrgID:     "org-123",
		Status:    "active",
		PolicyIDs: []string{"builtin-tg-access-admin", "builtin-tg-node-admin"},
	}

	existingHCL := ServiceUser{
		Name:   "old-name",
		Status: "inactive",
	}

	hclServiceUser := existingHCL.UpdateFromTG(tgServiceUser)
	result, ok := hclServiceUser.(ServiceUser)
	require.True(t, ok, "UpdateFromTG did not return ServiceUser type")

	assert.Equal(t, tgServiceUser.Name, result.Name, "Name does not match")
	assert.Equal(t, tgServiceUser.Status, result.Status, "Status does not match")
	assert.Equal(t, tgServiceUser.PolicyIDs, result.PolicyIDs, "PolicyIDs do not match")
}
