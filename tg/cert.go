package tg

type Cert struct {
	FQDN string `tf:"fqdn" json:"fqdn"`

	Body       string `tf:"body" json:"certificateBody"`
	Chain      string `tf:"chain" json:"certificateChain"`
	PrivateKey string `tf:"private_key" json:"privateKey"`
}
