data "frends_environment" "dev" {
  display_name = "Development"
}

data "frends_environment" "prod" {
  display_name = "Production"
}

resource "frends_environment_variable" "database_connection" {
  name        = "Database.ConnectionString"
  type        = "String"
  description = "Primary database connection string"

  values {
    environment_id = data.frends_environment.dev.id
    value          = "Server=dev-db;Database=myapp;Trusted_Connection=True;"
  }

  values {
    environment_id = data.frends_environment.prod.id
    value          = var.prod_db_connection
  }
}

variable "prod_db_connection" {
  type      = string
  sensitive = true
}
