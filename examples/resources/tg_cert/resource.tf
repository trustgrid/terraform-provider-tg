
resource "tg_cert" "cert" {
  fqdn        = "myapp.trustgrid.io"
  body        = file("path-to-domain.crt")
  chain       = file("path-to-domain.chain")
  private_key = file("path-to-domain.key")
}