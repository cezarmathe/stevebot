# Dockerfile

FROM alpine:edge

# add the stevebot user and group
RUN addgroup stevebot && adduser -S -G stevebot stevebot
# update package db and install dumb-init (init process) and su-exec (run
# application as another user and group)
RUN apk update && apk add --no-cache dumb-init su-exec

ENTRYPOINT ["/usr/bin/dumb-init", "--"]
CMD [ "su-exec", "stevebot:stevebot", "stevebot" ]

# required environment variables

ENV STEVEBOT_RCON_HOST="127.0.0.1"
ENV STEVEBOT_RCON_PORT="25575"
ENV STEVEBOT_RCON_PASSWORD=""
ENV STEVEBOT_DISCORD_TOKEN=""
ENV STEVEBOT_COMMAND_PREFIX="~"

# ---

# the version of stevebot to be packaged
ARG version="dev"

# add stevebot binary
COPY cmd/stevebot/stevebot-${version}-linux-amd64 /usr/local/bin/stevebot
