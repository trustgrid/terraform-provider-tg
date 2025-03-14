package hcl

import (
	"maps"
	"slices"

	"github.com/trustgrid/terraform-provider-tg/tg"
)

type Condition struct {
	Key    string   `tf:"key"`
	Values []string `tf:"values"`
}

type Conditions struct {
	All  []Condition `tf:"all"`
	Any  []Condition `tf:"any"`
	None []Condition `tf:"none"`
}

// Policy holds the HCL representation of a Policy
type Policy struct {
	Name        string       `tf:"name"`
	Description string       `tf:"description"`
	Statements  []Statement  `tf:"statement"`
	Resources   []string     `tf:"resources"`
	Conditions  []Conditions `tf:"conditions"`
}

type Statement struct {
	Actions []string `tf:"actions"`
	Effect  string   `tf:"effect"`
}

// UpdateFromTG updates the HCL representation of a Policy from the TG API representation
func (p Policy) UpdateFromTG(o tg.Policy) HCL[tg.Policy] {
	updated := Policy{
		Name:        o.Name,
		Description: o.Description,
		Resources:   o.Resources,
	}
	for _, statement := range o.Statements {
		updated.Statements = append(updated.Statements, Statement{
			Actions: statement.Actions,
			Effect:  statement.Effect,
		})
	}

	if len(o.Conditions.All) > 0 || len(o.Conditions.Any) > 0 || len(o.Conditions.None) > 0 {
		uc := Conditions{}

		for _, k := range slices.Sorted(maps.Keys(o.Conditions.All)) {
			uc.All = append(uc.All, Condition{Key: k, Values: o.Conditions.All[k]})
		}

		for _, k := range slices.Sorted(maps.Keys(o.Conditions.Any)) {
			uc.Any = append(uc.Any, Condition{Key: k, Values: o.Conditions.Any[k]})
		}

		for _, k := range slices.Sorted(maps.Keys(o.Conditions.None)) {
			uc.None = append(uc.None, Condition{Key: k, Values: o.Conditions.None[k]})
		}

		updated.Conditions = []Conditions{uc}
	} else {
		updated.Conditions = nil
	}

	return updated
}

// ToTG returns the TG API representation of a Policy from the HCL representation
func (p Policy) ToTG() tg.Policy {
	out := tg.Policy{
		Name:        p.Name,
		Description: p.Description,
		Resources:   p.Resources,
	}

	for _, statement := range p.Statements {
		out.Statements = append(out.Statements, tg.Statement{
			Actions: statement.Actions,
			Effect:  statement.Effect,
		})
	}

	if len(p.Conditions) > 0 {
		out.Conditions.All = make(map[string][]string)
		out.Conditions.Any = make(map[string][]string)
		out.Conditions.None = make(map[string][]string)

		for _, c := range p.Conditions[0].All {
			out.Conditions.All[c.Key] = c.Values
		}
		for _, c := range p.Conditions[0].Any {
			out.Conditions.Any[c.Key] = c.Values
		}
		for _, c := range p.Conditions[0].None {
			out.Conditions.None[c.Key] = c.Values
		}
	}

	return out
}
