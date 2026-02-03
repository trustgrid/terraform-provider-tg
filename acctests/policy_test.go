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

					resource.TestCheckResourceAttr("tg_policy.test", "conditions.0.all.0.eq.0.key", "tg:node:tags:env"),
					resource.TestCheckResourceAttr("tg_policy.test", "conditions.0.all.0.eq.0.values.0", "prod"),
					resource.TestCheckResourceAttr("tg_policy.test", "conditions.0.all.0.eq.1.key", "tg:node:tags:env2"),
					resource.TestCheckResourceAttr("tg_policy.test", "conditions.0.all.0.eq.1.values.0", "prod"),
					resource.TestCheckResourceAttr("tg_policy.test", "conditions.0.all.0.eq.1.values.1", "dev"),
					resource.TestCheckResourceAttr("tg_policy.test", "conditions.0.all.0.ne.0.key", "tg:node:tags:env3"),
					resource.TestCheckResourceAttr("tg_policy.test", "conditions.0.all.0.ne.0.values.0", "prod"),
					resource.TestCheckResourceAttr("tg_policy.test", "conditions.0.all.0.ne.0.values.1", "dev"),

					resource.TestCheckResourceAttr("tg_policy.test", "conditions.0.any.0.eq.0.key", "tg:node:tags:env"),
					resource.TestCheckResourceAttr("tg_policy.test", "conditions.0.any.0.eq.0.values.0", "any1"),
					resource.TestCheckResourceAttr("tg_policy.test", "conditions.0.any.0.eq.1.key", "tg:node:tags:env2"),
					resource.TestCheckResourceAttr("tg_policy.test", "conditions.0.any.0.eq.1.values.0", "any2"),
					resource.TestCheckResourceAttr("tg_policy.test", "conditions.0.any.0.eq.1.values.1", "any3"),
					resource.TestCheckResourceAttr("tg_policy.test", "conditions.0.any.0.ne.0.key", "tg:node:tags:env3"),
					resource.TestCheckResourceAttr("tg_policy.test", "conditions.0.any.0.ne.0.values.0", "anyne1"),
					resource.TestCheckResourceAttr("tg_policy.test", "conditions.0.any.0.ne.0.values.1", "anyne2"),

					resource.TestCheckResourceAttr("tg_policy.test", "conditions.0.none.0.eq.0.key", "tg:node:tags:env"),
					resource.TestCheckResourceAttr("tg_policy.test", "conditions.0.none.0.eq.0.values.0", "none1"),
					resource.TestCheckResourceAttr("tg_policy.test", "conditions.0.none.0.eq.1.key", "tg:node:tags:env2"),
					resource.TestCheckResourceAttr("tg_policy.test", "conditions.0.none.0.eq.1.values.0", "none2"),
					resource.TestCheckResourceAttr("tg_policy.test", "conditions.0.none.0.eq.1.values.1", "none3"),
					resource.TestCheckResourceAttr("tg_policy.test", "conditions.0.none.0.ne.0.key", "tg:node:tags:env3"),
					resource.TestCheckResourceAttr("tg_policy.test", "conditions.0.none.0.ne.0.values.0", "nonene1"),
					resource.TestCheckResourceAttr("tg_policy.test", "conditions.0.none.0.ne.0.values.1", "nonene2"),
					testAcc_CheckPolicyAPISide(provider, "tf-test-policy"),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					compareValuesSame.AddStateValue("tg_policy.test", tfjsonpath.New("id")),
				},
			},
		},
	})
}

func TestAccPolicyDataSource_HappyPath(t *testing.T) {
	compareValuesSame := statecheck.CompareValue(compare.ValuesSame())

	provider := provider.New("test")()

	resource.Test(t, resource.TestCase{
		Providers: map[string]*schema.Provider{
			"tg": provider,
		},
		Steps: []resource.TestStep{
			{
				Config: policyDataSourceConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.tg_policy.test", "name", "tf-test-policy-ds"),
					resource.TestCheckResourceAttr("data.tg_policy.test", "description", "test data source policy"),
					resource.TestCheckResourceAttr("data.tg_policy.test", "resources.0", "*"),
					resource.TestCheckResourceAttr("data.tg_policy.test", "statement.0.effect", "allow"),
					resource.TestCheckResourceAttr("data.tg_policy.test", "statement.0.actions.0", "nodes::read"),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					compareValuesSame.AddStateValue("data.tg_policy.test", tfjsonpath.New("id")),
				},
			},
		},
	})
}

func TestAccPoliciesDataSource_HappyPath(t *testing.T) {
	provider := provider.New("test")()

	resource.Test(t, resource.TestCase{
		Providers: map[string]*schema.Provider{
			"tg": provider,
		},
		Steps: []resource.TestStep{
			{
				Config: policiesDataSourceConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.tg_policies.all", "names.#"),
					resource.TestCheckResourceAttrSet("data.tg_policies.all", "policies.#"),
					resource.TestCheckResourceAttrSet("data.tg_policies.filtered", "names.#"),
					resource.TestCheckResourceAttrSet("data.tg_policies.filtered", "policies.#"),
				),
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
	  eq {
	    key = "tg:node:tags:env"
	    values = ["prod"]
	  }
	  eq {
	    key = "tg:node:tags:env2"
	    values = ["prod", "dev"]
	  }
	  ne {
	    key = "tg:node:tags:env3"
	    values = ["prod", "dev"]
	  }
	}

    any { 
	  eq {
	    key = "tg:node:tags:env"
	    values = ["any1"]
	  }
	  eq {
	    key = "tg:node:tags:env2"
	    values = ["any2", "any3"]
	  }
	  ne {
	    key = "tg:node:tags:env3"
	    values = ["anyne1", "anyne2"]
	  }
	}

    none { 
	  eq {
	    key = "tg:node:tags:env"
	    values = ["none1"]
	  }
	  eq {
	    key = "tg:node:tags:env2"
	    values = ["none2", "none3"]
	  }
	  ne {
	    key = "tg:node:tags:env3"
	    values = ["nonene1", "nonene2"]
	  }
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

func policyDataSourceConfig() string {
	return `
resource "tg_policy" "test" {
  name = "tf-test-policy-ds"
  description = "test data source policy"
  resources = ["*"]

  statement {
    actions = ["nodes::read"]
	effect = "allow"
  }
}

data "tg_policy" "test" {
  name = tg_policy.test.name
  depends_on = [tg_policy.test]
}
`
}

func policiesDataSourceConfig() string {
	return `
resource "tg_policy" "test1" {
  name = "tf-test-policies-ds-1"
  description = "test policies data source 1"
  resources = ["*"]

  statement {
    actions = ["nodes::read"]
	effect = "allow"
  }
}

resource "tg_policy" "test2" {
  name = "tf-test-policies-ds-2"
  description = "test policies data source 2"
  resources = ["*"]

  statement {
    actions = ["nodes::read"]
	effect = "deny"
  }
}

data "tg_policies" "all" {
  depends_on = [tg_policy.test1, tg_policy.test2]
}

data "tg_policies" "filtered" {
  name_filter = "tf-test-policies-ds"
  depends_on = [tg_policy.test1, tg_policy.test2]
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

		case len(pol.Conditions.All.EQ) != 2:
			return fmt.Errorf("expected policy to have 2 ALL conditions but got %d", len(pol.Conditions.All.EQ))
		case len(pol.Conditions.All.EQ["tg:node:tags:env"]) != 1:
			return fmt.Errorf("expected policy to have 1 condition value but got %d", len(pol.Conditions.All.EQ["tg:node:tags:env"]))
		case pol.Conditions.All.EQ["tg:node:tags:env"][0] != "prod":
			return fmt.Errorf("expected policy to have condition value prod but got %s", pol.Conditions.All.EQ["tg:node:tags:env"][0])
		case len(pol.Conditions.All.EQ["tg:node:tags:env2"]) != 2:
			return fmt.Errorf("expected policy to have 2 condition values but got %d", len(pol.Conditions.All.EQ["tg:node:tags:env2"]))
		case pol.Conditions.All.EQ["tg:node:tags:env2"][0] != "prod":
			return fmt.Errorf("expected policy to have condition value prod but got %s", pol.Conditions.All.EQ["tg:node:tags:env2"][0])
		case pol.Conditions.All.EQ["tg:node:tags:env2"][1] != "dev":
			return fmt.Errorf("expected policy to have condition value dev but got %s", pol.Conditions.All.EQ["tg:node:tags:env2"][1])

		case len(pol.Conditions.Any.EQ) != 2:
			return fmt.Errorf("expected policy to have 2 any conditions but got %d", len(pol.Conditions.Any.EQ))
		case len(pol.Conditions.Any.EQ["tg:node:tags:env"]) != 1:
			return fmt.Errorf("expected policy to have 1 any condition values but got %d", len(pol.Conditions.Any.EQ["tg:node:tags:env"]))
		case len(pol.Conditions.Any.EQ["tg:node:tags:env2"]) != 2:
			return fmt.Errorf("expected policy to have 2 any condition values but got %d", len(pol.Conditions.Any.EQ["tg:node:tags:env2"]))
		case pol.Conditions.Any.EQ["tg:node:tags:env"][0] != "any1":
			return fmt.Errorf("expected policy to have any condition value any1 but got %s", pol.Conditions.Any.EQ["tg:node:tags:env"][0])
		case pol.Conditions.Any.EQ["tg:node:tags:env2"][0] != "any2":
			return fmt.Errorf("expected policy to have any condition value any2 but got %s", pol.Conditions.Any.EQ["tg:node:tags:env2"][0])
		case pol.Conditions.Any.EQ["tg:node:tags:env2"][1] != "any3":
			return fmt.Errorf("expected policy to have any condition value any3 but got %s", pol.Conditions.Any.EQ["tg:node:tags:env2"][1])
		case len(pol.Conditions.Any.NE["tg:node:tags:env3"]) != 2:
			return fmt.Errorf("expected policy to have 2 any/ne condition values but got %d", len(pol.Conditions.Any.NE["tg:node:tags:env3"]))
		case pol.Conditions.Any.NE["tg:node:tags:env3"][0] != "anyne1":
			return fmt.Errorf("expected first any/ne tag to be anyne1, but got %s", pol.Conditions.Any.NE["tg:node:tags:env3"][0])
		case pol.Conditions.Any.NE["tg:node:tags:env3"][1] != "anyne2":
			return fmt.Errorf("expected first any/ne tag to be anyne2, but got %s", pol.Conditions.Any.NE["tg:node:tags:env3"][1])

		case len(pol.Conditions.None.EQ) != 2:
			return fmt.Errorf("expected policy to have 2 none conditions but got %d", len(pol.Conditions.None.EQ))
		case len(pol.Conditions.None.EQ["tg:node:tags:env"]) != 1:
			return fmt.Errorf("expected policy to have 1 none condition value but got %d", len(pol.Conditions.None.EQ["tg:node:tags:env"]))
		case pol.Conditions.None.EQ["tg:node:tags:env"][0] != "none1":
			return fmt.Errorf("expected policy to have none condition value none1 but got %s", pol.Conditions.None.EQ["tg:node:tags:env"][0])
		case len(pol.Conditions.Any.EQ["tg:node:tags:env2"]) != 2:
			return fmt.Errorf("expected policy to have 2 none condition values but got %d", len(pol.Conditions.Any.EQ["tg:node:tags:env2"]))
		}

		return nil
	}
}
