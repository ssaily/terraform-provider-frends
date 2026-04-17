resource "frends_api_specification" "order_api" {
  name        = "Order API"
  description = "OpenAPI specification for the Order Processing service"
}

output "api_spec_id" {
  value = frends_api_specification.order_api.id
}
