
resource "tg_app_access_rule" "rule1" {
  app    = "app-id"
  action = "allow"
  name   = "bigrule"

  exception {
    emails   = ["exception@whatever.com"]
    everyone = true
  }

  include {
    ip_ranges = ["0.0.0.0/0"]
    countries = ["US"]
  }

  require {
    emails_ending_in = ["trustgrid.io"]
    idp_groups       = ["your-idp-id"]
    access_groups    = ["your-access-group-id"]
  }
}

resource "tg_app_access_rule" "deny-everyone" {
  depends_on = [resource.tg_app_access_rule.rule1]
  app        = "app-id"
  action     = "block"
  name       = "block"

  include {
    everyone = true
  }
}
