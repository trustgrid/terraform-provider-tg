package tg

// User represents a Trustgrid user
type User struct {
	UID       string `json:"uid"`
	Email     string `json:"email"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Phone     string `json:"phone"`
	Admin     bool   `json:"admin"`
	Active    bool   `json:"active"`
}
