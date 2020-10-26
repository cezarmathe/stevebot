# stevebot

job "stevebot" {
  region      = "${region}"
  datacenters = ["${datacenter}"]

  type = "service"

  group "stevebot" {
    count = 1

    task "stevebot" {
      driver = "docker"

      config {
        image = "cezarmathe/stevebot:${stevebot_docker_version}"
      }

      env {
        STEVEBOT_TOKEN          = "${stevebot_token}"
        STEVEBOT_COMMAND_PREFIX = "${stevebot_command_prefix}"
        STEVEBOT_RCON_HOST      = "${stevebot_rcon_host}"
        STEVEBOT_RCON_PORT      = "${stevebot_rcon_port}"
        STEVEBOT_RCON_PASS      = "${stevebot_rcon_pass}"
      }
    }
  }
}
