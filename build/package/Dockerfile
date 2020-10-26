FROM golang:1.15.3-alpine as builder

RUN apk add git make

RUN mkdir /stevebot && \
        git clone https://github.com/cezarmathe/stevebot.git /stevebot && \
        cd /stevebot && \
        make

FROM alpine:latest

RUN addgroup stevebot && \
    adduser -S -G stevebot stevebot

RUN apk add --no-cache \
        dumb-init \
        su-exec

COPY --from=builder /stevebot/bin/stevebot /bin/stevebot

ENTRYPOINT ["/usr/bin/dumb-init", "--"]
CMD [ "su-exec", "stevebot:stevebot", "stevebot" ]
