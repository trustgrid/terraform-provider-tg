package acctests

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-testing/compare"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/trustgrid/terraform-provider-tg/provider"
	"github.com/trustgrid/terraform-provider-tg/tg"
)

func TestAccPolicy_HappyPath(t *testing.T) {
	compareValuesSame := statecheck.CompareValue(compare.ValuesSame())

	provider := provider.New("test")()

	resource.Test(t, resource.TestCase{
		Providers: map[string]*schema.Provider{
			"tg": provider,
		},
		Steps: []resource.TestStep{
			{
				Config: policyConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("tg_policy.test", "name", "tf-test-policy"),
					resource.TestCheckResourceAttr("tg_policy.test", "description", "my policy"),
					resource.TestCheckResourceAttr("tg_policy.test", "resources.0", "*"),
					resource.TestCheckResourceAttr("tg_policy.test", "resources.1", "tgrn:tg::nodes:node/8e684dec-f83a-49a1-b306-74d816559453"),
					resource.TestCheckResourceAttr("tg_policy.test", "statement.0.effect", "allow"),
					resource.TestCheckResourceAttr("tg_policy.test", "statement.0.actions.0", "nodes::read"),
					resource.TestCheckResourceAttr("tg_policy.test", "statement.0.actions.1", "certificates::read"),
					resource.TestCheckResourceAttr("tg_policy.test", "statement.1.effect", "deny"),
					resource.TestCheckResourceAttr("tg_policy.test", "statement.1.actions.0", "nodes::cluster"),
					resource.TestCheckResourceAttr("tg_policy.test", "statement.1.actions.1", "certificates::modify"),
					resource.TestCheckResourceAttr("tg_policy.test", "conditions.0.all.0.key", "tg:node:tags:env"),
					resource.TestCheckResourceAttr("tg_policy.test", "conditions.0.all.0.values.0", "prod"),
					resource.TestCheckResourceAttr("tg_policy.test", "conditions.0.all.1.key", "tg:node:tags:env2"),
					resource.TestCheckResourceAttr("tg_policy.test", "conditions.0.all.1.values.0", "prod"),
					resource.TestCheckResourceAttr("tg_policy.test", "conditions.0.all.1.values.1", "dev"),
					resource.TestCheckResourceAttr("tg_policy.test", "conditions.0.any.0.key", "tg:node:tags:env"),
					resource.TestCheckResourceAttr("tg_policy.test", "conditions.0.any.0.values.0", "prod"),
					resource.TestCheckResourceAttr("tg_policy.test", "conditions.0.any.0.values.1", "dev"),
					resource.TestCheckResourceAttr("tg_policy.test", "conditions.0.none.0.key", "tg:node:tags:env"),
					resource.TestCheckResourceAttr("tg_policy.test", "conditions.0.none.0.values.0", "dev"),
					testAcc_CheckPolicyAPISide(provider, "tf-test-policy"),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					compareValuesSame.AddStateValue("tg_policy.test", tfjsonpath.New("id")),
				},
			},
		},
	})
}

func policyConfig() string {
	return `
resource "tg_policy" "test" {
  name = "tf-test-policy"
  description = "my policy"
  resources = ["*", "tgrn:tg::nodes:node/8e684dec-f83a-49a1-b306-74d816559453"]
  conditions {
    all { 
	  key = "tg:node:tags:env"
	  values = ["prod"]
	}

	all { 
	  key = "tg:node:tags:env2"
	  values = ["prod", "dev"]
	}

	any { 
	  key = "tg:node:tags:env"
	  values = ["prod", "dev"]
	}

	none { 
	  key = "tg:node:tags:env"
	  values = ["dev"]
	}
  }

  statement {
    actions = ["nodes::read", "certificates::read"]
	effect = "allow"
  }

  statement {
	actions = ["nodes::cluster", "certificates::modify"]
	effect = "deny"
  }
}
`
}

func testAcc_CheckPolicyAPISide(provider *schema.Provider, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := provider.Meta().(*tg.Client)

		var pol tg.Policy

		if err := client.Get(context.Background(), "/v2/policy/"+name, &pol); err != nil {
			return err
		}

		switch {
		case len(pol.Resources) != 2:
			return fmt.Errorf("expected policy to have 2 resources but got %d", len(pol.Resources))
		case pol.Resources[0] != "*":
			return fmt.Errorf("expected policy to have resource * but got %s", pol.Resources[0])
		case pol.Resources[1] != "tgrn:tg::nodes:node/8e684dec-f83a-49a1-b306-74d816559453":
			return fmt.Errorf("expected policy to have resource tgrn:tg::nodes:node/8e684dec-f83a-49a1-b306-74d816559453 but got %s", pol.Resources[1])
		case len(pol.Statements) != 2:
			return fmt.Errorf("expected policy to have 2 statements but got %d", len(pol.Statements))
		case pol.Statements[0].Effect != "allow":
			return fmt.Errorf("expected policy to have effect allow but got %s", pol.Statements[0].Effect)
		case len(pol.Statements[0].Actions) != 2:
			return fmt.Errorf("expected policy to have 2 actions but got %d", len(pol.Statements[0].Actions))
		case pol.Statements[0].Actions[0] != "nodes::read":
			return fmt.Errorf("expected policy to have action nodes::read but got %s", pol.Statements[0].Actions[0])
		case pol.Statements[0].Actions[1] != "certificates::read":
			return fmt.Errorf("expected policy to have action certificates::read but got %s", pol.Statements[0].Actions[1])
		case pol.Statements[1].Effect != "deny":
			return fmt.Errorf("expected policy to have effect deny but got %s", pol.Statements[1].Effect)
		case len(pol.Statements[1].Actions) != 2:
			return fmt.Errorf("expected policy to have 2 actions but got %d", len(pol.Statements[1].Actions))
		case pol.Statements[1].Actions[0] != "nodes::cluster":
			return fmt.Errorf("expected policy to have action nodes::cluster but got %s", pol.Statements[1].Actions[0])
		case pol.Statements[1].Actions[1] != "certificates::modify":
			return fmt.Errorf("expected policy to have action certificates::modify but got %s", pol.Statements[1].Actions[1])
		case len(pol.Conditions.All) != 2:
			return fmt.Errorf("expected policy to have 2 ALL conditions but got %d", len(pol.Conditions.All))
		case len(pol.Conditions.All["tg:node:tags:env"]) != 1:
			return fmt.Errorf("expected policy to have 1 condition value but got %d", len(pol.Conditions.All["tg:node:tags:env"]))
		case pol.Conditions.All["tg:node:tags:env"][0] != "prod":
			return fmt.Errorf("expected policy to have condition value prod but got %s", pol.Conditions.All["tg:node:tags:env"][0])
		case len(pol.Conditions.All["tg:node:tags:env2"]) != 2:
			return fmt.Errorf("expected policy to have 2 condition values but got %d", len(pol.Conditions.All["tg:node:tags:env2"]))
		case pol.Conditions.All["tg:node:tags:env2"][0] != "prod":
			return fmt.Errorf("expected policy to have condition value prod but got %s", pol.Conditions.All["tg:node:tags:env2"][0])
		case pol.Conditions.All["tg:node:tags:env2"][1] != "dev":
			return fmt.Errorf("expected policy to have condition value dev but got %s", pol.Conditions.All["tg:node:tags:env2"][1])
		case len(pol.Conditions.Any) != 1:
			return fmt.Errorf("expected policy to have 1 any condition but got %d", len(pol.Conditions.Any))
		case len(pol.Conditions.Any["tg:node:tags:env"]) != 2:
			return fmt.Errorf("expected policy to have 2 any condition values but got %d", len(pol.Conditions.Any["tg:node:tags:env"]))
		case pol.Conditions.Any["tg:node:tags:env"][0] != "prod":
			return fmt.Errorf("expected policy to have any condition value prod but got %s", pol.Conditions.Any["tg:node:tags:env"][0])
		case pol.Conditions.Any["tg:node:tags:env"][1] != "dev":
			return fmt.Errorf("expected policy to have any condition value dev but got %s", pol.Conditions.Any["tg:node:tags:env"][1])
		case len(pol.Conditions.None) != 1:
			return fmt.Errorf("expected policy to have 1 none condition but got %d", len(pol.Conditions.None))
		case len(pol.Conditions.None["tg:node:tags:env"]) != 1:
			return fmt.Errorf("expected policy to have 1 none condition value but got %d", len(pol.Conditions.None["tg:node:tags:env"]))
		case pol.Conditions.None["tg:node:tags:env"][0] != "dev":
			return fmt.Errorf("expected policy to have none condition value dev but got %s", pol.Conditions.None["tg:node:tags:env"][0])
		}

		return nil
	}
}
