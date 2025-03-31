package hcl

import (
	"maps"
	"slices"

	"github.com/trustgrid/terraform-provider-tg/tg"
)

type Expression struct {
	Key    string   `tf:"key"`
	Values []string `tf:"values"`
}

type Condition struct {
	EQ []Expression `tf:"eq"`
	NE []Expression `tf:"ne"`
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

	if o.Conditions.All.Len() > 0 || o.Conditions.Any.Len() > 0 || o.Conditions.None.Len() > 0 {
		all := Condition{}
		anyc := Condition{}
		none := Condition{}

		for _, k := range slices.Sorted(maps.Keys(o.Conditions.All.EQ)) {
			all.EQ = append(all.EQ, Expression{Key: k, Values: o.Conditions.All.EQ[k]})
		}
		for _, k := range slices.Sorted(maps.Keys(o.Conditions.All.NE)) {
			all.NE = append(all.NE, Expression{Key: k, Values: o.Conditions.All.NE[k]})
		}

		for _, k := range slices.Sorted(maps.Keys(o.Conditions.Any.EQ)) {
			anyc.EQ = append(anyc.EQ, Expression{Key: k, Values: o.Conditions.Any.EQ[k]})
		}
		for _, k := range slices.Sorted(maps.Keys(o.Conditions.Any.NE)) {
			anyc.NE = append(anyc.NE, Expression{Key: k, Values: o.Conditions.Any.NE[k]})
		}

		for _, k := range slices.Sorted(maps.Keys(o.Conditions.None.EQ)) {
			none.EQ = append(none.EQ, Expression{Key: k, Values: o.Conditions.None.EQ[k]})
		}
		for _, k := range slices.Sorted(maps.Keys(o.Conditions.None.NE)) {
			none.NE = append(none.NE, Expression{Key: k, Values: o.Conditions.None.NE[k]})
		}

		updated.Conditions = []Conditions{{
			All:  []Condition{all},
			Any:  []Condition{anyc},
			None: []Condition{none},
		}}
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
		pc := p.Conditions[0]

		out.Conditions.All.EQ = make(map[string][]string)
		out.Conditions.All.NE = make(map[string][]string)
		out.Conditions.Any.EQ = make(map[string][]string)
		out.Conditions.Any.NE = make(map[string][]string)
		out.Conditions.None.EQ = make(map[string][]string)
		out.Conditions.None.NE = make(map[string][]string)

		if len(pc.All) > 0 {
			for _, c := range pc.All[0].EQ {
				out.Conditions.All.EQ[c.Key] = c.Values
			}
			for _, c := range pc.All[0].NE {
				out.Conditions.All.NE[c.Key] = c.Values
			}
		}
		if len(pc.Any) > 0 {
			for _, c := range pc.Any[0].EQ {
				out.Conditions.Any.EQ[c.Key] = c.Values
			}
			for _, c := range pc.Any[0].NE {
				out.Conditions.Any.NE[c.Key] = c.Values
			}
		}
		if len(pc.None) > 0 {
			for _, c := range pc.None[0].EQ {
				out.Conditions.None.EQ[c.Key] = c.Values
			}
			for _, c := range pc.None[0].NE {
				out.Conditions.None.NE[c.Key] = c.Values
			}
		}
	}

	return out
}
