package tg

type APIToken struct {
	ClientID string `json:"clientId"`
	Secret   string `json:"secret"` //nolint:gosec // Trustgrid API schema uses this field name
}
