package tg

// User represents a Trustgrid user
type User struct {
	UID       string   `json:"uid"`
	Email     string   `json:"email"`
	PolicyIDs []string `json:"policyIds"`
	Status    string   `json:"status"` // 'active' or 'inactive'
}
