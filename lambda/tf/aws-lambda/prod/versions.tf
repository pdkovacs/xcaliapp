terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
    cloudflare = {
      source  = "cloudflare/cloudflare"
      version = "~> 5.0"
    }
  }
}

provider "aws" {
  region                      = "eu-west-1"
  skip_credentials_validation = true
  skip_metadata_api_check     = true
}

provider "cloudflare" {
  api_token = var.cloudflare_api_token
}
