output "app_url" {
  description = "Public app URL (Cloudflare Access in front of Lambda Function URL)"
  value       = "https://${var.app_hostname}"
}

output "function_url" {
  description = "Lambda Function URL (AuthType=NONE; CF Access JWT is verified inside the Lambda handler)"
  value       = aws_lambda_function_url.xcali_prod.function_url
}

output "access_aud" {
  description = "Cloudflare Access AUD (matches CF_ACCESS_AUD env var on the Lambda)"
  value       = cloudflare_zero_trust_access_application.app.aud
}
