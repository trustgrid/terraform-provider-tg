
resource "tg_policy" "sample_policy" {
  name        = "sample-policy"
  description = "allow readonly access to nodes for developers"
  resources   = ["*"]
  conditions {
    any {
      key    = "tg:node:tags:env"
      values = ["dev", "test"]
    }

    none {
      key    = "tg:node:tags:env"
      values = ["prod"]
    }
  }

  statement {
    actions = ["nodes::read"]
    effect  = "allow"
  }
}