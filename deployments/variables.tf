# stevebot - variables

variable "nomad_addr" {
    type        = string
    description = "The address of Nomad."
}

variable "nomad_token" {
    type        = string
    description = "The address of Nomad."
}

variable "nomad_region" {
    type        = string
    description = "The deployment region."
    default     = "global"
}

variable "nomad_datacenter" {
    type        = string
    description = "The deployment datacenter."
}

variable "stevebot_docker_version" {
    type        = string
    description = "The Docker image to use."
    default     = "0.0.2"
}

variable "stevebot_token" {
    type        = string
    description = "The Discord token used by stevebot."
}

variable "stevebot_command_prefix" {
    type        = string
    description = "The prefix used by stevebot."
    default     = "~"
}

variable "stevebot_rcon_host" {
    type        = string
    description = "The host of the minecraft server."
    default     = "0.0.2"
}

variable "stevebot_rcon_port" {
    type        = number
    description = "The port used for rcon."
    default     = 25575
}

variable "stevebot_rcon_password" {
    type        = string
    description = "The rcon password."
}
