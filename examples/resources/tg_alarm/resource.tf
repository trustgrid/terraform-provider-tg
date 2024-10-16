resource "tg_alarm" "my_alarm" {
  name        = "my_alarm"
  description = "this alarm is for ..."
  enabled     = true
  channels    = ["channelID1", "channelID2"]
  nodes       = ["gateway1", "edge1"]
  types       = ["All Peers Disconnected", "All Gateways Disconnected"]
  operator    = "any"
  tag {
    name  = "prod_status"
    value = "Production"
  }
  tag {
    name  = "alertable"
    value = "true"
  }
  tag_operator = "any"
  threshold    = "WARNING"
  freetext     = "something terrible"
  expr         = "ctx.node.name.contains(\"prod\")"
}