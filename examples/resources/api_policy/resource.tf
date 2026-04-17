resource "frends_api_policy" "order_api_policy" {
  name               = "Order API Policy"
  description        = "Access policy for the Order Processing API"
  allow_public_access = false
  api_key_name       = "X-API-Key"
  api_key_location   = "Header"
  tags               = ["production", "order-api"]

  target_endpoints {
    url    = "https://my-company.frends.com/api/order"
    method = "POST"
  }

  target_endpoints {
    url    = "https://my-company.frends.com/api/order"
    method = "GET"
  }

  identities {
    name = "InternalServices"

    rules {
      claim_type  = "aud"
      claim_value = "frends-api"
      match_type  = "Exact"
    }
  }
}
