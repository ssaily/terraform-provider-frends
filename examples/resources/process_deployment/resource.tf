# Look up the environment and agent group where we want to deploy.
data "frends_environment" "production" {
  display_name = "Production"
}

data "frends_agent_group" "production_agents" {
  display_name   = "Production Agents"
  environment_id = data.frends_environment.production.id
}

# Look up the process to deploy (built in Frends Control Panel).
data "frends_process" "order_processor" {
  name = "Order Processor"
}

# Deploy the process to the production agent group.
resource "frends_process_deployment" "order_processor" {
  agent_group_id         = data.frends_agent_group.production_agents.id
  activate_triggers      = true
  deployment_description = "Deployed by Terraform — GitOps workflow"

  processes {
    process_guid = data.frends_process.order_processor.guid
    version      = data.frends_process.order_processor.version

    process_variables {
      name  = "ApiBaseUrl"
      value = "https://api.example.com"
    }

    process_variables {
      name      = "ApiSecret"
      value     = var.api_secret
      is_secret = true
    }
  }
}

variable "api_secret" {
  type      = string
  sensitive = true
}

output "deployment_id" {
  value = frends_process_deployment.order_processor.id
}
