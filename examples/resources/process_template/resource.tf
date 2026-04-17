resource "frends_process_template" "http_integration" {
  name        = "HTTP Integration Template"
  description = "Standard template for HTTP-based integrations"
}

output "template_id" {
  value = frends_process_template.http_integration.id
}
