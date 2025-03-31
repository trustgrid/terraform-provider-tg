package tg

type Statement struct {
	Actions []string `json:"actions"`
	Effect  string   `json:"effect"`
}

type Conditions struct {
	EQ map[string][]string `json:"eq"`
	NE map[string][]string `json:"ne"`
}

func (c Conditions) Len() int {
	return len(c.EQ) + len(c.NE)
}

type Policy struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Resources   []string `json:"resources"`
	Conditions  struct {
		All  Conditions `json:"all"`
		Any  Conditions `json:"any"`
		None Conditions `json:"none"`
	} `json:"conditions"`
	Statements []Statement `json:"statements"`
}
