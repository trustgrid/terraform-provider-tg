package tg

type Statement struct {
	Actions []string `json:"actions"`
	Effect  string   `json:"effect"`
}

type Policy struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Resources   []string `json:"resources"`
	Conditions  struct {
		All  map[string][]string `json:"all"`
		Any  map[string][]string `json:"any"`
		None map[string][]string `json:"none"`
	} `json:"conditions"`
	Statements []Statement `json:"statements"`
}
