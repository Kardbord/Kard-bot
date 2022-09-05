# syntax=docker/dockerfile:1

FROM debian:stable

WORKDIR /

RUN apt-get update && apt-get install -y \
  libffi-dev \
  python3 \
  python3-dev \
  python3-numpy \
  python3-numpy-dev \
  python3-matplotlib \
  python3-pip && \
  apt-get autoremove -y && \
  ln -sf python3 /usr/bin/python

RUN pip3 install --upgrade pip setuptools wheel
RUN printf '[global]\nextra-index-url=https://www.piwheels.org/simple' >/etc/pip.conf
RUN pip3 install "docarray[common]>=0.13.5" jina

COPY Robo_cat.png /
COPY README.md /
COPY LICENSE /
COPY config /config
COPY assets /assets
COPY .env_example /.env
COPY Kard-bot /

ENTRYPOINT [ "/Kard-bot" ]