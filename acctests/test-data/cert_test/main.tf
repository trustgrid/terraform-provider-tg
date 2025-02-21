
resource "tg_cert" "test" {
  fqdn = "portal.dev.trustgrid.io"
  body = file("domain.crt")
  chain = file("domain.chain")
  private_key = file("domain.key")
}