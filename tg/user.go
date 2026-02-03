package tg

// User represents a Trustgrid user
type User struct {
	Email     string   `json:"email"`
	IDP       string   `json:"idp"`
	PolicyIDs []string `json:"policyIds"`
	Status    string   `json:"status"` // 'active' or 'inactive'
}
