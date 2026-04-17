data "frends_environment" "staging" {
  display_name = "Staging"
}

resource "frends_api_key" "staging_key" {
  name           = "terraform-managed-key"
  environment_id = data.frends_environment.staging.id
}

output "api_key_value" {
  value     = frends_api_key.staging_key.value
  sensitive = true
}
