# Import a process package and deploy it to production.
# The filesha256() call ensures Terraform detects when the .frends file changes
# and automatically re-imports it as a new version on the next apply.

resource "frends_process" "order_processor" {
  package_path    = "${path.module}/packages/OrderProcessor.json"
  package_hash    = filesha256("${path.module}/packages/OrderProcessor.json")
  import_conflict = "NewVersion"
}

# Deploy the imported process to the production agent group.
data "frends_environment" "production" {
  display_name = "Production"
}

data "frends_agent_group" "production_agents" {
  display_name   = "Production Agents"
  environment_id = data.frends_environment.production.id
}

resource "frends_process_deployment" "order_processor" {
  agent_group_id    = data.frends_agent_group.production_agents.id
  activate_triggers = true

  processes {
    process_guid = frends_process.order_processor.guid
    version      = frends_process.order_processor.version
  }

  depends_on = [frends_process.order_processor]
}

output "process_guid" {
  description = "Stable GUID for the process (use in deployments)."
  value       = frends_process.order_processor.guid
}

output "process_version" {
  description = "Current version number after the last import."
  value       = frends_process.order_processor.version
}
