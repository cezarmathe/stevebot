# stevebot

provider "nomad" {
  address   = var.nomad_addr
  secret_id = var.nomad_token
}

resource "nomad_job" "stevebot" {
  jobspec = templatefile("${path.module}/jobspec.hcl", {
    stevebot_docker_version = var.stevebot_docker_version
    stevebot_token          = var.stevebot_token
    stevebot_command_prefix = var.stevebot_command_prefix
    stevebot_rcon_host      = var.stevebot_rcon_host
    stevebot_rcon_port      = var.stevebot_rcon_port
    stevebot_rcon_pass      = var.stevebot_rcon_password
    region                  = var.nomad_region
    datacenter              = var.nomad_datacenter
  })
}
