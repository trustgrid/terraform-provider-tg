resource "tg_tagging" "node1_tags" {
  node_id = "35ee5516-c6d5-409b-b1ba-6aa2d0dd92fcf"

  tags = {
    tagkey1     = "a value"
    prod_status = "Production"
  }
}
