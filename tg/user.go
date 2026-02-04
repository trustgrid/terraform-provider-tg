package tg

// User represents a Trustgrid user
type User struct {
	UID       string   `json:"uid,omitempty"`
	Email     string   `json:"email"`
	IDP       string   `json:"idp,omitempty"`
	PolicyIDs []string `json:"policyIds"`
	Status    string   `json:"status"` // 'active' or 'inactive'
}
