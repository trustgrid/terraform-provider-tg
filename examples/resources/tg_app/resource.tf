resource "tg_app" "mywebapp" {
  name                  = "my-web-app"
  description           = "my app2"
  gateway_node          = "gateway-node-uuid"
  ip                    = "2.2.2.2"
  idp                   = "idp-id"
  port                  = 443
  protocol              = "http"
  type                  = "web"
  hostname              = "myhostname"
  session_duration      = 60
  tls_verification_mode = "default"
  trust_mode            = "discovery"
}

resource "tg_app" "myremoteapp" {
  name         = "my-remote-app"
  description  = "my app2"
  gateway_node = "gateway-node-uuid"
  ip           = "2.2.2.2"
  idp          = "idp-id"
  port         = 3389
  protocol     = "rdp"
  type         = "web"
}

resource "tg_app" "mywgapp" {
  name             = "my-wg-app"
  description      = "my app2"
  gateway_node     = "gateway-node-uuid"
  ip               = "2.2.2.2"
  idp              = "idp-id"
  port             = 3389
  session_duration = 60
  protocol         = "wireguard"
  type             = "wireguard"
}
