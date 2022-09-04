# syntax=docker/dockerfile:1

FROM alpine:latest

WORKDIR /

RUN apk add --update --no-cache \
  python3 py3-numpy g++ linux-headers \
  python3-dev musl-dev build-base \
  libc-dev mariadb-dev postgresql-dev \
  freetype-dev libpng-dev libxml2-dev \
  libxslt-dev zlib-dev gcc make py3-numpy-dev && \
  ln -sf python3 /usr/bin/python && \
  python3 -m ensurepip && \
  pip3 install --no-cache --upgrade pip setuptools wheel matplotlib && \
  pip3 install "docarray[common]>=0.13.5" jina

COPY Robo_cat.png /
COPY README.md /
COPY LICENSE /
COPY config /config
COPY assets /assets
COPY .env_example /.env
COPY Kard-bot /

ENTRYPOINT [ "/Kard-bot" ]