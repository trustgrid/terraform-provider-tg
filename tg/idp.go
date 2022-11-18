package tg

type IDP struct {
	Type        string `json:"class"`
	Description string `json:"description"`
	Name        string `json:"name"`
	UID         string `json:"uid,omitempty"`
}
