resource "frends_private_application" "order_service" {
  name                       = "Order Service App"
  description                = "OAuth application for the order service"
  default_token_lifetime_days = 30

  custom_token_claims = {
    "service" = "order-service"
    "env"     = "production"
  }

  tags = ["production", "order-service"]
}

output "app_id" {
  value = frends_private_application.order_service.id
}
