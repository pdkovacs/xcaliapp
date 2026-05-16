variable "s3_bucket" {
  type    = string
  default = "test-xcalidrawings"
}

variable "cf_access_team_domain" {
  type        = string
  description = "Cloudflare Access team subdomain (the part before .cloudflareaccess.com)"
}

variable "allowed_emails" {
  type        = list(string)
  description = "Emails permitted by the Cloudflare Access policy"
  default     = ["peter.dunay.kovacs@gmail.com"]
}

variable "cloudflare_account_id" {
  type        = string
  description = "Cloudflare account ID that owns the zone and the tunnel"
  sensitive   = true
}

variable "cloudflare_api_token" {
  type        = string
  description = "Cloudflare API token with Account/Access:Apps and Policies/Edit, Account/Workers Scripts/Edit, Zone/DNS/Edit, Zone/Workers Routes/Edit, Zone/Zone/Read"
  sensitive   = true
}

variable "cloudflare_zone" {
  type        = string
  description = "Cloudflare DNS zone (apex domain) hosting the app"
  default     = "bitkit.click"
}

variable "app_hostname" {
  type        = string
  description = "Public hostname for the app (must be inside cloudflare_zone)"
  default     = "xcaliapp.bitkit.click"
}
