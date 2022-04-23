# syntax=docker/dockerfile:1

FROM golang:1.18-alpine
RUN apk add --no-cache tzdata

WORKDIR /
COPY Kard-bot /
COPY config /config
COPY assets /assets
COPY .env_example /.env
COPY Robo_cat.png /
COPY README.md /
COPY LICENSE /

ENTRYPOINT [ "/Kard-bot" ]