# Deploy a Docker container yourself

- Pull the latest image image from Docker Hub

`docker pull cezarmathe/stevebot:latest`

- Run the container (all environment variables have been presented with their
  default values)

```shell
docker run \
    --detach \
    --env "STEVEBOT_RCON_HOST=127.0.0.1" \
	--env "STEVEBOT_RCON_PORT=25575" \
	--env "STEVEBOT_RCON_PASSWORD=" \
	--env "STEVEBOT_DISCORD_TOKEN=" \
	--env "STEVEBOT_COMMAND_PREFIX=~" \
	--env "STEVEBOT_ALLOWED_COMMANDS=" \
	--env "STEVEBOT_FORBIDDEN_COMMANDS=" \
    --name "stevebot" \
    --restart "unless-stopped" \
    cezarmathe/stevebot:latest
```
