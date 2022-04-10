# syntax=docker/dockerfile:1

FROM golang:1.18-alpine

WORKDIR /
COPY Kard-bot /
COPY Robo_cat.png /
COPY README.md /
COPY LICENSE /

ENTRYPOINT [ "/Kard-bot" ]
