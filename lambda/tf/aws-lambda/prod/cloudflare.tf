data "cloudflare_zone" "this" {
  filter = {
    name = var.cloudflare_zone
  }
}

locals {
  function_url_host = trimsuffix(replace(aws_lambda_function_url.xcali_prod.function_url, "https://", ""), "/")
}

resource "cloudflare_dns_record" "app" {
  zone_id = data.cloudflare_zone.this.zone_id
  name    = var.app_hostname
  type    = "CNAME"
  ttl     = 1
  content = local.function_url_host
  proxied = true
}

resource "cloudflare_workers_script" "proxy" {
  account_id  = var.cloudflare_account_id
  script_name = "xcaliapp-proxy"
  content     = file("${path.module}/worker/proxy.js")
  main_module = "proxy.js"

  bindings = [
    {
      name = "FUNCTION_URL_HOST"
      type = "plain_text"
      text = local.function_url_host
    }
  ]
}

resource "cloudflare_workers_route" "proxy" {
  zone_id = data.cloudflare_zone.this.zone_id
  pattern = "${var.app_hostname}/*"
  script  = cloudflare_workers_script.proxy.script_name
}

resource "cloudflare_zero_trust_access_policy" "allow_owners" {
  account_id = var.cloudflare_account_id
  name       = "xcaliapp-allow-owners"
  decision   = "allow"

  include = [
    for email in var.allowed_emails : {
      email = {
        email = email
      }
    }
  ]
}

resource "cloudflare_zero_trust_access_application" "app" {
  account_id = var.cloudflare_account_id
  name       = "xcaliapp"
  type       = "self_hosted"
  domain     = var.app_hostname

  destinations = [
    {
      type = "public"
      uri  = var.app_hostname
    }
  ]

  policies = [
    {
      id         = cloudflare_zero_trust_access_policy.allow_owners.id
      precedence = 1
    }
  ]
}
