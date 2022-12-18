# syntax=docker/dockerfile:1

FROM alpine:latest

WORKDIR /
COPY Kard-bot /
COPY config /config
COPY assets /assets
COPY .env_example /.env
COPY Robo_cat.png /
COPY README.md /
COPY LICENSE /

ENTRYPOINT [ "/Kard-bot" ]