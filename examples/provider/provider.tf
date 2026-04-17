terraform {
  required_providers {
    frends = {
      source  = "frends/frends"
      version = "~> 0.1"
    }
  }
}

# Configure the Frends provider.
# Credentials can also be set via environment variables:
#   FRENDS_HOST_URL  - Base URL of your Frends Platform instance
#   FRENDS_TOKEN     - Bearer token for API authentication
provider "frends" {
  host_url = "https://my-company.frends.com"
  token    = var.frends_token
}

variable "frends_token" {
  description = "Frends Platform API bearer token"
  type        = string
  sensitive   = true
}
