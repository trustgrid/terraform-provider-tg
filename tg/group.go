package tg

type Group struct {
	UID         string `json:"uid"`
	Name        string `json:"name"`
	IDP         string `json:"idp"`
	Description string `json:"description"`
	ReferenceID string `json:"referenceId"`
}

type GroupMember struct {
	User string `json:"user"`
}
